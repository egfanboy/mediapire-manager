package node

import (
	"context"

	mediahost "mediapire/manager/internal/integrations/media-host"

	"github.com/egfanboy/mediapire-common/exceptions"
	"github.com/rs/zerolog/log"
)

type nodeApi interface {
	RegisterNode(ctx context.Context, req RegisterNodeRequest) error
}

type nodeService struct {
}

func (s *nodeService) RegisterNode(ctx context.Context, req RegisterNodeRequest) (err error) {
	log.Trace().Msg("RegisterNode start")
	port := 443

	if req.Port != nil {
		port = *req.Port
	}

	err = mediahost.NewMediaHostIntegration().VerifyConnectivity(req.Scheme, req.Host.String(), port)

	if err != nil {
		log.Error().Err(err)
		err = exceptions.NewBadRequestException(err)
	}

	log.Trace().Msg("RegisterNode end")
	return
}

func newNodeService() nodeApi {
	return &nodeService{}
}
