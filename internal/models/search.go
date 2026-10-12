// models-------------------------------------
// @file      : search.go
// @time      : 2025/5/7 19:27
// -------------------------------------------

package models

type SearchRequest struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`

	Index            string `json:"index"`
	SearchExpression string `json:"search"`

	Filter map[string][]interface{} `json:"filter"`

	FuzzyQuery map[string]string `json:"-"`
	Sort       map[string]string `json:"sort"`
	Sid        string            `json:"sid"`
}
