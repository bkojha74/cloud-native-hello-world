package repository

import (
	"context"
	"errors"

	"backend/internal/metrics"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type messageDoc struct {
	ID      string `bson:"_id"`
	Message string `bson:"message"`
}

type MongoMessageRepository struct {
	collection *mongo.Collection
}

func NewMongoMessageRepository(collection *mongo.Collection) *MongoMessageRepository {
	return &MongoMessageRepository{collection: collection}
}

func (r *MongoMessageRepository) GetMessage(ctx context.Context) (string, error) {
	var doc messageDoc
	err := r.collection.FindOne(ctx, bson.M{"_id": "1"}).Decode(&doc)
	if err != nil {
		metrics.DBErrorsTotal.WithLabelValues("mongodb").Inc()
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", ErrNotFound
		}
		return "", err
	}
	metrics.DBReadsTotal.WithLabelValues("mongodb").Inc()
	return doc.Message, nil
}

func (r *MongoMessageRepository) EnsureSeedData(ctx context.Context) error {
	filter := bson.M{"_id": "1"}
	update := bson.M{
		"$setOnInsert": bson.M{
			"_id":     "1",
			"message": "Hello World",
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}
