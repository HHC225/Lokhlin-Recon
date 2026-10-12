// models-------------------------------------
// @file      : dictionary.go
// @time      : 2025/4/28 20:29
// -------------------------------------------

package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Port struct {
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
}

type DictionaryMeta struct {
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	Name     string             `json:"name" bson:"name"`
	Category string             `json:"category" bson:"category"`
	Size     string             `json:"size" bson:"size"`
}

type PortDoc struct {
	ID    primitive.ObjectID `json:"id" bson:"_id"`
	Name  string             `json:"name" bson:"name"`
	Value string             `json:"value" bson:"value"`
}

type UpdatePort struct {
	ID    string `json:"id"`
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
}
