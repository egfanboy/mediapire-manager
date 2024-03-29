package node

import (
	"strconv"

	"github.com/google/uuid"
)

type NodeConfig struct {
	NodeHost   string    `json:"host"`
	NodePort   string    `json:"port"`
	NodeScheme string    `json:"scheme"`
	IsUp       bool      `json:"isUp"`
	Id         uuid.UUID `json:"id"`
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
