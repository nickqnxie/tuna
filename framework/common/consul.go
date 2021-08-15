/*
@Time : 2019/10/19 20:28
@Author : nickqnxie
@File : consul.go
*/

package common

import (
	"errors"
	"fmt"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/sirupsen/logrus"
)

var (
	consulDefautAddress = "127.0.0.1:8500"
)

type Option func(*Options)

type (
	Consul struct {
		Client  *consulApi.Client
		opts    Options
		address string
	}

	Options struct {
		Addrs string
		Token string
	}

	ServiceRegistration struct {
		Id                string
		Name              string
		Address           string
		Port              int
		Tags              []string
		Checkargs         *AgentServiceCheck
		AgentServiceCheck AgentServiceCheck
		Check             bool
	}

	AgentServiceCheck struct {
		HttpUrl                        string
		Timeout                        string
		Interval                       string
		DeregisterCriticalServiceAfter string
	}

	ServerInstance struct {
		ID      string
		Address string
		Port    int
	}
)

type ConsulWatchKey func(keyprefix string, kvPairsCh chan consulApi.KVPairs) error

func NewConsulClinet(opts ...Option) (*Consul, error) {
	c := new(Consul)
	var client *consulApi.Client
	var err error
	consulconfig := new(consulApi.Config)

	for _, o := range opts {
		o(&c.opts)
	}

	if c.opts.Addrs != "" {
		consulconfig.Address = c.opts.Addrs
	} else {
		consulconfig.Address = consulDefautAddress
	}
	//配置token
	if c.opts.Token != "" {
		consulconfig.Token = c.opts.Token
	}

	if client, err = consulApi.NewClient(consulconfig); err != nil {
		return nil, err
	}

	return &Consul{
		Client:  client,
		address: consulconfig.Address,
	}, nil

}

func (c *Consul) GetKV(keyPath string) (*consulApi.KVPair, error) {

	kvPair, _, err := c.Client.KV().Get(keyPath, nil)
	if err != nil {
		errorInfo := fmt.Sprintf("Get %s failed, error:%s", keyPath, err)
		logrus.Error(errorInfo)
		return nil, nil
	}

	return kvPair, nil

}

// PutKV 向consul中写入KV数据
func (c *Consul) PutKV(keyPath string, value []byte) error {

	var err error
	kv := c.Client.KV()
	putKV := &consulApi.KVPair{}
	putKV.Key = keyPath
	putKV.Value = value

	_, err = kv.Put(putKV, nil)
	if err != nil {
		errorInfo := fmt.Sprintf("kv.Put %s failed, error:%s", keyPath, err)
		logrus.Error(errorInfo)
		return errors.New(errorInfo)
	}

	return nil
}

// DeleteKV 向consul中写入KV数据
func (c *Consul) DeleteKV(keyPath string) error {
	var err error
	kv := c.Client.KV()
	_, err = kv.Delete(keyPath, nil)
	if err != nil {
		errorInfo := fmt.Sprintf("kv.Put %s failed, error:%s", keyPath, err)
		logrus.Error(errorInfo)
		return errors.New(errorInfo)
	}

	return nil
}

// DeleteTree 删除kv目录树
func (c *Consul) DeleteTree(keyPath string) error {
	var err error
	kv := c.Client.KV()
	_, err = kv.DeleteTree(keyPath, nil)
	if err != nil {
		errorInfo := fmt.Sprintf("kv.DeleteTree %s failed, error:%s", keyPath, err)
		logrus.Error(errorInfo)
		return errors.New(errorInfo)
	}

	return nil
}

// ListPreKey 获取kv前缀列表
func (c *Consul) ListPreKey(preKey string) (consulApi.KVPairs, error) {

	kvPairs, _, err := c.Client.KV().List(preKey, nil)
	if err != nil {
		errorInfo := fmt.Sprintf("list %s from  failed, error:%s", preKey, err)
		logrus.Error(errorInfo)
		return nil, nil
	}

	return kvPairs, nil
}

//注册服务
func (service *Consul) ServiceRegister(Serviceinfo *ServiceRegistration) (*ServiceRegistration, error) {

	//服务注册基本信息
	registration := new(consulApi.AgentServiceRegistration)
	registration.ID = Serviceinfo.Id
	registration.Name = Serviceinfo.Name

	registration.Address = Serviceinfo.Address
	registration.Port = Serviceinfo.Port
	registration.Tags = Serviceinfo.Tags

	//服务检查基本信息
	if Serviceinfo.Check {
		check := new(consulApi.AgentServiceCheck)
		check.HTTP = Serviceinfo.AgentServiceCheck.HttpUrl                                                  //check url
		check.Timeout = Serviceinfo.AgentServiceCheck.Timeout                                               //请求超时时间
		check.Interval = Serviceinfo.AgentServiceCheck.Interval                                             //间隔检查时间
		check.DeregisterCriticalServiceAfter = Serviceinfo.AgentServiceCheck.DeregisterCriticalServiceAfter //不健康节点删除时间
		registration.Check = check
	}
	//注册服务
	if err := service.Client.Agent().ServiceRegister(registration); err != nil {
		return Serviceinfo, err
	}
	return Serviceinfo, nil
}

//删除服务
func (service *Consul) ServiceDeregister(Serviceinfo *ServiceRegistration) (err error) {

	if err = service.Client.Agent().ServiceDeregister(Serviceinfo.Id); err != nil {
		return
	}
	return nil
}

func (service *Consul) GetServerInstance(serviceName string) ([]*ServerInstance, error) {
	var servers []*ServerInstance
	agentServices, err := service.Client.Agent().Services()

	if err != nil {
		return servers, nil
	}
	for _, items := range agentServices {
		if items.Service == serviceName {
			servers = append(servers, &ServerInstance{
				ID:      items.ID,
				Address: items.Address,
				Port:    items.Port,
			})
		}
	}
	return servers, nil

}

func (c *Consul) WatchKeyPrefix(keyprefix string, kvPairsCh chan consulApi.KVPairs) error {

	params := make(map[string]interface{})

	params["type"] = "keyprefix"
	params["prefix"] = keyprefix

	plan, err := watch.Parse(params)
	if err != nil {
		errorInfo := fmt.Sprintf("watchConfigGroups params failed, error:%s", err)
		logrus.Errorf("%v", errorInfo)
		return errors.New(errorInfo)
	}

	plan.Handler = func(index uint64, result interface{}) {
		if kvPairs, ok := result.(consulApi.KVPairs); ok {
			kvPairsCh <- kvPairs
		}
	}
	//创建consul客户端
	if err != nil {
		errorInfo := fmt.Sprintf("create CreateConsulClient  failed, error:%s", err)
		logrus.Errorf("%v", errorInfo)
		return errors.New(errorInfo)
	}

	return plan.Run(c.address)
}
