package node

import (
	"strconv"
)

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
