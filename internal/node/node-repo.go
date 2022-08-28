package node

import (
	"context"
	"mediapire/manager/internal/app"
	"mediapire/manager/internal/constants"
)

type NodeRepo interface {
	GetAllNodes(ctx context.Context) ([]NodeConfig, error)
}

type redisRepo struct {
	app *app.App
}

func (r redisRepo) GetAllNodes(ctx context.Context) (result []NodeConfig, err error) {
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

		hgetCmd := r.app.Redis.HGetAll(ctx, MakeNodeHash(h))

		//  partial failures
		if hgetCmd.Err() != nil {
			// handle error

		}

		var nodeModel NodeConfig
		err := hgetCmd.Scan(&nodeModel)
		if err != nil {
			// handle error
		} else {
			result = append(result, nodeModel)
		}

	}

	return
}

func NewNodeRepo() NodeRepo {
	return redisRepo{app: app.GetApp()}
}
