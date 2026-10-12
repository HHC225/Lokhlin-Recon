// models-------------------------------------
// @file      : configuration.go
// @time      : 2025/10/29 21:33
// -------------------------------------------

package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type System struct {
	ModulesConfig string `json:"modulesConfig" bson:"modulesConfig"`
}

type Notification struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `json:"name" bson:"name"`
	Method      string             `json:"method" bson:"method"`
	URL         string             `json:"url" bson:"url"`
	ContentType string             `json:"contentType" bson:"contentType"`
	Data        string             `json:"data" bson:"data"`
	State       bool               `json:"state" bson:"state"`
}

type UpdateNotification struct {
	ID          string `json:"id" bson:"-"`
	Name        string `json:"name" bson:"name"`
	Method      string `json:"method" bson:"method"`
	URL         string `json:"url" bson:"url"`
	ContentType string `json:"contentType" bson:"contentType"`
	Data        string `json:"data" bson:"data"`
	State       bool   `json:"state" bson:"state"`
}

type DepConfig struct {
	Asset                bool   `json:"asset" bson:"asset"`
	Subdomain            bool   `json:"subdomain" bson:"subdomain"`
	SubdomainTakerResult bool   `json:"SubdomainTakerResult" bson:"SubdomainTakerResult"`
	UrlScan              bool   `json:"UrlScan" bson:"UrlScan"`
	Crawler              bool   `json:"crawler" bson:"crawler"`
	SensitiveResult      bool   `json:"SensitiveResult" bson:"SensitiveResult"`
	DirScanResult        bool   `json:"DirScanResult" bson:"DirScanResult"`
	Vulnerability        bool   `json:"vulnerability" bson:"vulnerability"`
	PageMonitoring       bool   `json:"PageMonitoring" bson:"PageMonitoring"`
	Hour                 int    `json:"hour" bson:"hour"`
	Flag                 bool   `json:"flag" bson:"flag"`
	RunNow               bool   `json:"runNow" bson:"runNow"`
	Name                 string `bson:"name" json:"name"`
}
