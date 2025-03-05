package media

import (
	"context"
	"encoding/json"

	"github.com/egfanboy/mediapire-common/messaging"
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

	log.Info().Msgf("Node %s is ready", nodeMsg.NodeId)

	syncService, err := NewMediaSyncService(ctx)
	if err != nil {
		return
	}

	err = syncService.HandleNewNode(ctx, nodeMsg.NodeId)
	if err != nil {
		log.Err(err).Msgf("Could not sync media from node %s", nodeMsg.NodeId)
		return
	}

}

func init() {
	rabbitmq.RegisterConsumer(handleNodeReadyMessage, messaging.TopicNodeReady)
}
