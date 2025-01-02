package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	_ "go.mongodb.org/mongo-driver/mongo/options"

	"Dana/agent/model"
)

type SnmpRepo interface {
	AddSnmpInput(context.Context, *model.Snmp) error
	GetSnmps(context.Context) ([]*model.Snmp, error)
	DeleteSnmp(context.Context, string) error
}

func NewSnmpRepo(client *mongo.Client, databaseName, collectionName string) SnmpRepo {
	collection := client.Database(databaseName).Collection(collectionName)
	return &snmpRepo{
		collection: collection,
	}
}

type snmpRepo struct {
	collection *mongo.Collection
}

func (s *snmpRepo) AddSnmpInput(ctx context.Context, snmp *model.Snmp) error {
	// Create a new document for insertion
	document := bson.M{
		"service_address": snmp.ServiceAddress,
		"path":            snmp.Path,
		"timeout":         snmp.Timeout,
		"version":         snmp.Version,
		"sec_name":        snmp.SecName,
		"auth_protocol":   snmp.AuthProtocol,
		"auth_password":   snmp.AuthPassword,
		"sec_level":       snmp.SecLevel,
		"priv_protocol":   snmp.PrivProtocol,
		"priv_password":   snmp.PrivPassword,
	}

	// Insert the document into the collection
	result, err := s.collection.InsertOne(ctx, document)
	if err != nil {
		return err
	}

	// Set the ID from the inserted document (MongoDB generates the _id)
	snmp.ID = result.InsertedID.(int)
	return nil
}

func (s *snmpRepo) GetSnmps(ctx context.Context) ([]*model.Snmp, error) {
	// Find all documents in the collection
	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, ctx)

	var snmps []*model.Snmp
	for cursor.Next(ctx) {
		var snmp model.Snmp
		if err := cursor.Decode(&snmp); err != nil {
			return nil, err
		}

		// Append the decoded snmp to the result slice
		snmps = append(snmps, &snmp)
	}

	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	return snmps, nil
}

func (s *snmpRepo) DeleteSnmp(ctx context.Context, id string) error {
	// Delete the document with the given ID
	filter := bson.M{"_id": id}
	result, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("no document found with the given ID")
	}
	return nil
}
