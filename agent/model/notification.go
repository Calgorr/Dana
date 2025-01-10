package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Notification struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChannelName string             `json:"channel_name" bson:"channel_name"`
	ChatID      int                `json:"chat_id" bson:"chat_id"`
	CheckName   string             `json:"_check_name" bson:"check_name"`
	Level       string             `json:"_level" bson:"level"`
	Message     string             `json:"_message" bson:"message"`
}
