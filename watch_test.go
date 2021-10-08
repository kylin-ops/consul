package consul

import (
	"fmt"
	"testing"
	"time"
)

func TestWatch_WatchKey(t *testing.T) {
	w := Watch{Address: "127.0.0.1:8500"}
	ch, err := w.Key("my_kv")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for d := range ch {
			fmt.Println(d)
		}
		fmt.Println("end ch")
	}()
	time.Sleep(time.Second * 10)
	w.Stop()
	time.Sleep(time.Second * 5)
}

func TestWatch_WatchKeyPrefix(t *testing.T) {
	w := Watch{Address: "127.0.0.1:8500"}
	ch, err := w.KeyPrefix("my_kv")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for d := range ch {
			fmt.Println(d)
		}
		fmt.Println("end ch")
	}()
	time.Sleep(time.Second * 10)
	w.Stop()
	time.Sleep(time.Second * 5)
}

func TestWatch_WatchService(t *testing.T) {
	w := Watch{Address: "127.0.0.1:8500"}
	ch, err := w.Service("my_service", nil)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for d := range ch {
			fmt.Println(d)
		}
		fmt.Println("end ch")
	}()
	time.Sleep(time.Second * 10)
	w.Stop()
	time.Sleep(time.Second * 5)
}
