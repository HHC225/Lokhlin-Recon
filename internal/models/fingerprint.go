// models-------------------------------------
// @file      : fingerprint.go
// @time      : 2025/4/28 20:40
// -------------------------------------------

package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type FingerprintRule struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty" yaml:"-"`
	Name           string             `json:"name" bson:"name" yaml:"name"`
	Rule           string             `json:"rule" bson:"rule" yaml:"rule"`
	FingerprintId  string             `json:"fingerprint_id" bson:"fingerprint_id" yaml:"fingerprint_id"`
	Category       string             `json:"category" bson:"category" yaml:"category"`
	ParentCategory string             `json:"parent_category" bson:"parent_category" yaml:"parent_category"`
	Amount         int                `json:"amount" bson:"amount" yaml:"amount"`
	State          bool               `json:"state" bson:"state" yaml:"state"`
	Company        string             `json:"company" bson:"company" yaml:"company"`
}

type FingerprintQuery struct {
	Search    string `json:"search"`
	PageIndex int    `json:"pageIndex"`
	PageSize  int    `json:"pageSize"`
}
