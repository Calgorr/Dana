package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	"Dana/agent/model"
)

type NotificationRepo interface {
	CreateNotification(ctx context.Context, notification *model.Notification) error
	GetNotification(ctx context.Context, channelName string) (*model.Notification, error)
	DeleteNotification(ctx context.Context, channelName string) error
}

type notificationRepo struct {
	notificationCollection *mongo.Collection
}

func NewNotificationRepo(client *mongo.Client, databaseName, collectionName string) NotificationRepo {
	notificationCollection := client.Database(databaseName).Collection(collectionName)
	return &notificationRepo{
		notificationCollection: notificationCollection,
	}
}

func (r *notificationRepo) CreateNotification(ctx context.Context, notification *model.Notification) error {
	_, err := r.notificationCollection.InsertOne(ctx, notification)
	return err
}

func (r *notificationRepo) GetNotification(ctx context.Context, channelName string) (*model.Notification, error) {
	var notification model.Notification
	err := r.notificationCollection.FindOne(ctx, map[string]string{"channel_name": channelName}).Decode(&notification)
	return &notification, err
}

func (r *notificationRepo) DeleteNotification(ctx context.Context, channelName string) error {
	_, err := r.notificationCollection.DeleteOne(ctx, map[string]string{"channel_name": channelName})
	return err
}
