package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MPResult struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Icp         string             `bson:"icp" json:"icp"`
	Company     string             `bson:"company" json:"company"`
	Project     string             `bson:"project" json:"project"`
	Time        string             `bson:"time" json:"time"`
	Tags        []string           `bson:"tags" json:"tags"`
	Description string             `bson:"description" json:"description"`
	URL         string             `bson:"url" json:"url"`
}

type MPResponse struct {
	List  []MPResult `json:"list"`
	Total int64      `json:"total"`
}
