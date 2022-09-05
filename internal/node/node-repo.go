package node

import (
	"context"
	"fmt"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/internal/constants"
)

type NodeRepo interface {
	GetAllNodes(ctx context.Context) ([]NodeConfig, error)
	GetNode(ctx context.Context, nodeId string) (NodeConfig, error)
}

type redisRepo struct {
	app *app.App
}

func (r *redisRepo) GetAllNodes(ctx context.Context) (result []NodeConfig, err error) {
	// Get number of hosts saved
	cmd := r.app.Redis.LLen(ctx, constants.KeyListHosts)
	if cmd.Err() != nil {
		err = cmd.Err()

		return
	}

	rangeCmd := r.app.Redis.LRange(ctx, constants.KeyListHosts, 0, cmd.Val()-1)
	if rangeCmd.Err() != nil {
		err = rangeCmd.Err()

		return
	}

	for _, h := range rangeCmd.Val() {
		nodeModel, err := r.GetNode(ctx, h)
		if err != nil {
			// handle error
		} else {
			result = append(result, nodeModel)
		}

	}

	return
}

func (r *redisRepo) GetNode(ctx context.Context, nodeId string) (NodeConfig, error) {
	hgetCmd := r.app.Redis.HGetAll(ctx, MakeNodeHash(nodeId))

	//  partial failures
	if hgetCmd.Err() != nil {
		// handle error
		return NodeConfig{}, hgetCmd.Err()
	}

	var nodeModel NodeConfig
	err := hgetCmd.Scan(&nodeModel)

	if err != nil {
		return NodeConfig{}, err
	}

	if nodeModel.NodeHost == "" || nodeModel.NodePort == "" || nodeModel.NodeScheme == "" {
		return nodeModel, fmt.Errorf("no node found with id %s", nodeId)
	}

	return nodeModel, nil
}

func NewNodeRepo() NodeRepo {
	return &redisRepo{app: app.GetApp()}
}
