package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Dashboard struct {
	ID   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Data string             `json:"data" bson:"data"`
}
