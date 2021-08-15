/*
@Time : 2019/8/29 07:50
@Author : nickqnxie
@File : consul.go
*/

package baseclass

import (
	"github.com/google/uuid"
	consulapi "github.com/hashicorp/consul/api"
)

var (
	client *consulapi.Client
)

func NewId() string {
	return uuid.New().String()
}

type ServiceRegistration struct {
	Id        string
	Name      string
	Address   string
	Port      int
	Tags      []string
	Checkargs *AgentServiceCheck
	Check     bool
}

type AgentServiceCheck struct {
	HttpUrl                        string
	Timeout                        string
	Interval                       string
	DeregisterCriticalServiceAfter string
}

func newConsulClient() (err error) {
	config := consulapi.DefaultConfig()
	//创建consul客户端
	if client, err = consulapi.NewClient(config); err != nil {
		return
	}
	return

}

func NewAgentServiceCheck(Timeout, Interval,
	DeregisterCriticalServiceAfter, HttpUrl string) *AgentServiceCheck {
	return &AgentServiceCheck{
		HttpUrl:                        HttpUrl,
		Timeout:                        Timeout,
		Interval:                       Interval,
		DeregisterCriticalServiceAfter: DeregisterCriticalServiceAfter,
	}
}

func NewConsulService(id, name, address string, port int,
	tags []string, check bool, Checkargs *AgentServiceCheck) *ServiceRegistration {

	s := ServiceRegistration{
		Id:      id,
		Name:    name,
		Address: address,
		Port:    port,
		Tags:    tags,
	}
	if check && Checkargs != nil {
		s.Checkargs = Checkargs
		s.Check = check
	}
	return &s
}

//注册服务
func (service *ServiceRegistration) ServiceRegister() (err error) {

	if err = newConsulClient(); err != nil {
		return
	}
	//服务注册基本信息
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = service.Id
	registration.Name = service.Name
	registration.Address = service.Address
	registration.Port = service.Port
	registration.Tags = service.Tags

	//服务检查基本信息
	if service.Check {
		check := new(consulapi.AgentServiceCheck)
		check.HTTP = service.Checkargs.HttpUrl                                                  //check url
		check.Timeout = service.Checkargs.Timeout                                               //请求超时时间
		check.Interval = service.Checkargs.Interval                                             //间隔检查时间
		check.DeregisterCriticalServiceAfter = service.Checkargs.DeregisterCriticalServiceAfter //不健康节点删除时间
		registration.Check = check
	}
	//注册服务
	if err = client.Agent().ServiceRegister(registration); err != nil {
		return
	}
	return nil
}

//删除服务
func (service *ServiceRegistration) ServiceDeregister() (err error) {
	if err = newConsulClient(); err != nil {
		return
	}
	if err = client.Agent().ServiceDeregister(service.Id); err != nil {
		return
	}
	return nil
}
