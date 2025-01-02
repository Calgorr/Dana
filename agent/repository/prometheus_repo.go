package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"Dana/agent/model"
)

type PrometheusRepo interface {
	AddServerInput(context.Context, *model.Prometheus) error
	GetServers(context.Context) ([]*model.Prometheus, error)
	DeleteServer(context.Context, string) error
}

func NewPrometheusRepo(client *mongo.Client, databaseName, collectionName string) PrometheusRepo {
	collection := client.Database(databaseName).Collection(collectionName)
	return &prometheusRepo{
		collection: collection,
	}
}

type prometheusRepo struct {
	collection *mongo.Collection
}

func (p *prometheusRepo) AddServerInput(ctx context.Context, prometheus *model.Prometheus) error {
	// Create a new document for insertion
	document := bson.M{
		"urls":           prometheus.URLs,
		"metric_version": prometheus.MetricVersion,
		"timeout":        prometheus.Timeout,
	}

	// Insert the document into the collection
	result, err := p.collection.InsertOne(ctx, document)
	if err != nil {
		return err
	}

	// Set the ID from the inserted document (_id is used as the unique identifier in MongoDB)
	prometheus.ID = result.InsertedID.(int)
	return nil
}

func (p *prometheusRepo) GetServers(ctx context.Context) ([]*model.Prometheus, error) {
	// Find all documents in the collection
	cursor, err := p.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, ctx)

	var servers []*model.Prometheus
	for cursor.Next(ctx) {
		var prometheus model.Prometheus
		if err := cursor.Decode(&prometheus); err != nil {
			return nil, err
		}

		// Append the decoded prometheus to the result slice
		servers = append(servers, &prometheus)
	}

	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	return servers, nil
}

func (p *prometheusRepo) DeleteServer(ctx context.Context, id string) error {
	// Delete the document with the given ID
	filter := bson.M{"_id": id}
	result, err := p.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("no document found with the given ID")
	}
	return nil
}
