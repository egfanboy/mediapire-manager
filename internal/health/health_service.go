package node

import (
	"context"
)

type healthApi interface {
	GetHealth(ctx context.Context) error
}

type healthService struct {
}

func (s *healthService) GetHealth(ctx context.Context) (err error) {
	// For now treat reaching here as a success
	return nil
}

func newNodeService() healthApi {
	return &healthService{}
}
