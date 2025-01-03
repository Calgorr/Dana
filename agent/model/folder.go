package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Folder struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Dashboards []Dashboard        `json:"dashboards" bson:"dashboards"`
}
