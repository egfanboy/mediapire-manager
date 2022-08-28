package node

import (
	"net"
	"strconv"
)

// todo: validate request
type RegisterNodeRequest struct {
	Host   net.IP `json:"host"`
	Scheme string `json:"scheme"`
	Port   *int   `json:"port"`
}

type NodeConfig struct {
	NodeHost   string `redis:"host"`
	NodePort   string `redis:"port"`
	NodeScheme string `redis:"scheme"`
}

func (c NodeConfig) Host() string {
	return c.NodeHost
}

func (c NodeConfig) Scheme() string {
	return c.NodeScheme
}

func (c NodeConfig) Port() int {
	p, err := strconv.Atoi(c.NodePort)

	if err != nil {
		panic("cannot convert the node port to an integer")
	}

	return p
}
