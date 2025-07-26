package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saxenaaman628/redis-voting-system/internal/middleware"
)

func RegisterRoutes(r *gin.Engine) {
	r.POST("/login", LoginHandler)

	pollGroup := r.Group("/polls").Use(middleware.JWTAuthMiddleware())
	auth := r.Group("/api").Use(middleware.JWTAuthMiddleware())

	auth.GET("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, map[string]string{"data": "working fine"})
	})

	pollGroup.POST("/createPoll", CreatePollHandler)
	pollGroup.GET("/getAllPoll", ListPollsHandler)
	pollGroup.GET("/getPoll/:id", GetPollByID)
	pollGroup.GET("close/:id", ClosePoll)
}
