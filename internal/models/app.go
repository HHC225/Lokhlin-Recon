package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type App struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	PackageName string             `bson:"package_name" json:"package_name"`
	Version     string             `bson:"version" json:"version"`
	Platform    string             `bson:"platform" json:"platform"`
	Size        int64              `bson:"size" json:"size"`
	Icon        string             `bson:"icon" json:"icon"`
	Description string             `bson:"description" json:"description"`
	Developer   string             `bson:"developer" json:"developer"`
	Category    string             `bson:"category" json:"category"`
	Tags        []string           `bson:"tags" json:"tags"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type AppResult struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Icp         string             `bson:"icp" json:"icp"`
	Company     string             `bson:"company" json:"company"`
	Project     string             `bson:"project" json:"project"`
	Time        string             `bson:"time" json:"time"`
	Tags        []string           `bson:"tags" json:"tags"`
	Category    string             `bson:"category" json:"category"`
	Description string             `bson:"description" json:"description"`
	BundleID    string             `bson:"bundleID" json:"bundleID"`
	Apk         string             `bson:"apk" json:"apk"`
	URL         string             `bson:"url" json:"url"`
}

type AppResponse struct {
	List  []AppResult `json:"list"`
	Total int64       `json:"total"`
}
