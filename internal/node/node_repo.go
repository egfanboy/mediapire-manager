package node

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/egfanboy/mediapire-common/exceptions"
	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/internal/constants"
	"github.com/egfanboy/mediapire-manager/internal/consul"
	"github.com/hashicorp/consul/api"
)

func nodeConfigFromConsul(source *api.AgentService, status string) (NodeConfig, error) {

	cfg := NodeConfig{
		NodeHost:   source.Address,
		NodePort:   strconv.Itoa(source.Port),
		NodeScheme: source.Meta[consul.KeyScheme],
		Id:         source.ID,
		Name:       source.Service,
	}

	if status == api.HealthCritical {
		cfg.IsUp = false
	} else {
		cfg.IsUp = true
	}

	return cfg, nil
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
	services, err := r.client.Agent().ServicesWithFilter(fmt.Sprintf("Tags contains \"%s\"", constants.ConsulMediaHostServiceTag))
	if err != nil {
		return
	}

	for _, service := range services {
		status, _, errHealth := r.client.Agent().AgentHealthServiceByID(service.ID)
		if errHealth != nil {
			err = errHealth
			return
		}

		cfg, err := nodeConfigFromConsul(service, status)
		if err != nil {
			return nil, err
		}
		result = append(result, cfg)
	}

	return
}

func (r *consulRepo) GetNode(ctx context.Context, nodeId string) (NodeConfig, error) {
	service, _, err := r.client.Agent().Service(nodeId, &api.QueryOptions{UseCache: false})
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

	status, _, err := r.client.Agent().AgentHealthServiceByID(service.ID)
	if err != nil {
		return NodeConfig{}, &exceptions.ApiException{
			Err: err, StatusCode: http.StatusInternalServerError,
		}
	}

	return nodeConfigFromConsul(service, status)
}

func NewNodeRepo() (NodeRepo, error) {

	consul, err := consul.GetClient()

	if err != nil {
		return nil, err
	}

	return &consulRepo{app: app.GetApp(), client: consul}, nil
}
