package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Notification struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChannelName string             `json:"channel_name"`
	Alert       string             `json:"alert"`
	ChatID      int                `json:"chat_id"`
}
