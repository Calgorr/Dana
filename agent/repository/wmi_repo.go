package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"Dana/agent/model"
)

type WmiRepo interface {
	AddWmiInput(context.Context, *model.Wmi) error
	GetWmis(context.Context) ([]*model.Wmi, error)
	DeleteWmi(context.Context, string) error
}

func NewWmiRepo(client *mongo.Client, databaseName, collectionName string) WmiRepo {
	collection := client.Database(databaseName).Collection(collectionName)
	return &wmiRepo{
		collection: collection,
	}
}

type wmiRepo struct {
	collection *mongo.Collection
}

func (w *wmiRepo) AddWmiInput(ctx context.Context, wmi *model.Wmi) error {
	// Create a new document for insertion
	document := bson.M{
		"host":     wmi.Host,
		"username": wmi.Username,
		"password": wmi.Password,
		"queries":  wmi.Queries,
		"methods":  wmi.Methods,
	}

	// Insert the document into the collection
	result, err := w.collection.InsertOne(ctx, document)
	if err != nil {
		return err
	}

	// Set the ID from the inserted document (MongoDB generates the _id)
	wmi.ID = result.InsertedID.(int)
	return nil
}

func (w *wmiRepo) GetWmis(ctx context.Context) ([]*model.Wmi, error) {
	// Find all documents in the collection
	cursor, err := w.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, ctx)

	var wmis []*model.Wmi
	for cursor.Next(ctx) {
		var wmi model.Wmi
		if err := cursor.Decode(&wmi); err != nil {
			return nil, err
		}

		// Append the decoded wmi to the result slice
		wmis = append(wmis, &wmi)
	}

	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	return wmis, nil
}

func (w *wmiRepo) DeleteWmi(ctx context.Context, id string) error {
	// Delete the document with the given ID
	filter := bson.M{"_id": id}
	result, err := w.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("no document found with the given ID")
	}
	return nil
}
