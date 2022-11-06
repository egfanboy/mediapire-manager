package node

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/egfanboy/mediapire-common/exceptions"
	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/internal/consul"
	"github.com/hashicorp/consul/api"
)

func nodeConfigFromConsul(source *api.AgentService) NodeConfig {
	return NodeConfig{NodeHost: source.Address,
		NodePort:   strconv.Itoa(source.Port),
		NodeScheme: source.Meta[consul.KeyScheme]}

}

type NodeRepo interface {
	GetAllNodes(ctx context.Context) ([]NodeConfig, error)
	GetNode(ctx context.Context, nodeId string) (NodeConfig, error)
}

type consulRepo struct {
	app *app.App

	client *api.Client
}

func (r *consulRepo) GetAllNodes(ctx context.Context) (result []NodeConfig, err error) {
	result = make([]NodeConfig, 0)
	services, err := r.client.Agent().ServicesWithFilter("Service == \"media-host-node\"")

	if err != nil {
		return
	}

	for _, service := range services {
		result = append(result, nodeConfigFromConsul(service))
	}

	return
}

func (r *consulRepo) GetNode(ctx context.Context, nodeId string) (NodeConfig, error) {

	service, _, err := r.client.Agent().Service("media-host-node-"+nodeId, &api.QueryOptions{UseCache: false})

	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return NodeConfig{}, &exceptions.ApiException{
				Err: err, StatusCode: http.StatusNotFound,
			}
		}

		return NodeConfig{}, &exceptions.ApiException{
			Err: err, StatusCode: http.StatusInternalServerError,
		}
	}

	return nodeConfigFromConsul(service), nil
}

func NewNodeRepo() (NodeRepo, error) {

	consul, err := consul.GetClient()

	if err != nil {
		return nil, err
	}

	return &consulRepo{app: app.GetApp(), client: consul}, nil
}
