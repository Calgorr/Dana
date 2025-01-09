package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"Dana/agent/model"
)

type FolderRepo interface {
	// CreateFolder creates a new folder
	CreateFolder(ctx context.Context, folder *model.Folder) (primitive.ObjectID, error)
	// GetFolder gets a folder by id
	GetFolder(ctx context.Context, id string) (*model.Folder, error)
	// UpdateDashboardInFolder updates a dashboard in a folder by folder id and dashboard id
	UpdateDashboardInFolder(ctx context.Context, folderID string, dashboardID string, dashboard *model.Dashboard) error
	// DeleteFolder deletes a folder by id
	DeleteFolder(ctx context.Context, id string) error
	// GetFolders gets all folders
	GetFolders(ctx context.Context) ([]*model.Folder, error)
}

func NewFolderRepo(client *mongo.Client, databaseName, collectionName string) FolderRepo {
	collection := client.Database(databaseName).Collection(collectionName)
	return &folderRepo{
		collection: collection,
	}
}

type folderRepo struct {
	collection *mongo.Collection
}

func (f *folderRepo) CreateFolder(ctx context.Context, folder *model.Folder) (primitive.ObjectID, error) {
	document := bson.M{
		"dashboards": folder.Dashboards,
	}

	result, err := f.collection.InsertOne(ctx, document)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (f *folderRepo) GetFolder(ctx context.Context, id string) (*model.Folder, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	var folder model.Folder

	err = f.collection.FindOne(ctx, filter).Decode(&folder)
	if err != nil {
		return nil, err
	}

	return &folder, nil
}

func (f *folderRepo) UpdateDashboardInFolder(ctx context.Context, folderID string, dashboardID string, dashboard *model.Dashboard) error {
	folderObjectID, err := primitive.ObjectIDFromHex(folderID)
	if err != nil {
		return err
	}

	dashboardObjectID, err := primitive.ObjectIDFromHex(dashboardID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": folderObjectID}
	update := bson.M{
		"$set": bson.M{
			"dashboards.$[elem].name":      dashboard.Name,
			"dashboards.$[elem].panels":    dashboard.Panels,
			"dashboards.$[elem].variables": dashboard.Variables,
		},
	}
	arrayFilters := options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem._id": dashboardObjectID},
		},
	}

	_, err = f.collection.UpdateOne(ctx, filter, update, &options.UpdateOptions{
		ArrayFilters: &arrayFilters,
	})
	return err
}

func (f *folderRepo) DeleteFolder(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	_, err = f.collection.DeleteOne(ctx, filter)
	return err
}

func (f *folderRepo) GetFolders(ctx context.Context) ([]*model.Folder, error) {
	cursor, err := f.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, ctx)

	var folders []*model.Folder
	for cursor.Next(ctx) {
		var folder model.Folder
		if err := cursor.Decode(&folder); err != nil {
			return nil, err
		}

		folders = append(folders, &folder)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return folders, nil
}
