package router

import (
	"cylawcase/internal/constants"
	"cylawcase/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerUserRoutes 用户路由。
func (r *Router) registerUserRoutes(g *gin.RouterGroup) {
	users := g.Group("/users")
	users.Use(middleware.AuthRequired(r.cfg))
	users.GET("/me", r.user.Me)
	users.PUT("/me", r.user.UpdateProfile)
	users.GET("/lawyers", r.user.ListLawyers)
	users.GET("", middleware.RequireRole(constants.RoleAdmin), r.user.List)
}
