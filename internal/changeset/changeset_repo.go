package changeset

import (
	"context"
	"errors"

	mediapireMongo "github.com/egfanboy/mediapire-manager/internal/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type changesetRepository interface {
	Save(ctx context.Context, d *Changeset) error
	GetById(ctx context.Context, objectId primitive.ObjectID) (*Changeset, error)
	GetAll(ctx context.Context) ([]*Changeset, error)
}

type repo struct {
}

func (r *repo) getCollection() *mongo.Collection {
	// TODO: do not ignore error but panic, without causing everything else to break
	collection, _ := mediapireMongo.NewCollection("changesets")

	return collection
}

func (r *repo) Save(ctx context.Context, d *Changeset) error {
	_, err := r.GetById(ctx, d.Id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			_, iErr := r.getCollection().InsertOne(ctx, d)
			return iErr
		}

		// other error not related to the document not existing
		return err
	}

	// Already exists, update it
	_, rErr := r.getCollection().ReplaceOne(ctx, bson.M{"_id": d.Id}, d)
	return rErr
}

func (r *repo) GetById(ctx context.Context, objectId primitive.ObjectID) (*Changeset, error) {
	dl := &Changeset{}
	result := r.getCollection().FindOne(ctx, bson.M{"_id": objectId})

	err := result.Decode(dl)

	return dl, err
}

func (r *repo) GetAll(ctx context.Context) ([]*Changeset, error) {
	cur, err := r.getCollection().Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var result []*Changeset

	err = cur.All(ctx, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func newChangesetRepository(ctx context.Context) (changesetRepository, error) {
	r := &repo{}

	return r, nil
}
