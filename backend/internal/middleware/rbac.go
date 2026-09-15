package middleware

import (
	"net/http"

	"cylawcase/internal/constants"

	"github.com/gin-gonic/gin"
)

// RequireRole 基于用户角色校验权限。
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetUserRole(c)
		for _, r := range roles {
			if role == r {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": constants.CodeForbidden, "message": constants.MsgForbidden, "data": nil})
	}
}
