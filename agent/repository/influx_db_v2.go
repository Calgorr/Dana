package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"Dana/agent/model"
)

type InfluxDbV2Repo interface {
	AddServerInput(context.Context, *model.InfluxDbV2) error
	GetServers(context.Context) ([]*model.InfluxDbV2, error)
	DeleteServer(context.Context, string) error
}

func NewInfluxDbV2Repo(client *mongo.Client, databaseName, collectionName string) InfluxDbV2Repo {
	collection := client.Database(databaseName).Collection(collectionName)
	return &influxDbV2Repo{
		collection: collection,
	}
}

type influxDbV2Repo struct {
	collection *mongo.Collection
}

func (i *influxDbV2Repo) AddServerInput(ctx context.Context, influxDbV2 *model.InfluxDbV2) error {
	// Create a new document for insertion
	document := bson.M{
		"service_address":         influxDbV2.ServiceAddress,
		"max_undelivered_metrics": influxDbV2.MaxUndeliveredMetrics,
		"read_timeout":            influxDbV2.ReadTimeout,
		"write_timeout":           influxDbV2.WriteTimeout,
		"max_body_size":           influxDbV2.MaxBodySize,
		"bucket_tag":              influxDbV2.BucketTag,
	}

	// Insert the document into the collection
	result, err := i.collection.InsertOne(ctx, document)
	if err != nil {
		return err
	}

	// Set the ID from the inserted document (_id is used as the unique identifier in MongoDB)
	influxDbV2.ID = result.InsertedID.(int)
	return nil
}

func (i *influxDbV2Repo) GetServers(ctx context.Context) ([]*model.InfluxDbV2, error) {
	// Find all documents in the collection
	cursor, err := i.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, ctx)

	var servers []*model.InfluxDbV2
	for cursor.Next(ctx) {
		var influxDbV2 model.InfluxDbV2
		if err := cursor.Decode(&influxDbV2); err != nil {
			return nil, err
		}

		// Append the decoded influxDbV2 to the result slice
		servers = append(servers, &influxDbV2)
	}

	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	return servers, nil
}

func (i *influxDbV2Repo) DeleteServer(ctx context.Context, id string) error {
	// Delete the document with the given ID
	filter := bson.M{"_id": id}
	result, err := i.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("no document found with the given ID")
	}
	return nil
}
