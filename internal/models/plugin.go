// Package models -----------------------------
// @file      : plugin.go
// @time      : 2025/4/27 17:20
// -------------------------------------------
package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Plugin struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Module        string             `json:"module"`
	Name          string             `json:"name"`
	Hash          string             `json:"hash"`
	Parameter     string             `json:"parameter"`
	Help          string             `json:"help"`
	Status        bool               `json:"status"`
	Type          string             `bson:"type" json:"type"`
	LastTime      string             `bson:"lastTime" json:"lastTime"`
	NextTime      string             `bson:"nextTime" json:"nextTime"`
	Introduction  string             `json:"introduction"`
	ParameterList string             `json:"parameterList"`
	IsSystem      bool               `json:"isSystem"`
	Version       string             `json:"version"`
	Source        string             `json:"source"`
}

type PluginListRequest struct {
	PageIndex int    `json:"pageIndex" binding:"required,min=1"`
	PageSize  int    `json:"pageSize" binding:"required,min=1"`
	Type      string `json:"type"`
	Search    string `json:"search"`
}

type PluginListResponse struct {
	List  []Plugin `json:"list"`
	Total int64    `json:"total"`
}

type PluginDetailRequest struct {
	ID string `json:"id" binding:"required"`
}

type PluginSaveRequest struct {
	ID            string `json:"id"`
	Name          string `json:"name" binding:"required"`
	Module        string `json:"module"`
	Hash          string `json:"hash"`
	Parameter     string `json:"parameter"`
	Help          string `json:"help"`
	Introduction  string `json:"introduction"`
	ParameterList string `json:"parameterList"`
	Source        string `json:"source"`
	Version       string `json:"version"`
	IsSystem      bool   `json:"isSystem"`
	Type          string `json:"type"`
	Key           string `json:"key"`
}

type PluginDeleteRequest struct {
	Data []struct {
		Hash   string `json:"hash" binding:"required"`
		Module string `json:"module" binding:"required"`
	} `json:"data" binding:"required"`
}

type PluginLogRequest struct {
	Module string `json:"module"`
	Hash   string `json:"hash" binding:"required"`
	Type   string `json:"type"`
}

type PluginReinstallRequest struct {
	Node   string `json:"node"`
	Hash   string `json:"hash"`
	Module string `json:"module"`
}

type PluginRecheckRequest struct {
	Node   string `json:"node" binding:"required"`
	Hash   string `json:"hash" binding:"required"`
	Module string `json:"module" binding:"required"`
}

type PluginUninstallRequest struct {
	Node   string `json:"node" binding:"required"`
	Hash   string `json:"hash" binding:"required"`
	Module string `json:"module" binding:"required"`
}

type PluginKeyCheckRequest struct {
	Key string `json:"key" binding:"required"`
}

type PluginImportByDataRequest struct {
	JSON     string `json:"json" `
	Source   string `json:"source" `
	IsSystem bool   `json:"isSystem"`
	Key      string `json:"key"`
}

type PluginStatusRequest struct {
	ID     string `json:"id"`
	Status bool   `json:"status"`
}

type PluginRunRequest struct {
	Hash string `json:"hash" binding:"required"`
}

type PluginInfo struct {
	Help          string `json:"help"`
	Parameter     string `json:"parameter"`
	Module        string `json:"module"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Introduction  string `json:"introduction"`
	Hash          string `json:"hash"`
	ParameterList string `json:"parameterList"`
	Type          string `json:"type"`

	Source   string `json:"source,omitempty"`
	IsSystem bool   `json:"isSystem,omitempty"`
}

type NodePluginInfo struct {
	Name    string `json:"name"`
	Install string `json:"install"`
	Check   string `json:"check"`
	Hash    string `json:"hash"`
	Module  string `json:"module"`
}
