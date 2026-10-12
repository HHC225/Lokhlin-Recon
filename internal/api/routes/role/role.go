package role

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoleRoutes(r *gin.Engine) {

	//api := r.Group("/api")
	//{

	//	roleGroup := api.Group("/roles")
	//	{

	//		roleGroup.Use(middleware.AuthMiddleware())
	//		roleGroup.Use(middleware.PermissionMiddleware(constants.PermissionRoleRead))
	//		{
	//			roleGroup.GET("", role.List)
	//			roleGroup.GET("/:id", role.Get)
	//		}
	//

	//		roleGroup.Use(middleware.PermissionMiddleware(constants.PermissionRoleCreate))
	//		{
	//			roleGroup.POST("", role.Create)
	//		}
	//

	//		roleGroup.Use(middleware.PermissionMiddleware(constants.PermissionRoleUpdate))
	//		{
	//			roleGroup.PUT("/:id", role.Update)
	//		}
	//

	//		roleGroup.Use(middleware.PermissionMiddleware(constants.PermissionRoleDelete))
	//		{
	//			roleGroup.DELETE("/:id", role.Delete)
	//		}
	//	}
	//}
}
