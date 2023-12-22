package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/egfanboy/mediapire-common/messaging"
	"github.com/egfanboy/mediapire-manager/internal/app"
	"github.com/egfanboy/mediapire-manager/internal/rabbitmq"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

func handleTransferUpdateMessage(ctx context.Context, msg amqp091.Delivery) {
	msg.Ack(false)

	var updateMsg messaging.TransferUpdateMessage

	err := json.Unmarshal(msg.Body, &updateMsg)
	if err != nil {
		log.Err(err).Msg("failed to unmarshal update transfer message")

		return
	}

	transferRepo, err := NewTransferRepository(ctx)
	if err != nil {
		log.Err(err).Msg("failed to instantiate transfer repository")

		return
	}

	transferObjectId, err := primitive.ObjectIDFromHex(updateMsg.TransferId)
	if err != nil {
		log.Err(err).Msgf("failed to convert transfer id %s to object id", updateMsg.TransferId)

		return
	}

	transferRecord, err := transferRepo.GetById(ctx, transferObjectId)
	if err != nil {
		log.Err(err).Msgf("failed to find transfer with id %s", updateMsg.TransferId)

		return
	}

	if transferRecord.Status == StatusPending || transferRecord.Status == StatusInProgress {
		handleTransferInProgress(ctx, updateMsg, transferRecord)
	}
}

func handleTransferInProgress(ctx context.Context, updateMsg messaging.TransferUpdateMessage, transferRecord *Transfer) error {
	transferRepo, err := NewTransferRepository(ctx)
	if err != nil {
		log.Err(err).Msg("failed to instantiate transfer repository")

		return err
	}

	// Set the record to failed if it already isn't
	if !transferRecord.DidFail() && !updateMsg.Success {
		transferRecord.Status = StatusFailed
		transferRecord.FailureReason = updateMsg.FailureReason
	}

	// set that the current node has been handled
	transferRecord.Outputs[updateMsg.NodeId] = true

	err = transferRepo.Save(ctx, transferRecord)
	if err != nil {
		log.Err(err).Msgf("failed to update transfer with id %s", updateMsg.TransferId)

		return err
	}

	// if the transfer hasn't failed and we handled all nodes set it to processing complete
	if !transferRecord.DidFail() && transferRecord.AllNodesHandled() {
		transferRecord.Status = StatusProcessComplete

		handleProcessedtransfer(ctx, transferRecord)
		// trigger process to download all media from the node
		err = transferRepo.Save(ctx, transferRecord)
		if err != nil {
			log.Err(err).Msgf("failed to update transfer with id %s to processing complete", updateMsg.TransferId)

			return err
		}

	}

	return nil
}

func handleProcessedtransfer(ctx context.Context, transferRecord *Transfer) {
	transferRepo, err := NewTransferRepository(ctx)
	if err != nil {
		log.Err(err).Msg("failed to instantiate download repository")

		return
	}

	nodeIds := make([]uuid.UUID, 0)
	for k, v := range transferRecord.Outputs {
		// should not be the case but check for sanity purposes
		if !v {
			errMsg := fmt.Errorf("cannot process transfer with id %q since the content on node %q is not ready", transferRecord.Id.Hex(), k)
			log.Err(errMsg)

			transferRecord.Status = StatusFailed
			transferRecord.FailureReason = err.Error()

			err = transferRepo.Save(ctx, transferRecord)
			if err != nil {
				log.Err(err)
			}

			return
		}

		nodeIds = append(nodeIds, k)
	}

	content, err := mediaDownloader{}.Download(ctx, transferRecord.Id, nodeIds)
	if err != nil {
		log.Err(err).Msgf("cannot process transfer with id %q", transferRecord.Id.Hex())

		transferRecord.Status = StatusFailed
		transferRecord.FailureReason = err.Error()

		err = transferRepo.Save(ctx, transferRecord)
		if err != nil {
			log.Err(err).Msg("failed to save transfer record")
		}

		return
	}

	// we are the target, save the content to a file
	if app.GetApp().NodeId == transferRecord.TargetId {
		saveContent(ctx, transferRecord, content)

		return
	}

	// another node is the target, send a message and let them handle the content
	sendTransferReadyMessage(ctx, transferRecord, content)
}

// TODO: implement
func sendTransferReadyMessage(ctx context.Context, transfer *Transfer, content []byte) {
	return
}

func saveContent(ctx context.Context, transfer *Transfer, content []byte) {
	transferRepo, err := NewTransferRepository(ctx)
	if err != nil {
		log.Err(err).Msg("failed to instantiate transfer repository")

		return
	}

	file, err := os.Create(path.Join(app.GetApp().Config.DownloadPath, transfer.Id.Hex()+".zip"))
	if err != nil {
		transfer.SetFailed(err.Error())

		err := transferRepo.Save(ctx, transfer)
		if err != nil {
			log.Err(err).Msg("failed to save transfer record")
		}
		return
	}

	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		msg := "Failed to write content to file"
		log.Err(err).Msg(msg)
		transfer.SetFailed(msg)

		err := transferRepo.Save(ctx, transfer)
		if err != nil {
			log.Err(err).Msg("failed to save transfer record")
		}
		return
	}

	err = file.Sync()
	if err != nil {
		msg := "failed to commit file content to disk"
		log.Err(err).Msg(msg)
		transfer.SetFailed(msg)

		err := transferRepo.Save(ctx, transfer)
		if err != nil {
			log.Err(err).Msg("failed to save transfer record")
		}
		return
	}

	// all is good, set transfer record to complete
	transfer.Status = StatusComplete

	err = transferRepo.Save(ctx, transfer)
	if err != nil {
		log.Err(err).Msg("failed to save transfer record")
	}
}

func init() {
	rabbitmq.RegisterConsumer(handleTransferUpdateMessage, messaging.TopicTransferUpdate)
}
