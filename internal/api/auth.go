package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saxenaaman628/redis-voting-system/internal/models"
	"github.com/saxenaaman628/redis-voting-system/internal/utils"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	for _, u := range models.DummyUsers { //update later
		if u.Username == req.Username && u.Password == req.Password {
			token, err := utils.GenerateJWTToken(u.ID, u.Username, u.Role)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"token": token})
			return
		}
	}
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
}
