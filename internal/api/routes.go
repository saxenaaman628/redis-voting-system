package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saxenaaman628/redis-voting-system/internal/controller"
	"github.com/saxenaaman628/redis-voting-system/internal/middleware"
)

func RegisterRoutes(r *gin.Engine) {
	r.POST("/login", controller.LoginHandler)

	authGroup := r.Group("/api").Use(middleware.JWTAuthMiddleware())
	pollGroup := r.Group("/polls").Use(middleware.JWTAuthMiddleware())
	voteGroup := r.Group("/votes").Use(middleware.JWTAuthMiddleware())

	authGroup.GET("/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, map[string]string{"data": "working fine"})
	})

	pollGroup.POST("/createPoll", controller.CreatePollHandler)
	pollGroup.GET("/getAllPoll", controller.ListPollsHandler)
	pollGroup.GET("/getPoll/:id", controller.GetPollByID)
	pollGroup.GET("close/:id", controller.ClosePoll)
	pollGroup.GET("delete/:id", controller.DeletePollHandler)
	pollGroup.GET("/search", controller.SearchPollsHandler)
	pollGroup.GET("get_polls_by_vote", controller.GetPollsWithVotes)

	voteGroup.POST("/:poll_id/vote", controller.VoteHandler)
}
