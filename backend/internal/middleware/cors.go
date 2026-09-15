package middleware

import (
	"cylawcase/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 跨域配置，来源从环境变量 APP_CORS_ORIGINS 读取，生产默认不允许 *。
func CORS(cfg *config.Config) gin.HandlerFunc {
	origins := cfg.CORSOrigins
	if len(origins) == 0 {
		origins = []string{"http://localhost:28031"}
	}
	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-Id"},
		ExposeHeaders:    []string{"X-Request-Id"},
		AllowCredentials: false,
	})
}
