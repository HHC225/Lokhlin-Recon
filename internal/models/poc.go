package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Poc struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	Content    string             `bson:"content" json:"content"`
	TemplateId string             `bson:"id" json:"Template ID"`
	Level      string             `bson:"level" json:"level"`
	Tags       []string           `bson:"tags" json:"tags"`
	Time       string             `bson:"time" json:"time"`
}

type PocListRequest struct {
	Search    string `json:"search"`
	PageIndex int    `json:"pageIndex"`
	PageSize  int    `json:"pageSize"`
	Filter    struct {
		Level []string `json:"level"`
	} `json:"filter"`
}

type PocListResponse struct {
	List  []Poc `json:"list"`
	Total int64 `json:"total"`
}

type PocDetailRequest struct {
	ID string `json:"id" binding:"required"`
}

type PocUpdateRequest struct {
	ID      string `json:"id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type PocAddRequest struct {
	Content string `json:"content" binding:"required"`
}

type PocDeleteRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

type TemplateInfo struct {
	Name        string                 `yaml:"name"`
	Author      string                 `yaml:"author"`
	Severity    string                 `yaml:"severity,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Reference   interface{}            `yaml:"reference,omitempty"`
	Remediation string                 `yaml:"remediation,omitempty"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty"`
	Tags        string                 `yaml:"tags,omitempty"`
}

type PocTemplate struct {
	ID string `yaml:"id" json:"id" jsonschema:"title=id of the template,description=The Unique ID for the template,required,example=cve-2021-19520,pattern=^([a-zA-Z0-9]+[-_])*[a-zA-Z0-9]+$"`

	Info TemplateInfo `yaml:"info" json:"info" jsonschema:"title=info for the template,description=Info contains metadata for the template,required,type=object"`
}

type PocImportRequest struct {
}

type PocImportResponse struct {
	SuccessNum int    `json:"success_num"`
	ErrorNum   int    `json:"error_num"`
	RepeatNum  int    `json:"repeat_num"`
	Message    string `json:"message"`
}
