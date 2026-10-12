package apikey

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/models"
	"recon/internal/services/apikey"
)

var apiKeyService apikey.Service

func init() {
	apiKeyService = apikey.NewService()
}

func List(c *gin.Context) {
	keys, err := apiKeyService.List(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"list": keys}, "")
}

func Create(c *gin.Context) {
	var req models.CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	username, _ := c.Get("username")
	createdBy, _ := username.(string)

	result, err := apiKeyService.Create(c.Request.Context(), req.Name, createdBy)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, result, "api.success")
}

func Delete(c *gin.Context) {
	var req models.DeleteApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := apiKeyService.Delete(c.Request.Context(), req.ID); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}
