package consul

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/internal/constants"
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

func findTrafficIp() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return net.IP{}, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

func generateNodeId(appName string) string {
	data := fmt.Sprintf("mediapire-manager-%s", appName)
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	return hashStr[len(hashStr)-12:]
}

func RegisterService() error {
	trafficIp, err := findTrafficIp()
	if err != nil {
		return err
	}

	managerApp := app.GetApp()
	self := managerApp.Config

	nodeId := generateNodeId(self.Name)

	registration := &api.AgentServiceRegistration{
		ID:      nodeId,
		Name:    self.Name,
		Port:    self.Port,
		Address: trafficIp.String(),
		Tags:    []string{constants.ConsulServiceTag},
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

	managerApp.NodeId = nodeId

	return nil
}

func UnregisterService() error {
	managerApp := app.GetApp()

	return consulClient.Agent().ServiceDeregister(managerApp.NodeId)
}
