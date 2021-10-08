package main

import (
	"fmt"
	"github/kylin-ops/consul"
	"time"
)

func main() {
	w := consul.NewWatchChange("127.0.0.1:8500")
	ch, err := w.Service("my_service", nil)
	if err != nil {
		panic(err)
	}
	go func() {
		for d := range ch {
			fmt.Println("add", d.Add.Service)
			fmt.Println("change", d.Change.Service)
			fmt.Println("del", d.Delete.Service)
		}
		fmt.Println("end ch")
	}()
	time.Sleep(time.Second * 3000)
	fmt.Println("close ctx")
	w.Stop()
	fmt.Println("cccccc")
	time.Sleep(time.Second * 300)
}
