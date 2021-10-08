package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"math/rand"
	"time"
)

// 服务端注册服务到consul
func ServiceRegister(connectConfig *api.Config, registerConfig *api.AgentServiceRegistration,
	checkConfig *api.AgentServiceCheck) error {
	client, err := api.NewClient(connectConfig)
	if err != nil {
		return err
	}
	registerConfig.Check = checkConfig
	return client.Agent().ServiceRegister(registerConfig)
}

// 客户端从consul取服务,随机返回一个服务
func ServiceDiscovery(connectConfig *api.Config, serviceName string, tags []string) (*api.ServiceEntry, error) {
	var lastIndex uint64
	client, err := api.NewClient(connectConfig)
	if err != nil {
		return nil, err
	}
	services, _, err := client.Health().ServiceMultipleTags(serviceName, tags, true,
		&api.QueryOptions{WaitIndex: lastIndex})
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("no nodes are available for the specified service \"%s\"", serviceName)
	}
	rand.Seed(time.Now().Unix())
	svc := services[rand.Intn(len(services))]
	return svc, nil
}

// 取消移除注册的服务
func ServiceDeregister(connectConfig *api.Config, serviceId string) error {
	client, err := api.NewClient(connectConfig)
	if err != nil {
		return err
	}
	return client.Agent().ServiceDeregister(serviceId)
}

// 获取所有的服务
func ServiceGetAll(connectConfig *api.Config) (map[string]*api.AgentService, error) {
	client, err := api.NewClient(connectConfig)
	if err != nil {
		return nil, err
	}
	return client.Agent().Services()
}
