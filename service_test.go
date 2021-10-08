package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"testing"
)

func TestServiceRegister(t *testing.T) {
	err := ServiceRegister(
		&api.Config{
			Address: "127.0.0.1:8500",
		},
		&api.AgentServiceRegistration{
			ID:      "004",
			Name:    "my_service",
			Tags:    []string{"tag1", "tag2"},
			Port:    8080,
			Address: "10.100.201.1",
		},
		&api.AgentServiceCheck{
			Interval:                       "5s",
			Timeout:                        "5s",
			HTTP:                           "http://www.baidu.com",
			Method:                         "get",
			DeregisterCriticalServiceAfter: "15s",
		},
	)
	if err != nil {
		t.FailNow()
	}
}

func TestServiceDiscovery(t *testing.T) {
	serviceName := "my_service"
	consulAddr := "127.0.0.1:8500"
	svc, err := ServiceDiscovery(
		&api.Config{
			Address: consulAddr,
		}, serviceName, nil,
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(svc.Service.Address, svc.Service.Port, svc.Service.ID)
}

func TestServiceDeregister(t *testing.T) {
	serviceId := "003"
	consulAddr := "127.0.0.1:8500"
	err := ServiceDeregister(
		&api.Config{
			Address: consulAddr,
		}, serviceId,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceGetAll(t *testing.T) {
	consulAddr := "127.0.0.1:8500"
	svc, err := ServiceGetAll(&api.Config{
		Address: consulAddr,
	})
	if err != nil {
		t.Fatal(err)
	}
	for k, s := range svc {
		fmt.Println(k, s)
	}
}
