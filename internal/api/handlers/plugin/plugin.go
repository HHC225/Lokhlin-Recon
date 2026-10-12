package plugin

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/constants"
	"recon/internal/models"
	"recon/internal/services/plugin"
	"recon/internal/utils/random"
)

var pluginService plugin.Service

func init() {
	pluginService = plugin.NewService()
}

// @Summary List plugins
// @Description List plugins with pagination
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body plugin.PluginListRequest true "Request parameters"
// @Success 200 {object} response.Response{data=plugin.PluginListResponse}
// @Router /api/plugin [post]
func List(c *gin.Context) {
	var req models.PluginListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := pluginService.List(c, &req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}

// @Summary List plugins by module
// @Description List plugins in the specified module
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body map[string]string true "Request parameters"
// @Success 200 {object} response.Response{data=[]plugin.Plugin}
// @Router /api/plugin/module [post]
func ListByModule(c *gin.Context) {
	var req struct {
		Module string `json:"module" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := pluginService.ListByModule(c, req.Module)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, map[string]interface{}{
		"list": result,
	}, "api.success")
}

// @Summary Get plugin details
// @Description Get details of the specified plugin
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body plugin.PluginDetailRequest true "Request parameters"
// @Success 200 {object} response.Response{data=plugin.Plugin}
// @Router /api/plugin/detail [post]
func Detail(c *gin.Context) {
	var req models.PluginDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := pluginService.Detail(c, &req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}

// @Summary Save plugin
// @Description Create or update plugin
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body plugin.PluginSaveRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/save [post]
func Save(c *gin.Context) {
	var req models.PluginSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if req.Key == "" && !c.GetBool(constants.GatewayAuthenticatedContextKey) {
		response.BadRequest(c, "api.bad_request", nil)
		return
	}

	if err := pluginService.Save(c, &req); err != nil {
		response.InternalServerError(c, "api.plugin.key.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Delete plugin
// @Description Delete the specified plugin
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body plugin.PluginDeleteRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/delete [post]
func Delete(c *gin.Context) {
	var req models.PluginDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.Delete(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Get plugin logs
// @Description Get logs for the specified plugin
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body plugin.PluginLogRequest true "Request parameters"
// @Success 200 {object} response.Response{data=string}
// @Router /api/plugin/log [post]
func GetLogs(c *gin.Context) {
	var req models.PluginLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	result, err := pluginService.GetLogs(c, &req)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, gin.H{
		"data": result,
	}, "api.success")
}

// @Summary Clear plugin logs
// @Description Clear logs for the specified plugin
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body plugin.PluginLogRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/log/clean [post]
func CleanLogs(c *gin.Context) {
	var req models.PluginLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.CleanLogs(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Clear all plugin logs
// @Description Clear logs for all plugins
// @Tags Plugin management
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/plugin/log/clean/all [post]
func CleanAllLogs(c *gin.Context) {
	if err := pluginService.CleanAllLogs(c); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Import plugin
// @Description Import a plugin
// @Tags Plugin management
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Plugin file"
// @Param key query string false "Standalone plugin key; omitted for signed console requests"
// @Success 200 {object} response.Response
// @Router /api/plugin/import [post]
func Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	key := c.Query("key")
	if key == "" && !c.GetBool(constants.GatewayAuthenticatedContextKey) {
		response.BadRequest(c, "api.bad_request", nil)
		return
	}

	filePath := "uploads/" + random.GenerateString(5) + ".zip"
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	if err := pluginService.Import(c, filePath, key); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Reinstall plugin
// @Description Reinstall the specified plugin
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body plugin.PluginReinstallRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/reinstall [post]
func Reinstall(c *gin.Context) {
	var req models.PluginReinstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.Reinstall(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Recheck plugin
// @Description Recheck the specified plugin
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body plugin.PluginRecheckRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/recheck [post]
func Recheck(c *gin.Context) {
	var req models.PluginRecheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.Recheck(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Uninstall plugin
// @Description Uninstall the specified plugin
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body plugin.PluginUninstallRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/uninstall [post]
func Uninstall(c *gin.Context) {
	var req models.PluginUninstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.Uninstall(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Check plugin key
// @Description Check plugin key validity
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body plugin.PluginKeyCheckRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/key/check [post]
func CheckKey(c *gin.Context) {
	var req models.PluginKeyCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.CheckKey(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Search plugin catalog
// @Description List catalog plugins and mark installed plugins and available updates
// @Tags Plugin management
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /api/plugin/remote/search [post]
func SearchRemotePlugins(c *gin.Context) {
	result, err := pluginService.SearchRemotePlugins(c)
	if err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, result, "api.success")
}

// @Summary Import plugins through POST JSON
// @Description Import plugins through POST JSON containing serialized info.json, source and isSystem
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body models.PluginImportByDataRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/import/data [post]
func ImportByData(c *gin.Context) {
	var req models.PluginImportByDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.ImportByData(c, &req); err != nil {
		response.InternalServerError(c, "api.plugin.key.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Update plugin status
// @Description Enable or disable the specified plugin
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body models.PluginStatusRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/status [post]
func Status(c *gin.Context) {
	var req models.PluginStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.UpdateStatus(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}

// @Summary Run plugin once
// @Description Run the specified server plugin once
// @Tags Plugin management
// @Accept json
// @Produce json
// @Param request body models.PluginRunRequest true "Request parameters"
// @Success 200 {object} response.Response
// @Router /api/plugin/run [post]
func Run(c *gin.Context) {
	var req models.PluginRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	if err := pluginService.Run(c, &req); err != nil {
		response.InternalServerError(c, "api.error", err)
		return
	}

	response.Success(c, nil, "api.success")
}
