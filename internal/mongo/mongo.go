package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/egfanboy/mediapire-manager/internal/app"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

var mediapireDB *mongo.Database

var (
	errNoClientError = errors.New("mongo client was never initialized")
)

func InitMongo(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	var err error
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(app.GetApp().Config.MongoURI).SetRegistry(mongoRegistry))
	if err != nil {
		return err
	}

	mediapireDB = mongoClient.Database("mediapire-manager")

	return nil
}

func CleanUpMongo(ctx context.Context) error {
	if mongoClient == nil {
		return errors.New("cannot disconnect from mongo since it was never initialized")
	}
	return mongoClient.Disconnect(ctx)
}

func NewCollection(collection string) (*mongo.Collection, error) {
	if mongoClient == nil {
		return nil, errNoClientError
	}

	// TODO: may need to handle db being nil
	return mediapireDB.Collection(collection), nil
}
