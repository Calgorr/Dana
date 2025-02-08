package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	"Dana/agent/model"
)

type NetworkRepo interface {
	// CreateNetwork creates a new network
	CreateNetwork(ctx context.Context, network *model.KnownServer) error
	// GetNetwork gets a network by name
	GetNetwork(ctx context.Context, name string) (*model.KnownServer, error)
	// GetNetworks gets all networks
	GetNetworks(ctx context.Context) ([]*model.KnownServer, error)
	// DeleteNetwork deletes a network by name
	DeleteNetwork(ctx context.Context, name string) error
}

type networkRepo struct {
	collection *mongo.Collection
}

func NewNetworkRepo(client *mongo.Client, databaseName, collectionName string) NetworkRepo {
	collection := client.Database(databaseName).Collection(collectionName)
	return &networkRepo{
		collection: collection,
	}
}

func (n *networkRepo) CreateNetwork(ctx context.Context, network *model.KnownServer) error {
	_, err := n.collection.InsertOne(ctx, network)
	if err != nil {
		return err
	}
	return nil
}

func (n *networkRepo) GetNetwork(ctx context.Context, name string) (*model.KnownServer, error) {
	var network model.KnownServer
	err := n.collection.FindOne(ctx, map[string]string{"name": name}).Decode(&network)
	if err != nil {
		return nil, err
	}
	return &network, nil
}

func (n *networkRepo) GetNetworks(ctx context.Context) ([]*model.KnownServer, error) {
	cursor, err := n.collection.Find(ctx, map[string]string{})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, ctx)

	var networks []*model.KnownServer
	for cursor.Next(ctx) {
		var network model.KnownServer
		if err := cursor.Decode(&network); err != nil {
			return nil, err
		}
		networks = append(networks, &network)
	}
	return networks, nil
}

func (n *networkRepo) DeleteNetwork(ctx context.Context, name string) error {
	_, err := n.collection.DeleteOne(ctx, map[string]string{"name": name})
	if err != nil {
		return err
	}
	return nil
}
