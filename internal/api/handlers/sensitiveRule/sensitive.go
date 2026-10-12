// sensitiveRule-------------------------------------
// @file      : sensitive.go
// @time      : 2025/10/28 22:08
// -------------------------------------------

package sensitiveRule

import (
	"sync"

	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	svc "recon/internal/services/sensitive_rule"
)

var (
	sensitiveService     svc.Service
	sensitiveServiceOnce sync.Once
)

type listRequest struct {
	Search    string `json:"search" binding:"omitempty"`
	PageIndex int    `json:"pageIndex" binding:"required,min=1"`
	PageSize  int    `json:"pageSize" binding:"required,min=1"`
}

type updateRequest struct {
	ID      string `json:"id" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Regular string `json:"regular" binding:"required"`
	Color   string `json:"color" binding:"required"`
	State   *bool  `json:"state" binding:"required"`
}

type addRequest struct {
	Name    string `json:"name" binding:"required"`
	Regular string `json:"regular" binding:"required"`
	Color   string `json:"color" binding:"required"`
	State   *bool  `json:"state" binding:"required"`
}

type idsStateRequest struct {
	IDs   []string `json:"ids" `
	State bool     `json:"state"`
}

type idsRequest struct {
	IDs []string `json:"ids" binding:"required,min=1,dive,required"`
}

func Data(c *gin.Context) {
	var req listRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	list, total, err := getSensitiveService().RuleList(c, req.Search, req.PageIndex, req.PageSize)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"list": list, "total": total}, "api.success")
}

func Update(c *gin.Context) {
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if err := getSensitiveService().RuleUpdate(c, req.ID, req.Name, req.Regular, req.Color, *req.State); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

func Add(c *gin.Context) {
	var req addRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if err := getSensitiveService().RuleAdd(c, req.Name, req.Regular, req.Color, *req.State); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

func UpdateState(c *gin.Context) {
	var req idsStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if _, err := getSensitiveService().RuleUpdateState(c, req.IDs, req.State); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

func Delete(c *gin.Context) {
	var req idsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if _, err := getSensitiveService().RuleDelete(c, req.IDs); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

func getSensitiveService() svc.Service {
	sensitiveServiceOnce.Do(func() {
		sensitiveService = svc.NewService()
	})
	return sensitiveService
}
