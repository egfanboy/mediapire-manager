package node

import (
	"context"
	"fmt"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/rs/zerolog/log"
)

const (
	keyHostList = "hosts"
)

func MakeNodeHash(h string) string {
	return fmt.Sprintf("mediahost:%s", h)
}

type nodeApi interface {
	GetAllNodes(ctx context.Context) ([]NodeConfig, error)
}

type nodeService struct {
	app  *app.App
	repo NodeRepo
}

func (s *nodeService) GetAllNodes(ctx context.Context) ([]NodeConfig, error) {
	log.Info().Msg("Getting all registered nodes")

	nodes, err := s.repo.GetAllNodes(ctx)

	if err != nil {
		log.Error().Err(err).Msg("Failed to get all mediahost nodes")
		return nil, err
	}

	return nodes, nil
}

func newNodeService() (nodeApi, error) {

	repo, err := NewNodeRepo()

	if err != nil {
		return nil, err
	}

	return &nodeService{app: app.GetApp(), repo: repo}, nil
}
