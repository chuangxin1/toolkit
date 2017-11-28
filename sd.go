package toolkit

import (
	"fmt"

	"github.com/go-kit/kit/log"
	consulsd "github.com/go-kit/kit/sd/consul"
	consulapi "github.com/hashicorp/consul/api"
)

// ServiceRegister register new server
func ServiceRegister(
	consulAddr, id, name, addr string,
	port int,
	tags []string) error {
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	client, err := consulapi.NewClient(config)
	if err != nil {
		return err
	}
	//创建一个新服务。
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = id
	registration.Name = name
	registration.Port = port
	registration.Tags = tags
	registration.Address = addr

	//增加check。
	check := new(consulapi.AgentServiceCheck)
	check.HTTP = fmt.Sprintf("http://%s:%d%s", addr, port, "/health")
	//设置超时 5s。
	check.Timeout = "5s"
	//设置间隔 5s。
	check.Interval = "30s"
	//注册check服务。
	registration.Check = check

	return client.Agent().ServiceRegister(registration)
}

// NewConsulInstancer consul Instancer
func NewConsulInstancer(
	consulAddr, name string,
	tags []string,
	passingOnly bool,
	logger log.Logger) (instancer *consulsd.Instancer, err error) {
	var client consulsd.Client

	config := consulapi.DefaultConfig()
	if len(consulAddr) > 0 {
		config.Address = consulAddr
	}
	consulClient, err := consulapi.NewClient(config)
	if err != nil {
		logger.Log("err", err)
		return
	}
	client = consulsd.NewClient(consulClient)
	instancer = consulsd.NewInstancer(client, logger, name, tags, passingOnly)

	return
}
