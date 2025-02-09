package settings

import (
	"context"
	"fmt"
	"net/http"

	"github.com/egfanboy/mediapire-common/exceptions"
	"github.com/egfanboy/mediapire-manager/internal/node"
	"github.com/egfanboy/mediapire-manager/pkg/types"
	mhApi "github.com/egfanboy/mediapire-media-host/pkg/api"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type settingsApi interface {
	GetSettings(ctx context.Context) (types.MediaSettings, error)
	GetNodeSettings(ctx context.Context, nodeId uuid.UUID) (interface{}, error)
}

type settingsService struct {
	nodeRepo node.NodeRepo
}

func (s *settingsService) GetSettings(ctx context.Context) (result types.MediaSettings, err error) {
	log.Info().Msg("Getting Mediapire settings")
	// create a map where each key will be a filetype to ensure they are all unique across nodes
	fileTypeMapping := map[string]string{}

	nodes, err := s.nodeRepo.GetAllNodes(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all nodes")
		return
	}

	for _, node := range nodes {
		if !node.IsUp {
			log.Warn().Msgf("Not fetching settings from node %s since it is not up.", node.NodeHost)
			continue
		}

		// TODO: restrict time it takes to get a response
		nodeSettings, _, errSettings := mhApi.NewClient(ctx).GetSettings(node)
		if errSettings != nil {
			log.Error().Err(errSettings).Msgf("Failed to get settings from node %s", node.NodeHost)

			// do not fail the request if one node is unreachable.
			continue
		}

		for _, v := range nodeSettings.FileTypes {
			fileTypeMapping[v] = ""
		}

		for k := range fileTypeMapping {
			result.FileTypes = append(result.FileTypes, k)
		}
	}

	return
}

func (s *settingsService) GetNodeSettings(ctx context.Context, nodeId uuid.UUID) (result interface{}, err error) {
	log.Info().Msgf("Getting settings for media host %q", nodeId)
	node, err := s.nodeRepo.GetNode(ctx, nodeId)
	if err != nil {
		return
	}

	if !node.IsUp {
		err = &exceptions.ApiException{
			Err: fmt.Errorf("node %q is not up", nodeId), StatusCode: http.StatusBadRequest,
		}
		return
	}

	result, _, err = mhApi.NewClient(ctx).GetSettings(node)
	return
}

func newSettingsService() (settingsApi, error) {
	nodeRepo, err := node.NewNodeRepo()
	if err != nil {
		return nil, err
	}

	return &settingsService{nodeRepo: nodeRepo}, nil
}
