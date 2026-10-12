// assets-------------------------------------
// @file      : asset.go
// @time      : 2025/5/8 19:51
// -------------------------------------------

package assets

import (
	"github.com/gin-gonic/gin"
	"recon/internal/api/routes/common"
	"recon/internal/models"
)

func RegisterAssetsRoutes(api *gin.RouterGroup) {
	nodeRoutes := models.RouteGroup{
		Path: "/assets",
		Groups: []models.RouteGroup{
			registerAssetRoutes(),
			registerStatisticsRoutes(),
			registerRootDomainRoutes(),
			registerCommonRoutes(),
			registerSubDomainRoutes(),
			registerAppRoutes(),
			registerMpRoutes(),
			registerUrlRoutes(),
			registerIPRoutes(),
			registerCrawlerRoutes(),
			registerSensitivveRoutes(),
			registerDirscanRoutes(),
			registerVulnerabilityRoutes(),
			registerPageMonitoringRoutes(),
		},
	}

	common.RegisterRouteGroup(api, nodeRoutes)
}
