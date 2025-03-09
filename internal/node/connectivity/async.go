package connectivity

import (
	"context"
	"encoding/json"

	"github.com/egfanboy/mediapire-common/messaging"
	"github.com/egfanboy/mediapire-manager/internal/media"
	"github.com/egfanboy/mediapire-manager/internal/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

func handleNodeReadyMessage(ctx context.Context, msg amqp091.Delivery) {
	var nodeMsg messaging.NodeReadyMessage

	err := json.Unmarshal(msg.Body, &nodeMsg)
	if err != nil {
		return
	}

	log.Info().Msgf("Node %s is ready", nodeMsg.Id)

	syncService, err := media.NewMediaSyncService(ctx)
	if err != nil {
		return
	}

	// start goroutine watchers for this node
	WatchNode(nodeMsg.Name, nodeMsg.Id)

	err = syncService.HandleNewNode(ctx, nodeMsg.Id)
	if err != nil {
		log.Err(err).Msgf("Could not sync media from node %s", nodeMsg.Id)
		return
	}
}

func init() {
	rabbitmq.RegisterConsumer(handleNodeReadyMessage, messaging.TopicNodeReady)
}
