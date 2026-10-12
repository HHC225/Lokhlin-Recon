package crawler

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/response"
	"recon/internal/models"
	"recon/internal/services/assets/crawler"
)

var crawlerService crawler.Service

func init() {
	crawlerService = crawler.NewService()
}

// GetCrawlers godoc
// @Summary List crawler tasks
// @Description List crawler tasks，Supports pagination, sorting and filtering
// @Tags Crawler
// @Accept json
// @Produce json
// @Param query body models.SearchRequest true "Query parameters"
// @Success 200 {object} response.SuccessResponse{data=object{list=[]models.CrawlerResult}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/assets/crawler [post]
func GetCrawlers(c *gin.Context) {
	var query models.SearchRequest
	if err := c.ShouldBindJSON(&query); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	tasks, err := crawlerService.GetCrawlers(c, query)
	if err != nil {
		response.InternalServerError(c, "api.crawler.tasks.list.failed", err)
		return
	}

	response.Success(c, gin.H{
		"list": tasks,
	}, "api.crawler.list.success")
}
