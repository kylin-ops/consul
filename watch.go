package consul

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

type KeyValue struct {
	Key   string
	Value []byte
}

type WatchValue struct {
	KV      []*KeyValue
	Service []*api.ServiceEntry
}

type ChangeWatchValue struct {
	Add    *WatchValue
	Delete *WatchValue
	Change *WatchValue
}

type lastData struct {
	Hash    string
	Type    string
	KV      *KeyValue
	Service *api.ServiceEntry
}

type Watch struct {
	Address string
	wp      *watch.Plan
	ch      chan *WatchValue
}

// 监控的内容变化后，获取全量的数据
func (w *Watch) watcher(addr string, wp *watch.Plan) (chan *WatchValue, error) {
	var ch = make(chan *WatchValue)
	var err error
	w.ch = ch
	go func() {
		wp.Handler = func(i uint64, d interface{}) {
			switch ds := d.(type) {
			case api.KVPair:
				ch <- &WatchValue{KV: []*KeyValue{{Key: ds.Key, Value: ds.Value}}}
			case api.KVPairs:
				var tData []*KeyValue
				for _, data := range ds {
					tData = append(tData, &KeyValue{Key: data.Key, Value: data.Value})
				}
				ch <- &WatchValue{KV: tData}
			case []*api.ServiceEntry:
				ch <- &WatchValue{Service: ds}
			}
		}
		_ = wp.Run(addr)
	}()
	return ch, err
}

// 监控keyprefix
func (w *Watch) KeyPrefix(keyprefix string) (chan *WatchValue, error) {
	wp, err := watch.Parse(map[string]interface{}{
		"type":   "keyprefix",
		"prefix": keyprefix,
	})
	if err != nil {
		return nil, err
	}
	w.wp = wp
	return w.watcher(w.Address, wp)
}

// 监控key
func (w *Watch) Key(key string) (chan *WatchValue, error) {
	wp, err := watch.Parse(map[string]interface{}{
		"type": "key",
		"key":  key,
	})
	if err != nil {
		return nil, err
	}
	w.wp = wp
	return w.watcher(w.Address, wp)
}

// 监控service
func (w *Watch) Service(serviceName string, tags []string) (chan *WatchValue, error) {
	wp, err := watch.Parse(map[string]interface{}{
		"type":    "service",
		"service": serviceName,
		"tag":     tags,
	})
	if err != nil {
		return nil, err
	}
	w.wp = wp
	return w.watcher(w.Address, wp)
}

// 停止watch
func (w *Watch) Stop() {
	w.wp.Stop()
	close(w.ch)
}

func dataHash(data interface{}) string {
	d, _ := json.Marshal(data)
	m := md5.New()
	m.Write(d)
	return fmt.Sprintf("%x", m.Sum(nil))
}

// 只获取consul，增加，删除，修改的数据，没有变化的数据不换行
// 变化的数据在ChangeWatchValue结构器中，通过channel传输
type WatchChange struct {
	Address string
	wp      *watch.Plan
	ch      chan *ChangeWatchValue
	last    map[string]*lastData // 上一次数据
	new     map[string]*lastData // 最后一次的数据
}

// 对比上一次和最新一次的数据，返回增加，修改，删除的数据，并分类存储
func (w *WatchChange) diff() *ChangeWatchValue {
	var rData = &ChangeWatchValue{Add: &WatchValue{}, Delete: &WatchValue{}, Change: &WatchValue{}}
	// 获取删除的数据
	for key, data := range w.last {
		if _, ok := w.new[key]; !ok {
			switch data.Type {
			case "service":
				rData.Delete.Service = append(rData.Delete.Service, data.Service)
			case "kv":
				rData.Delete.KV = append(rData.Delete.KV, data.KV)
			}
		}
	}
	// 获取新增加和改变的数据
	for key, data := range w.new {
		// 获取新增加的数据
		if d, ok := w.last[key]; !ok {
			switch data.Type {
			case "service":
				rData.Add.Service = append(rData.Add.Service, data.Service)
			case "kv":
				rData.Add.KV = append(rData.Add.KV, data.KV)
			}
		} else {
			// 获取变化过的数据
			fmt.Println("hash", data.Hash, d.Hash)
			if data.Hash != d.Hash {
				switch data.Type {
				case "service":
					rData.Change.Service = append(rData.Change.Service, data.Service)
				case "kv":
					rData.Change.KV = append(rData.Change.KV, data.KV)
				}
			}
		}
	}
	return rData
}

// 监控的内容变化后，获取变化的数据
func (w *WatchChange) changeWatcher(addr string, wp *watch.Plan) (chan *ChangeWatchValue, error) {
	var ch = make(chan *ChangeWatchValue)
	var err error
	w.ch = ch
	go func() {
		wp.Handler = func(i uint64, d interface{}) {
			w.new = map[string]*lastData{}
			switch ds := d.(type) {
			case api.KVPairs:
				for _, data := range ds {
					tData := &KeyValue{Key: data.Key, Value: data.Value}
					h := dataHash(tData)
					w.new[data.Key] = &lastData{Type: "kv", KV: tData, Hash: h}
				}
			case []*api.ServiceEntry:
				for _, svc := range ds {
					h := dataHash(svc)
					w.new[svc.Service.ID] = &lastData{Type: "service", Service: svc, Hash: h}
				}
			}
			if chData := w.diff(); chData != nil {
				ch <- chData
			}
			w.last = w.new
		}
		_ = wp.Run(addr)
	}()
	return ch, err
}

// 监控service，只返回变化的数据
func (w *WatchChange) Service(serviceName string, tags []string) (chan *ChangeWatchValue, error) {
	wp, err := watch.Parse(map[string]interface{}{
		"type":    "service",
		"service": serviceName,
		"tag":     tags,
	})
	if err != nil {
		return nil, err
	}
	w.wp = wp
	return w.changeWatcher(w.Address, wp)
}

// 监控KeyPrefix，只返回变化的数据
func (w *WatchChange) KeyPrefix(keyprefix string) (chan *ChangeWatchValue, error) {
	wp, err := watch.Parse(map[string]interface{}{
		"type":   "keyprefix",
		"prefix": keyprefix,
	})
	if err != nil {
		return nil, err
	}
	w.wp = wp
	return w.changeWatcher(w.Address, wp)
}

// 停止watch
func (w *WatchChange) Stop() {
	w.wp.Stop()
	close(w.ch)
}

func NewWatchChange(addr string) *WatchChange {
	return &WatchChange{
		last:    map[string]*lastData{},
		new:     map[string]*lastData{},
		Address: addr,
	}
}
