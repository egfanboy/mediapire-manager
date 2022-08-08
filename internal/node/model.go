package node

import "net"

// todo: validate request
type RegisterNodeRequest struct {
	Host   net.IP `json:"host"`
	Scheme string `json:"scheme"`
	Port   *int   `json:"port"`
}
