// configuration-------------------------------------
// @file      : notification.go
// @time      : 2025/10/29 21:16
// -------------------------------------------

package configuration

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/models"
	"recon/internal/services/configuration"
)

var notificationService *configuration.Service

// GetNotificationData @Summary List notifications
// @Tags        configuration
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} response.SuccessResponse{data=object{list=[]map[string]interface{}}}
// @Router      /configuration/notification/data [get]
func GetNotificationData(c *gin.Context) {
	list, err := notificationService.GetNotificationList(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, gin.H{"list": list}, "api.success")
}

// AddNotificationData @Summary Create notification
// @Tags        configuration
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body map[string]interface{} true "Notification object"
// @Success     200 {object} response.SuccessResponse
// @Router      /configuration/notification/add [post]
func AddNotificationData(c *gin.Context) {
	var body models.Notification
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if err := notificationService.AddNotification(c, body); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

// UpdateNotificationData @Summary Update notification
// @Tags        configuration
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body map[string]interface{} true "Notification object with id"
// @Success     200 {object} response.SuccessResponse
// @Router      /configuration/notification/update [post]
func UpdateNotificationData(c *gin.Context) {
	var body models.UpdateNotification
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if body.ID == "" {
		response.BadRequest(c, "api.bad_request", nil)
		return
	}
	if err := notificationService.UpdateNotification(c, body); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

type deleteNotificationRequest struct {
	IDs []string `json:"ids" binding:"required" example:"[\"64f...\",\"64e...\"]"`
}

// DeleteNotification @Summary Delete notification
// @Tags        configuration
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body deleteNotificationRequest true "IDs to delete"
// @Success     200 {object} response.SuccessResponse
// @Router      /configuration/notification/delete [post]
func DeleteNotification(c *gin.Context) {
	var req deleteNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if err := notificationService.DeleteNotifications(c, req.IDs); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

// GetNotificationConfigData @Summary Get notification configuration
// @Tags        configuration
// @Produce     json
// @Security    ApiKeyAuth
// @Success     200 {object} response.SuccessResponse{data=object}
// @Router      /configuration/notification/config/data [get]
func GetNotificationConfigData(c *gin.Context) {
	data, err := notificationService.GetNotificationConfig(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, data, "api.success")
}

// UpdateNotificationConfigData @Summary Update notification configuration
// @Tags        configuration
// @Accept      json
// @Produce     json
// @Security    ApiKeyAuth
// @Param       body body map[string]interface{} true "Notification configuration"
// @Success     200 {object} response.SuccessResponse
// @Router      /configuration/notification/config/update [post]
func UpdateNotificationConfigData(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}
	if err := notificationService.UpdateNotificationConfig(c, body); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}
	response.Success(c, nil, "api.success")
}

func init() {
	notificationService = configuration.NewService()
}
