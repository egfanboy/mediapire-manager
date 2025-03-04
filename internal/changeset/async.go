package changeset

import (
	"context"
	"encoding/json"

	"github.com/egfanboy/mediapire-common/messaging"
	"github.com/egfanboy/mediapire-manager/internal/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type updatedMessageHandler struct{}

func (u updatedMessageHandler) HandleMessage(ctx context.Context, msg amqp091.Delivery) {
	log.Info().Msg("Received media updated message")
	var updateMsg messaging.MediaUpdatedMessage

	err := json.Unmarshal(msg.Body, &updateMsg)
	if err != nil {
		log.Err(err).Msg("failed to unmarshal media updated message")

		return
	}

	changesetObjectId, err := primitive.ObjectIDFromHex(updateMsg.ChangesetId)
	if err != nil {
		log.Err(err).Msgf("cannot process media updated message for changeset %s", updateMsg.ChangesetId)
		return
	}

	changesetService, err := newChangesetService(ctx)
	if err != nil {
		log.Err(err).Msgf("cannot process media updated message for changeset %s", updateMsg.ChangesetId)
		return
	}

	changeset, err := changesetService.GetChangesetById(ctx, changesetObjectId)
	if err != nil {
		log.Err(err).Msgf("failed to get changeset %s", updateMsg.ChangesetId)
		return
	}

	repo, err := newChangesetRepository(ctx)
	if err != nil {
		log.Err(err).Msgf("cannot process media updated message for changeset %s", updateMsg.ChangesetId)
		return
	}

	if !updateMsg.Success {
		log.Error().Msgf("Error occured in the updating of media for changeset %s", updateMsg.ChangesetId)

		changeset.Status = StatusFailed
		if updateMsg.FailureReason != nil {
			changeset.FailureReason = *updateMsg.FailureReason
			log.Error().Msgf("failure reason for changeset %s: %s", updateMsg.ChangesetId, *updateMsg.FailureReason)
		}

		err = repo.Save(ctx, changeset)
		if err != nil {
			log.Err(err).Msgf("failed to update changeset %s", updateMsg.ChangesetId)
			return
		}

	} else {
		changeset.Outputs[updateMsg.NodeId] = true

		if changeset.IsDone() {
			changeset.Status = StatusComplete
		}

		err = repo.Save(ctx, changeset)
		if err != nil {
			log.Err(err).Msgf("failed to update changeset %s to %s", updateMsg.ChangesetId, StatusComplete)
			return
		}
	}

	log.Info().Msg("Handled media updated message")
}

func init() {
	rabbitmq.RegisterConsumer(updatedMessageHandler{}.HandleMessage, messaging.TopicMediaUpdated)
}
