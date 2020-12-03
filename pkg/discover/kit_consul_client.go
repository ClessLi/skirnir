package discover

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"strconv"
	"sync"
)

type KitConsulClient struct {
	Host   string // Consul Host
	Port   uint16 // Consul Port
	client consul.Client
	// 连接 consul 的配置
	config *api.Config
	mutex  sync.Mutex
	// 服务实例缓存字段
	instancesMap sync.Map
}

func NewKitConsulRegistryClient(consulHost string, consulPort uint16) (RegistryClient, error) {
	// 通过 Consul Host 和 Consul Port 创建一个 consul.Client
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulHost + ":" + strconv.Itoa(int(consulPort))
	apiClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}
	client := consul.NewClient(apiClient)
	return &KitConsulClient{
		Host:   consulHost,
		Port:   consulPort,
		config: consulConfig,
		client: client,
	}, err
}

func NewKitConsulDiscoveryClient(consulHost string, consulPort uint16) (DiscoveryClient, error) {
	// 通过 Consul Host 和 Consul Port 创建一个 consul.Client
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulHost + ":" + strconv.Itoa(int(consulPort))
	apiClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}
	client := consul.NewClient(apiClient)
	return &KitConsulClient{
		Host:   consulHost,
		Port:   consulPort,
		config: consulConfig,
		client: client,
	}, err
}

func (consulClient *KitConsulClient) Register(serviceName, instanceId, instanceHost string, instancePort uint16, meta map[string]string, logger log.Logger) bool {
	serviceRegistration := &api.AgentServiceRegistration{
		ID:      instanceId,
		Name:    serviceName,
		Port:    int(instancePort),
		Address: instanceHost,
		Meta:    meta,
		Check: &api.AgentServiceCheck{
			Interval:                       "15s",
			GRPC:                           fmt.Sprintf("%v:%v/%v", instanceHost, instancePort, serviceName),
			DeregisterCriticalServiceAfter: "30s",
		},
	}
	err := consulClient.client.Register(serviceRegistration)

	if err != nil {
		logger.Log(fmt.Sprintf("Register Service Error! Cased by: %s\n", err))
		return false
	}
	logger.Log("Register Service Success!")
	return true
}

func (consulClient *KitConsulClient) DeRegister(instanceId string, logger log.Logger) bool {

	// 构建包含服务实例 ID 的元数据结构体
	serviceRegistration := &api.AgentServiceRegistration{
		ID: instanceId,
	}
	// 发送服务注销请求
	err := consulClient.client.Deregister(serviceRegistration)

	if err != nil {
		logger.Log(fmt.Sprintf("Deregister Service Error! Cased by: %s\n", err))
		return false
	}
	logger.Log("Deregister Service Success!")

	return true
}

func (consulClient *KitConsulClient) DiscoverServices(serviceName string, logger log.Logger) sd.Instancer {

	return consul.NewInstancer(consulClient.client, logger, serviceName, nil, true)
	////  该服务已监控并缓存
	//instanceList, ok := consulClient.instancesMap.Load(serviceName)
	//if ok {
	//	return instanceList.([]interface{})
	//}
	//// 申请锁
	//consulClient.mutex.Lock()
	//// 再次检查是否监控
	//instanceList, ok = consulClient.instancesMap.Load(serviceName)
	//if ok {
	//	return instanceList.([]interface{})
	//} else {
	//	// 注册监控
	//	go func() {
	//		// 使用 consul 服务实例监控来监控某个服务名的服务实例列表变化
	//		params := make(map[string]interface{})
	//		params["type"] = "service"
	//		params["service"] = serviceName
	//		plan, _ := watch.Parse(params)
	//		plan.Handler = func(u uint64, i interface{}) {
	//			if i == nil {
	//				return
	//			}
	//			v, ok := i.([]*api.ServiceEntry)
	//			if !ok {
	//				return // 数据异常，忽略
	//			}
	//			// 没有服务实例在线
	//			if len(v) == 0 {
	//				consulClient.instancesMap.Store(serviceName, []interface{}{})
	//			}
	//			var healthServices []interface{}
	//			for _, service := range v {
	//				if service.Checks.AggregatedStatus() == api.HealthPassing {
	//					healthServices = append(healthServices, service.Service)
	//				}
	//			}
	//			consulClient.instancesMap.Store(serviceName, healthServices)
	//		}
	//		defer plan.Stop()
	//		plan.Run(consulClient.config.Address)
	//	}()
	//}
	//defer consulClient.mutex.Unlock()
	//
	//// 根据服务名请求服务实例列表
	//entries, _, err := consulClient.client.Service(serviceName, "", false, nil)
	//if err != nil {
	//	consulClient.instancesMap.Store(serviceName, []interface{}{})
	//	logger.Log(fmt.Sprintf("Discover Service Error! Cased by: %s\n", err))
	//	return nil
	//}
	//instances := make([]interface{}, len(entries))
	//for i := 0; i < len(instances); i++ {
	//	instances[i] = entries[i].Service
	//}
	//consulClient.instancesMap.Store(serviceName, instances)
	//return instances
}
