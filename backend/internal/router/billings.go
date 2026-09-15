package router

import (
	"cylawcase/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerBillingRoutes 账单路由。
func (r *Router) registerBillingRoutes(g *gin.RouterGroup) {
	billings := g.Group("/billings")
	billings.Use(middleware.AuthRequired(r.cfg))
	billings.GET("", r.billing.List)
	billings.GET("/summary", r.billing.Summary)
	billings.GET("/by-case/:id", r.billing.ListByCase)
	billings.POST("", r.billing.Create)
	billings.POST("/:id/paid", r.billing.MarkPaid)
	billings.POST("/:id/invoiced", r.billing.MarkInvoiced)
	billings.POST("/:id/void", r.billing.Void)
}
