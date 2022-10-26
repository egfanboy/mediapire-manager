package consul

import (
	"fmt"
	"net"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/hashicorp/consul/api"
)

var consulClient *api.Client

const (
	KeyScheme = "scheme"
)

func GetClient() (*api.Client, error) {
	if consulClient == nil {
		err := NewConsulClient()

		if err != nil {
			return nil, err
		}

	}

	return consulClient, nil
}

func NewConsulClient() error {
	if consulClient != nil {
		return nil
	} else {
		defaultConfig := api.DefaultConfig()

		consulCfg := app.GetApp().Config.Consul

		defaultConfig.Address = fmt.Sprintf("%s:%d", consulCfg.Address, consulCfg.Port)
		defaultConfig.Scheme = consulCfg.Scheme

		client, err := api.NewClient(defaultConfig)

		if err != nil {
			return err
		}

		consulClient = client

		return nil
	}
}

func findTrafficIp() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return net.IP{}, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

func RegisterService() error {
	trafficIp, err := findTrafficIp()

	if err != nil {
		return err
	}

	self := app.GetApp().Config

	registration := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("mediapire-manager-%s", trafficIp),
		Name:    "mediapire-manager",
		Port:    self.Port,
		Address: trafficIp.String(),
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("%s://%s:%v/api/v1/health", self.Scheme, trafficIp, self.Port),
			Interval: "10s",
			Timeout:  "30s",
		},
		Meta: map[string]string{KeyScheme: self.Scheme},
	}

	return consulClient.Agent().ServiceRegister(registration)
}

func UnregisterService() error {

	trafficIp, err := findTrafficIp()

	if err != nil {
		return err
	}

	return consulClient.Agent().ServiceDeregister(fmt.Sprintf("mediapire-manager-%s", trafficIp))
}
