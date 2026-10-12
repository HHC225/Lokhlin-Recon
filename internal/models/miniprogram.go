package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MiniProgram struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	AppID       string             `bson:"appid" json:"appid"`
	Type        string             `bson:"type" json:"type"`
	Version     string             `bson:"version" json:"version"`
	Icon        string             `bson:"icon" json:"icon"`
	Description string             `bson:"description" json:"description"`
	Developer   string             `bson:"developer" json:"developer"`
	Category    string             `bson:"category" json:"category"`
	Tags        []string           `bson:"tags" json:"tags"`
	Status      string             `bson:"status" json:"status"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}
