// Package models -----------------------------
// @file      : task.go
// @time      : 2025/5/4 22:04
// -------------------------------------------
package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	ID             primitive.ObjectID       `bson:"_id,omitempty" json:"id"`
	TaskID         string                   `bson:"id"`
	Name           string                   `bson:"name" json:"name"`
	Node           []string                 `bson:"node" json:"node"`
	AllNode        bool                     `bson:"allNode" json:"allNode"`
	ScheduledTasks bool                     `bson:"scheduledTasks" json:"scheduledTasks"`
	Target         string                   `bson:"target" json:"target"`
	Ignore         string                   `bson:"ignore" json:"ignore"`
	Template       string                   `bson:"template" json:"template"`
	Project        []string                 `bson:"project" json:"project"`
	Search         string                   `bson:"search" json:"search"`
	CycleType      string                   `bson:"cycleType" json:"cycleType"`
	Hour           int                      `bson:"hour" json:"hour"`
	Minute         int                      `bson:"minute" json:"minute"`
	Day            int                      `bson:"day" json:"day"`
	Week           int                      `bson:"week" json:"week"`
	Status         int                      `bson:"status" json:"status"`
	Progress       int                      `bson:"progress" json:"progress"`
	CreatTime      string                   `bson:"creatTime" json:"creatTime"`
	EndTime        string                   `bson:"endTime" json:"endTime"`
	TaskNum        int                      `bson:"taskNum" json:"taskNum"`
	Duplicates     string                   `bson:"duplicates" json:"duplicates"`
	Filter         map[string][]interface{} `bson:"filter" json:"filter"`
	TargetNumber   int                      `bson:"targetNumber" json:"targetNumber"`
	TargetIds      []string                 `bson:"targetIds" json:"targetIds"`
	TargetTp       string                   `bson:"targetTp" json:"targetTp"`
	TargetSource   string                   `bson:"targetSource" json:"targetSource"`
	BindProject    string                   `bson:"bindProject" json:"bindProject"`
	Type           string                   `bson:"type" json:"type"`
	IsStart        bool                     `bson:"isStart" json:"isStart"`
	LastTime       string                   `bson:"lastTime"  json:"lastTime"`
	NextTime       string                   `bson:"nextTime" json:"nextTime"`
	Cycle          string                   `bson:"-" json:"cycle"`
}

type TaskProgress struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TargetHandler       int                `bson:"targetHandler" json:"targetHandler"`
	SubdomainScan       int                `bson:"subdomainScan" json:"subdomainScan"`
	SubdomainSecurity   int                `bson:"subdomainSecurity" json:"subdomainSecurity"`
	PortScanPreparation int                `bson:"portScanPreparation" json:"portScanPreparation"`
	PortScan            int                `bson:"portScan" json:"portScan"`
	PortFingerprint     int                `bson:"portFingerprint" json:"portFingerprint"`
	AssetMapping        int                `bson:"assetMapping" json:"assetMapping"`
	AssetHandle         int                `bson:"assetHandle" json:"assetHandle"`
	URLScan             int                `bson:"urlScan" json:"urlScan"`
	WebCrawler          int                `bson:"webCrawler" json:"webCrawler"`
	URLSecurity         int                `bson:"urlSecurity" json:"urlSecurity"`
	DirScan             int                `bson:"dirScan" json:"dirScan"`
	VulnerabilityScan   int                `bson:"vulnerabilityScan" json:"vulnerabilityScan"`
	All                 int                `bson:"all" json:"all"`
	Target              string             `bson:"target" json:"target"`
	Node                string             `bson:"node" json:"node"`
	Children            []TaskProgress     `bson:"children" json:"children"`
}

type TaskProgressDetail struct {
	ID                  primitive.ObjectID   `bson:"_id,omitempty" json:"_id"`
	TargetHandler       []string             `json:"TargetHandler"`       // [start, end]
	SubdomainScan       []string             `json:"SubdomainScan"`       // [start, end]
	SubdomainSecurity   []string             `json:"SubdomainSecurity"`   // [start, end]
	PortScanPreparation []string             `json:"PortScanPreparation"` // [start, end]
	PortScan            []string             `json:"PortScan"`            // [start, end]
	PortFingerprint     []string             `json:"PortFingerprint"`     // [start, end]
	AssetMapping        []string             `json:"AssetMapping"`        // [start, end]
	AssetHandle         []string             `json:"AssetHandle"`         // [start, end]
	URLScan             []string             `json:"URLScan"`             // [start, end]
	WebCrawler          []string             `json:"WebCrawler"`          // [start, end]
	URLSecurity         []string             `json:"URLSecurity"`         // [start, end]
	DirScan             []string             `json:"DirScan"`             // [start, end]
	VulnerabilityScan   []string             `json:"VulnerabilityScan"`   // [start, end]
	All                 []string             `json:"All"`                 // [start, end]
	Target              string               `json:"target"`
	Node                string               `json:"node"`
	Children            []TaskProgressDetail `json:"children,omitempty"`
}

type ListRequest struct {
	Search    string `json:"search" binding:"omitempty" example:"test"`
	PageIndex int    `json:"pageIndex" binding:"required,min=1" example:"1"`
	PageSize  int    `json:"pageSize" binding:"required,min=1" example:"10"`
}

