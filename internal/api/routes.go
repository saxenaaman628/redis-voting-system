package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saxenaaman628/redis-voting-system/internal/middleware"
)

func RegisterRoutes(r *gin.Engine) {
	r.POST("/login", LoginHandler)

	auth := r.Group("/api")
	auth.Use(middleware.JWTAuthMiddleware())
	{
		// Poll and vote routes will be added here
		auth.GET("/test", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, map[string]string{"data": "working fine"})
		})
	}
}
