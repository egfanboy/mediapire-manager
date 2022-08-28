package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/egfanboy/mediapire-manager/pkg/types"
)

type MediaManagerApi interface {
	RegisterNode(r types.RegisterNodeRequest) (*http.Response, error)
}

type managerClient struct{}

func (c *managerClient) RegisterNode(registerRequest types.RegisterNodeRequest) (r *http.Response, err error) {
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(registerRequest)
	if err != nil {
		return
	}

	port := 443

	if registerRequest.Port != nil {
		port = *registerRequest.Port
	}

	r, err = http.Post(
		fmt.Sprintf(
			"%s://%s:%d/api/v1/nodes/register",
			registerRequest.Scheme,
			registerRequest.Host.String(),
			port,
		),
		"application/json",
		&buf)

	return
}

func NewManagerClient(ctx context.Context) MediaManagerApi {
	return &managerClient{}
}
