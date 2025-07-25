package main

import (
	"github.com/gin-gonic/gin"
	"github.com/saxenaaman628/redis-voting-system/config"
	"github.com/saxenaaman628/redis-voting-system/internal/redis"
)

func main() {
	config.LoadEnv()
	redis.InitRedis()

	r := gin.Default()
	// api.SetupRoutes(r)

	port := config.GetEnv("PORT", "8080")
	r.Run(":" + port)
}
