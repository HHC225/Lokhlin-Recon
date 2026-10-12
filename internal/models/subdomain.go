package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subdomain struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Host       string             `bson:"host" json:"host"`
	Type       string             `bson:"type" json:"type"`
	Value      []string           `bson:"value" json:"value"`
	IP         []string           `bson:"ip" json:"ip"`
	Time       string             `bson:"time" json:"time"`
	Tags       []string           `bson:"tags" json:"tags"`
	Project    string             `bson:"project" json:"project"`
	TaskName   string             `bson:"taskName" json:"taskName"`
	RootDomain string             `bson:"rootDomain" json:"rootDomain"`
}

type SubdomainResponse struct {
	List  []Subdomain `json:"list"`
	Total int64       `json:"total"`
}

type SubdomainTakerResult struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	Input    string             `bson:"input" json:"host"`
	Value    string             `bson:"value" json:"value"`
	Cname    string             `bson:"cname" json:"type"`
	Response string             `bson:"response" json:"response"`
	Tags     []string           `bson:"tags" json:"tags"`
}

type SubdomainTakerResponse struct {
	List  []SubdomainTakerResult `json:"list"`
	Total int64                  `json:"total"`
}