type ScheduledTask struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	TaskID         string             `bson:"id" json:"id"`
	Name           string             `bson:"name" json:"name"`
	Type           string             `bson:"type" json:"type"`
	LastTime       string             `bson:"lastTime" json:"lastTime"`
	NextTime       string             `bson:"nextTime" json:"nextTime"`
	State          bool               `bson:"scheduledTasks" json:"state"`
	Node           []string           `bson:"node" json:"node"`
	Cycle          string             `bson:"-" json:"cycle"`
	AllNode        bool               `bson:"allNode" json:"allNode"`
	RunnerID       string             `bson:"runner_id" json:"runner_id"`
	Project        []string           `bson:"project" json:"project"`
	TargetSource   string             `bson:"targetSource" json:"targetSource"`
	Day            int                `bson:"day" json:"day"`
	Minute         int                `bson:"minute" json:"minute"`
	Hour           int                `bson:"hour" json:"hour"`
	Search         string             `bson:"search" json:"search"`
	CycleType      string             `bson:"cycleType" json:"cycleType"`
	ScheduledTasks bool               `bson:"scheduledTasks" json:"scheduledTasks"`
	Target         string             `bson:"target" json:"target"`
	Ignore         string             `bson:"ignore" json:"ignore"`
	Template       string             `bson:"template" json:"template"`
	Duplicates     string             `bson:"duplicates" json:"duplicates"`
	Week           int                `bson:"week" json:"week"`
}

type ScheduledTaskListRequest struct {
	Search    string `json:"search" binding:"omitempty" example:"test"`
	PageIndex int    `json:"pageIndex" binding:"omitempty,min=1" example:"1"`
	PageSize  int    `json:"pageSize" binding:"omitempty,min=1" example:"10"`
}

type ScheduledTaskDetailRequest struct {
	ID string `json:"id" binding:"required" example:"507f1f77bcf86cd799439011"`
}

type ScheduledTaskDeleteRequest struct {
	IDs []string `json:"ids" binding:"required" example:"[\"507f1f77bcf86cd799439011\"]"`
}

type ScheduledTaskAddRequest struct {
	Name           string   `json:"name" binding:"required" example:"Scheduled scan task"`
	Type           string   `json:"type" binding:"omitempty" example:"scan"`
	Node           []string `json:"node" binding:"required" example:"[\"node1\", \"node2\"]"`
	AllNode        bool     `json:"allNode" binding:"omitempty" example:"true"`
	Project        []string `json:"project" binding:"omitempty" example:"[\"project1\"]"`
	TargetSource   string   `json:"targetSource" binding:"omitempty" example:"general"`
	Day            int      `json:"day" binding:"omitempty" example:"1"`
	Minute         int      `json:"minute" binding:"omitempty" example:"0"`
	Hour           int      `json:"hour" binding:"omitempty" example:"1"`
	Search         string   `json:"search" binding:"omitempty" example:""`
	CycleType      string   `json:"cycleType" binding:"omitempty" example:"nhours"`
	ScheduledTasks bool     `json:"scheduledTasks" binding:"omitempty" example:"true"`
	Target         string   `json:"target" binding:"omitempty" example:"example.com"`
	Ignore         string   `json:"ignore" binding:"omitempty" example:""`
	Template       string   `json:"template" binding:"omitempty" example:"default"`
	Duplicates     string   `json:"duplicates" binding:"omitempty" example:"true"`
	Week           int      `json:"week" binding:"omitempty" example:"1"`
}

type ScheduledTaskUpdateRequest struct {
	ID             string   `json:"id" binding:"required" example:"507f1f77bcf86cd799439011"`
	Name           string   `json:"name" binding:"omitempty" example:"Scheduled scan task"`
	Type           string   `json:"type" binding:"omitempty" example:"scan"`
	Node           []string `json:"node" binding:"omitempty" example:"[\"node1\", \"node2\"]"`
	AllNode        bool     `json:"allNode" binding:"omitempty" example:"true"`
	Project        []string `json:"project" binding:"omitempty" example:"[\"project1\"]"`
	TargetSource   string   `json:"targetSource" binding:"omitempty" example:"general"`
	Day            int      `json:"day" binding:"omitempty" example:"1"`
	Minute         int      `json:"minute" binding:"omitempty" example:"0"`
	Hour           int      `json:"hour" binding:"omitempty" example:"1"`
	Search         string   `json:"search" binding:"omitempty" example:""`
	CycleType      string   `json:"cycleType" binding:"omitempty" example:"nhours"`
	ScheduledTasks bool     `json:"scheduledTasks" binding:"omitempty" example:"true"`
	Target         string   `json:"target" binding:"omitempty" example:"example.com"`
	Ignore         string   `json:"ignore" binding:"omitempty" example:""`
	Template       string   `json:"template" binding:"omitempty" example:"default"`
	Duplicates     string   `json:"duplicates" binding:"omitempty" example:"true"`
	Week           int      `json:"week" binding:"omitempty" example:"1"`
}

type PageMonitoringTask struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	URL     string             `bson:"url" json:"url"`
	Hash    []string           `bson:"hash" json:"hash"`
	MD5     string             `bson:"md5" json:"md5"`
	State   int                `bson:"state" json:"state"`
	Project string             `bson:"project" json:"project"`
	Time    string             `bson:"time" json:"time"`
}

type PageMonitoringListRequest struct {
	Search    string `json:"search" binding:"omitempty" example:"example.com"`
	PageIndex int    `json:"pageIndex" binding:"omitempty,min=1" example:"1"`
	PageSize  int    `json:"pageSize" binding:"omitempty,min=1" example:"10"`
}

type PageMonitoringAddRequest struct {
	URL string `json:"url" binding:"required" example:"https://example.com"`
}

type PageMonitoringDeleteRequest struct {
	IDs []string `json:"ids" binding:"required" example:"[\"507f1f77bcf86cd799439011\"]"`
}

type PageMonitoringUpdateRequest struct {
	Hour    int      `json:"hour" binding:"omitempty" example:"24"`
	Node    []string `json:"node" binding:"omitempty" example:"[\"node1\"]"`
	AllNode bool     `json:"allNode" binding:"omitempty" example:"true"`
	State   bool     `json:"state" binding:"omitempty" example:"true"`
}
