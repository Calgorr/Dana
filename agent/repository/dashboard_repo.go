package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"Dana/agent/model"
)

type DashboardRepo interface {
	// CreateDashboard creates a new dashboard
	CreateDashboard(ctx context.Context, dashboard *model.Dashboard) (primitive.ObjectID, error)
	// GetDashboard gets a dashboard by id
	GetDashboard(ctx context.Context, id string) (*model.Dashboard, error)
	// UpdateDashboard updates a dashboard by id
	UpdateDashboard(ctx context.Context, dashboard *model.Dashboard, dashboardID primitive.ObjectID) error
	// DeleteDashboard deletes a dashboard by id
	DeleteDashboard(ctx context.Context, id string) error
	//GetDashboards gets all dashboards
	GetDashboards(ctx context.Context) ([]*model.Dashboard, error)
}

func NewDashboardRepo(client *mongo.Client, databaseName, collectionName string) DashboardRepo {
	collection := client.Database(databaseName).Collection(collectionName)
	return &dashboardRepo{
		collection: collection,
	}
}

type dashboardRepo struct {
	collection *mongo.Collection
}

func (d *dashboardRepo) CreateDashboard(ctx context.Context, dashboard *model.Dashboard) (primitive.ObjectID, error) {
	// Create a new document for insertion
	document := bson.M{
		"name":      dashboard.Name,
		"panels":    dashboard.Panels,
		"variables": dashboard.Variables,
	}

	// Insert the document into the collection
	result, err := d.collection.InsertOne(ctx, document)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (d *dashboardRepo) GetDashboard(ctx context.Context, id string) (*model.Dashboard, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	var dashboard model.Dashboard

	err = d.collection.FindOne(ctx, filter).Decode(&dashboard)
	if err != nil {
		return nil, err
	}

	return &dashboard, nil
}

func (d *dashboardRepo) UpdateDashboard(ctx context.Context, dashboard *model.Dashboard, dashboardID primitive.ObjectID) error {
	filter := bson.M{"_id": dashboardID}
	updateFields := bson.M{}

	if dashboard.Name != "" {
		updateFields["name"] = dashboard.Name
	}
	if dashboard.Panels != nil && len(dashboard.Panels) > 0 {
		updateFields["panels"] = dashboard.Panels
	}
	if dashboard.Variables != nil && len(dashboard.Variables) > 0 {
		updateFields["variables"] = dashboard.Variables
	}

	if len(updateFields) == 0 {
		return nil
	}

	update := bson.M{"$set": updateFields}

	_, err := d.collection.UpdateOne(ctx, filter, update)
	return err
}

func (d *dashboardRepo) DeleteDashboard(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	_, err = d.collection.DeleteOne(ctx, filter)
	return err
}

func (d *dashboardRepo) GetDashboards(ctx context.Context) ([]*model.Dashboard, error) {
	// Find all documents in the collection
	cursor, err := d.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, ctx)

	var dashboards []*model.Dashboard
	for cursor.Next(ctx) {
		var dashboard model.Dashboard
		if err := cursor.Decode(&dashboard); err != nil {
			return nil, err
		}

		dashboards = append(dashboards, &dashboard)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return dashboards, nil
}
