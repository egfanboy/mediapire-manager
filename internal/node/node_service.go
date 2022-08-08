package node

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"mediapire/manager/internal/app"
	mediahost "mediapire/manager/internal/integrations/media-host"

	"github.com/egfanboy/mediapire-common/exceptions"
	"github.com/go-redis/redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	keyHostList = "hosts"
)

func MakeNodeHash(h string) string {
	return fmt.Sprintf("mediahost:%s", h)
}

type nodeApi interface {
	RegisterNode(ctx context.Context, req RegisterNodeRequest) error
}

type nodeService struct {
	app *app.App
}

func (s *nodeService) RegisterNode(ctx context.Context, req RegisterNodeRequest) (err error) {
	log.Trace().Msg("RegisterNode start")
	port := 443

	if req.Port != nil {
		port = *req.Port
	}

	q := s.app.Redis.LPos(ctx, keyHostList, req.Host.String(), redis.LPosArgs{})

	// record was found, throw error
	if q.Err() == nil {
		return &exceptions.ApiException{Err: errors.New("host already registered"), StatusCode: http.StatusConflict}
	} else if q.Err() != redis.Nil {
		// real error, return

		return q.Err()
	}

	err = mediahost.NewMediaHostIntegration().VerifyConnectivity(req.Scheme, req.Host.String(), port)

	if err != nil {
		log.Error().Err(err)
		err = exceptions.NewBadRequestException(err)
	}

	// add host to the list of hosts
	q = s.app.Redis.LPush(ctx, keyHostList, req.Host.String())

	if q.Err() != nil {
		log.Error().Err(q.Err()).Msg("failed to add host to redis")

		return q.Err()
	}

	// Save host info as a hash
	q = s.app.Redis.HSet(ctx, MakeNodeHash(req.Host.String()), map[string]interface{}{
		"host":   req.Host.String(),
		"port":   fmt.Sprintf("%d", port),
		"scheme": req.Scheme,
	})

	if q.Err() != nil {
		log.Error().Err(q.Err()).Msg("failed to save host info to redis")
		return q.Err()
	}

	log.Trace().Msg("RegisterNode end")
	return
}

func newNodeService() nodeApi {
	return &nodeService{app: app.GetApp()}
}
