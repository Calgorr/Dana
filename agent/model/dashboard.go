package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Dashboard struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Panels    []Panel            `json:"panels" bson:"panels"`
	Variables []Variable         `json:"variables" bson:"variables"`
}

type Panel struct {
	Name   string   `json:"name" bson:"name"`
	Query  []string `json:"query" bson:"query"`
	Index  int      `json:"index" bson:"index"`
	Colors []string `json:"color" bson:"color"`
}

type Variable struct {
	Name  string `json:"name" bson:"name"`
	Query string `json:"query" bson:"query"`
}
