package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type HandlerInput struct {
	ID   primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	Type string                 `json:"type" bson:"type"`
	Data map[string]interface{} `json:"data" bson:"data"`
}
