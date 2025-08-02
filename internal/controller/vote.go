package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"github.com/saxenaaman628/redis-voting-system/internal/models"
	"github.com/saxenaaman628/redis-voting-system/internal/redis"
)

// VotePayload is the expected vote request
type VotePayload struct {
	Option string `json:"option" binding:"required"`
}

func VoteHandler(c *gin.Context) {
	userID := c.GetString("userID")
	pollID := c.Param("poll_id")

	// Parse request body
	var payload VotePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vote payload"})
		return
	}

	// Get poll data
	key := "poll:" + pollID
	data, err := redis.Rdb.HGetAll(c, key).Result()
	if err != nil || len(data) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
		return
	}
	//get Options data
	optionKey := "poll:" + pollID + ":options"

	optiondata, err := redis.Rdb.HGetAll(c, optionKey).Result()
	if err != nil || len(data) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
		return
	}

	var poll models.Poll
	mapstructure.Decode(data, &poll)
	// Parse time fields
	if t, err := time.Parse(time.RFC3339, data["created_at"]); err == nil {
		poll.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339, data["expires_at"]); err == nil {
		poll.ExpiresAt = t
	}
	//get is_closed flag
	poll.IsClosed = true
	if data["is_closed"] == "0" {
		poll.IsClosed = false
	}
	// Convert map values to slice
	options := make([]string, 0, len(optiondata))
	for _, v := range optiondata {
		options = append(options, v)
	}
	poll.Options = options
	// Check if poll is closed or expired
	if poll.IsClosed || time.Now().After(poll.ExpiresAt) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Poll is closed or expired"})
		return
	}
	// Check if option is valid
	isValid := false
	for _, opt := range poll.Options {
		if strings.EqualFold(opt, payload.Option) {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid option selected"})
		return
	}
	// Check if user already voted
	voterKey := "poll:" + pollID + ":voters"
	if exists, _ := redis.Rdb.SIsMember(c, voterKey, userID).Result(); exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already voted"})
		return
	}

	pipe := redis.Rdb.TxPipeline()

	// Add user to voter set
	pipe.SAdd(c, "poll:"+pollID+":voters", userID)

	// Increment vote count for selected option
	pipe.HIncrBy(c, "poll:"+pollID+":votes", payload.Option, 1)

	// Log user's vote (optional)
	pipe.HSet(c, "user:"+userID+":votes", pollID, payload.Option)

	// Execute all commands atomically
	_, err = pipe.Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record vote"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote recorded successfully"})
}
