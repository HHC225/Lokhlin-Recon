package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ApiKey struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	KeyHash    string             `bson:"keyHash" json:"-"`
	KeyPrefix  string             `bson:"keyPrefix" json:"keyPrefix"`
	Enabled    bool               `bson:"enabled" json:"enabled"`
	CreatedBy  string             `bson:"createdBy" json:"createdBy"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	LastUsedAt *time.Time         `bson:"lastUsedAt,omitempty" json:"lastUsedAt,omitempty"`
}

type CreateApiKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateApiKeyResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	KeyPrefix string `json:"keyPrefix"`
	CreatedAt string `json:"createdAt"`
}

type DeleteApiKeyRequest struct {
	ID string `json:"id" binding:"required"`
}
