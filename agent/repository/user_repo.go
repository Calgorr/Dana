package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"Dana/agent/model"
)

type UserRepo interface {
	AddUser(ctx context.Context, user *model.User) error
	UserAuth(ctx context.Context, username, password string) error
}

type userRepo struct {
	collection *mongo.Collection
}

func NewUserRepo(client *mongo.Client, databaseName, collectionName string) UserRepo {
	collection := client.Database(databaseName).Collection(collectionName)
	return &userRepo{
		collection: collection,
	}
}

func (r *userRepo) AddUser(ctx context.Context, user *model.User) error {

	// Set creation time
	user.CreatedAt = primitive.NewDateTimeFromTime(time.Now())

	// Check if username already exists
	count, err := r.collection.CountDocuments(ctx, bson.M{"username": user.Username})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("username already exists")
	}

	// Insert the user
	_, err = r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepo) UserAuth(ctx context.Context, username, password string) error {

	var user model.User
	err := r.collection.FindOne(ctx, bson.M{
		"username": username,
		"password": password,
	}).Decode(&user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errors.New("invalid username or password")
		}
		return err
	}
	return nil
}
