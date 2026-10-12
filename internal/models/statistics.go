// models-------------------------------------
// @file      : STATISTICS.go
// @time      : 2025/5/9 21:16
// -------------------------------------------

package models

type StatisticsData struct {
	AssetCount         int64 `json:"assetCount"`
	SubdomainCount     int64 `json:"subdomainCount"`
	SensitiveCount     int64 `json:"sensitiveCount"`
	UrlCount           int64 `json:"urlCount"`
	VulnerabilityCount int64 `json:"vulnerabilityCount"`
}

type PortStatisticsRequest struct {
	Filter map[string]interface{} `json:"filter" binding:"required" example:"{\"type\":\"http\"}"`
}

type PortStatisticsResponse struct {
	Port []PortItem `json:"Port"`
}

type PortItem struct {
	Value  string `bson:"_id" json:"value" example:"80"`
	Number int64  `bson:"number" json:"number" example:"100"`
}

type TitleStatisticsRequest struct {
	Filter map[string]interface{} `json:"filter" binding:"required" example:"{\"type\":\"http\"}"`
}

type TitleStatisticsResponse struct {
	Title []TitleItem `json:"Title"`
}

type TitleItem struct {
	Value  string `json:"value" example:"Example Title"`
	Number int64  `json:"number" example:"50"`
}

type TypeStatisticsRequest struct {
	Filter map[string]interface{} `json:"filter" binding:"required" example:"{\"type\":\"http\"}"`
}

type TypeStatisticsResponse struct {
	Service []TypeItem `json:"Service"`
}

type TypeItem struct {
	Value  string `bson:"_id" json:"value" example:"nginx"`
	Number int64  `bson:"number" json:"number" example:"200"`
}

type IconStatisticsRequest struct {
	Filter   map[string]interface{} `json:"filter" binding:"required" example:"{\"type\":\"http\"}"`
	Page     int                    `json:"page" binding:"required" example:"1"`
	PageSize int                    `json:"page_size" binding:"required" example:"10"`
}

type IconStatisticsResponse struct {
	Icon []IconItem `json:"Icon"`
}

type IconItem struct {
	Value    string `bson:"iconcontent" json:"value" example:"base64_icon_content"`
	Number   int64  `bson:"number" json:"number" example:"30"`
	IconHash string `bson:"_id" json:"icon_hash" example:"abc123"`
}

type AppStatisticsRequest struct {
	Filter map[string]interface{} `json:"filter" binding:"required" example:"{\"type\":\"http\"}"`
}

type AppStatisticsResponse struct {
	Product []AppItem `json:"Product"`
}

type AppItem struct {
	Value  string `bson:"_id" json:"value" example:"WordPress"`
	Number int64  `bson:"number" json:"number" example:"150"`
}

var APP = map[string]string{}
