package toolkit

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	consulsd "github.com/go-kit/kit/sd/consul"
	consulapi "github.com/hashicorp/consul/api"
)

// ServiceOptions server options
type ServiceOptions struct {
	// Address is the address of the Consul server
	Address string

	// Scheme is the URI scheme for the Consul server
	Scheme string

	// Datacenter to use. If not provided, the default agent datacenter is used.
	Datacenter string

	// Transport is the Transport to use for the http client.
	//Transport *http.Transport

	// HttpClient is the client to use. Default will be
	// used if not provided.
	//HttpClient *http.Client

	// HttpAuth is the auth info to use for http access.
	//HttpAuth *HttpBasicAuth

	// WaitTime limits how long a Watch will block. If not provided,
	// the agent default values will be used.
	WaitTime time.Duration

	// Token is used to provide a per-request ACL token
	// which overrides the agent's default token.
	Token string

	//TLSConfig TLSConfig
}

// AgentServiceOptions agent options
type AgentServiceOptions struct {
	ID                string   `json:",omitempty"`
	Name              string   `json:",omitempty"`
	Tags              []string `json:",omitempty"`
	Port              int      `json:",omitempty"`
	Address           string   `json:",omitempty"`
	EnableTagOverride bool     `json:",omitempty"`
}

// ServiceRegister register new server
func ServiceRegister(options ServiceOptions, agent AgentServiceOptions) error {
	config := consulapi.DefaultConfig()
	config.Address = options.Address
	config.Scheme = options.Scheme
	config.Datacenter = options.Datacenter
	config.WaitTime = options.WaitTime
	config.Token = options.Token

	client, err := consulapi.NewClient(config)
	if err != nil {
		return err
	}
	//创建一个新服务。
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = agent.ID
	registration.Name = agent.Name
	registration.Port = agent.Port
	registration.Tags = agent.Tags
	registration.Address = agent.Address

	//增加check。
	check := new(consulapi.AgentServiceCheck)
	check.HTTP = fmt.Sprintf("http://%s:%d%s", agent.Address, agent.Port, "/health")
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
	options ServiceOptions,
	name string,
	tags []string,
	passingOnly bool,
	logger log.Logger) (instancer *consulsd.Instancer, err error) {
	var client consulsd.Client

	config := consulapi.DefaultConfig()
	config.Address = options.Address
	config.Scheme = options.Scheme
	config.Datacenter = options.Datacenter
	config.WaitTime = options.WaitTime
	config.Token = options.Token

	consulClient, err := consulapi.NewClient(config)
	if err != nil {
		logger.Log("err", err)
		return
	}
	client = consulsd.NewClient(consulClient)
	instancer = consulsd.NewInstancer(client, logger, name, tags, passingOnly)

	return
}
