package consul

import (
	"fmt"
	"net"
	"strconv"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

var consulClient *api.Client

const (
	KeyScheme = "scheme"
	KeyHost   = "host"
	KeyPort   = "port"
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

// Finds any service that is registered for this address and port and returns it
func findServiceForHost(host string, port int) (*api.AgentService, error) {
	services, err := consulClient.Agent().ServicesWithFilter("Service == \"mediapire-manager\"")
	if err != nil {
		return nil, err
	}

	servicesForHost := make([]*api.AgentService, 0)

	for i := range services {
		service := services[i]

		h, ok := service.Meta["host"]
		if !ok {
			continue
		}

		p, ok := service.Meta["port"]
		if !ok {
			continue
		}

		metaPort, err := strconv.Atoi(p)
		if err != nil {
			continue
		}

		if h == host && metaPort == port {
			servicesForHost = append(servicesForHost, service)
		}
	}

	if len(servicesForHost) == 0 {
		return nil, nil
	}

	if len(servicesForHost) > 1 {
		return nil, fmt.Errorf("found multiple services for host %s and port %d", host, port)
	}

	return servicesForHost[0], nil
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

	app := app.GetApp()
	self := app.Config

	service, err := findServiceForHost(trafficIp.String(), self.Port)
	if err != nil {
		return err
	}

	var nodeId uuid.UUID

	if service == nil {
		nodeId, err = uuid.NewUUID()
		if err != nil {
			return err
		}
	} else {
		nodeId = uuid.MustParse(service.ID)
	}

	registration := &api.AgentServiceRegistration{
		ID:      nodeId.String(),
		Name:    "mediapire-manager",
		Port:    self.Port,
		Address: trafficIp.String(),
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("%s://%s:%v/api/v1/health", self.Scheme, trafficIp, self.Port),
			Interval: "10s",
			Timeout:  "30s",
		},
		Meta: map[string]string{
			KeyScheme: self.Scheme,
			KeyHost:   trafficIp.String(),
			KeyPort:   strconv.Itoa(self.Port),
		},
	}

	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		return err
	}

	app.NodeId = nodeId

	return nil
}

func UnregisterService() error {

	trafficIp, err := findTrafficIp()

	if err != nil {
		return err
	}

	return consulClient.Agent().ServiceDeregister(fmt.Sprintf("mediapire-manager-%s", trafficIp))
}
