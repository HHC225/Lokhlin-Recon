package url

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/models"
	"recon/internal/services/assets/url"
)

var urlService url.Service

func init() {
	urlService = url.NewService()
}

// GetURLs godoc
// @Summary List URLs
// @Description List URLs，Supports pagination, sorting and filtering
// @Tags URL
// @Accept json
// @Produce json
// @Param query body models.SearchRequest true "Query parameters"
// @Success 200 {object} response.SuccessResponse{data=object{list=[]models.URL}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/url [post]
func GetURLs(c *gin.Context) {
	var query models.SearchRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	urls, err := urlService.GetURLs(c, query)
	if err != nil {
		response.InternalServerError(c, "api.url.list.failed", err)
		return
	}

	response.Success(c, gin.H{
		"list": urls,
	}, "api.url.list.success")
}
