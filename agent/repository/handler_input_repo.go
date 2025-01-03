package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"Dana/agent/model"
)

type HandlerInputRepo interface {
	AddServerInput(context.Context, *model.HandlerInput) error
	GetServers(context.Context) ([]*model.HandlerInput, error)
	GetServersByType(context.Context, string) ([]*model.HandlerInput, error)
}

func NewHandlerInputRepo(client *mongo.Client, databaseName, collectionName string) HandlerInputRepo {
	collection := client.Database(databaseName).Collection(collectionName)
	return &handlerInputRepo{
		collection: collection,
	}
}

type handlerInputRepo struct {
	collection *mongo.Collection
}

func (p *handlerInputRepo) AddServerInput(ctx context.Context, handlerInput *model.HandlerInput) error {
	// Create a new document for insertion
	document := bson.M{
		"type": handlerInput.Type,
		"data": handlerInput.Data,
	}

	// Insert the document into the collection
	result, err := p.collection.InsertOne(ctx, document)
	if err != nil {
		return err
	}

	// Set the ID from the inserted document
	handlerInput.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (p *handlerInputRepo) GetServers(ctx context.Context) ([]*model.HandlerInput, error) {
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

	var servers []*model.HandlerInput
	for cursor.Next(ctx) {
		var handlerInput model.HandlerInput
		if err := cursor.Decode(&handlerInput); err != nil {
			return nil, err
		}

		servers = append(servers, &handlerInput)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}

func (p *handlerInputRepo) GetServersByType(ctx context.Context, serverType string) ([]*model.HandlerInput, error) {
	// Filter documents by the "type" field
	filter := bson.M{"type": serverType}

	cursor, err := p.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, ctx)

	var servers []*model.HandlerInput
	for cursor.Next(ctx) {
		var handlerInput model.HandlerInput
		if err := cursor.Decode(&handlerInput); err != nil {
			return nil, err
		}

		servers = append(servers, &handlerInput)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}
