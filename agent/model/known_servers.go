package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type KnownServer struct {
	ID   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name"`
	IP   string             `json:"network_address" bson:"network_address"`
}
