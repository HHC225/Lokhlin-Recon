// fingerprint-------------------------------------
// @file      : fingerprint.go
// @time      : 2025/10/28 20:30
// -------------------------------------------

package fingerprint

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"recon/internal/api/response"
	"recon/internal/models"
	service "recon/internal/services/fingerprint"
	"recon/internal/utils/random"
)

var fingerprintService service.Service

type listRequest struct {
	Search    string `json:"search" binding:"omitempty" example:"nginx"`
	PageIndex int    `json:"pageIndex" binding:"required,min=1" example:"1"`
	PageSize  int    `json:"pageSize" binding:"required,min=1" example:"10"`
}

type updateRequest struct {
	ID      string `json:"id" binding:"required" example:"6642b0..."`
	Content string `json:"content" binding:"required" example:"name: Nginx\ncategory: web"`
	State   bool   `json:"state"`
}

type addRequest struct {
	Content string `json:"content" binding:"required" example:"name: Nginx\ncategory: web"`
	State   bool   `json:"state" example:"1"`
}

type deleteRequest struct {
	IDs []string `json:"ids" binding:"required,min=1,dive,required"`
}

type batchAddItem struct {
	ID         string `json:"id" example:"SAS_xxxx"`
	Name       string `json:"name" example:"druid-server"`
	Content    string `json:"content" binding:"required" example:"fingerprint:\n  name: \"druid-server\"\n  category: \"Service\""`
	UpdateTime string `json:"updateTime" example:"2026-01-19 23:50"`
}

type batchAddRequest struct {
	Items []batchAddItem `json:"items" binding:"required,min=1,dive"`
}

type Fingerprint struct {
	Name           string   `yaml:"name"`
	Id             string   `yaml:"id"`
	Tags           []string `yaml:"tags"`
	Category       string   `yaml:"category"`
	ParentCategory string   `yaml:"parent_category"`
	Company        string   `yaml:"company"`
	Rules          []Rule   `yaml:"rules"`
}

type FingerprintYaml struct {
	Fingerprint Fingerprint `yaml:"fingerprint"`
}

type Rule struct {
	Logic      string      `yaml:"logic"`
	Conditions []Condition `yaml:"conditions"`
}

type Condition struct {
	Location    string `yaml:"location,omitempty"`   // body, header, title, request
	MatchType   string `yaml:"match_type,omitempty"` // regex, contains, extract, active
	Pattern     string `yaml:"pattern,omitempty"`
	Group       int    `yaml:"group,omitempty"`
	SaveAs      string `yaml:"save_as,omitempty"`
	Path        string `yaml:"path,omitempty"`
	DynamicPath string `yaml:"dynamic_path,omitempty"`
	Method      string `yaml:"method,omitempty"`

	Logic      string      `yaml:"logic,omitempty"`
	Conditions []Condition `yaml:"conditions,omitempty"`
}

// Data @Summary      Query fingerprint rules
// @Description  Search by partial name with pagination
// @Tags         fingerprint
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      listRequest  true  "Query parameters"
// @Success      200      {object}  response.SuccessResponse{data=object{list=[]models.FingerprintRule,total=int}}
// @Failure      400      {object}  response.BadRequestResponse
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /fingerprint/data [post]
func Data(c *gin.Context) {
	var req listRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	list, total, err := fingerprintService.List(c, req.Search, req.PageIndex, req.PageSize)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"list": list, "total": total}, "api.success")
}

// Update @Summary      Update fingerprint rule
// @Tags         fingerprint
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      updateRequest  true  "Update data"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.BadRequestResponse
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /fingerprint/update [post]
func Update(c *gin.Context) {
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	var fp FingerprintYaml
	if err := yaml.Unmarshal([]byte(req.Content), &fp); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	data := models.FingerprintRule{
		Name:           fp.Fingerprint.Name,
		Rule:           req.Content,
		Category:       fp.Fingerprint.Category,
		ParentCategory: fp.Fingerprint.ParentCategory,
		State:          req.State,
		Company:        fp.Fingerprint.Company,
	}

	if err := fingerprintService.Update(c, req.ID, data); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

// Add @Summary      Create fingerprint rule
// @Tags         fingerprint
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      addRequest  true  "Create data"
// @Success      200      {object}  response.SuccessResponse{data=object{id=string}}
// @Failure      400      {object}  response.BadRequestResponse
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /fingerprint/add [post]
func Add(c *gin.Context) {
	var req addRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	var fp FingerprintYaml
	if err := yaml.Unmarshal([]byte(req.Content), &fp); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	FingerId := "SAA_" + random.GenerateRandomString(12)

	data := models.FingerprintRule{
		Name:           fp.Fingerprint.Name,
		Rule:           req.Content,
		FingerprintId:  FingerId,
		Category:       fp.Fingerprint.Category,
		ParentCategory: fp.Fingerprint.ParentCategory,
		State:          req.State,
		Company:        fp.Fingerprint.Company,
	}

	id, err := fingerprintService.Add(c, data)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"id": id}, "api.success")
}

// Delete @Summary      Delete fingerprint rules in bulk
// @Tags         fingerprint
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      deleteRequest  true  "IDs to delete"
// @Success      200      {object}  response.SuccessResponse
// @Failure      400      {object}  response.BadRequestResponse
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /fingerprint/delete [post]
func Delete(c *gin.Context) {
	var req deleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	_, err := fingerprintService.Delete(c, req.IDs)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

// GetVersion @Summary      Get fingerprint version
// @Tags         fingerprint
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200      {object}  response.SuccessResponse{data=object{version=string}}
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /fingerprint/version [get]
func GetVersion(c *gin.Context) {
	version, err := fingerprintService.GetVersion(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"version": version}, "api.success")
}

// BatchAdd @Summary      Create fingerprint rules in bulk
// @Tags         fingerprint
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      batchAddRequest  true  "Bulk create data"
// @Success      200      {object}  response.SuccessResponse{data=object{success=int,failed=int,results=[]object{id=int,success=bool,error=string}}}
// @Failure      400      {object}  response.BadRequestResponse
// @Failure      500      {object}  response.InternalServerErrorResponse
// @Router       /fingerprint/batch-add [post]
func BatchAdd(c *gin.Context) {
	var req batchAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	items := make([]service.BatchAddItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = service.BatchAddItem{
			ID:         item.ID,
			Name:       item.Name,
			Content:    item.Content,
			UpdateTime: item.UpdateTime,
		}
	}

	results, successCount, failedCount, err := fingerprintService.BatchAdd(c, items)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, gin.H{
		"success": successCount,
		"failed":  failedCount,
		"results": results,
	}, "api.success")
}

func init() {
	fingerprintService = service.NewService()
}
