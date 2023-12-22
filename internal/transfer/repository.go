package transfer

import (
	"context"
	"errors"

	mediapireMongo "github.com/egfanboy/mediapire-manager/internal/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransferRepository interface {
	Save(ctx context.Context, d *Transfer) error
	GetById(ctx context.Context, objectId primitive.ObjectID) (*Transfer, error)
}

type repo struct {
}

func (r *repo) getCollection() *mongo.Collection {
	// TODO: do not ignore error but panic, without causing everything else to break
	collection, _ := mediapireMongo.NewCollection("transfers")

	return collection
}

func (r *repo) Save(ctx context.Context, d *Transfer) error {
	_, err := r.GetById(ctx, d.Id)

	// Already exists, update it
	if err == nil {
		_, err := r.getCollection().ReplaceOne(ctx, bson.M{"_id": d.Id}, d)
		return err
	}

	if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
		_, err := r.getCollection().InsertOne(ctx, d)
		return err
	}

	// other error not related to the document not existing
	if err != nil {
		return err
	}

	return nil
}

func (r *repo) GetById(ctx context.Context, objectId primitive.ObjectID) (*Transfer, error) {
	dl := &Transfer{}
	result := r.getCollection().FindOne(ctx, bson.M{"_id": objectId})

	err := result.Decode(dl)

	return dl, err
}

func NewTransferRepository(ctx context.Context) (TransferRepository, error) {
	r := &repo{}

	return r, nil
}
