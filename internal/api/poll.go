package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/saxenaaman628/redis-voting-system/internal/models"
	"github.com/saxenaaman628/redis-voting-system/internal/redis"
)

var createPollInput struct {
	Question  string   `json:"question" binding:"required"`
	Options   []string `json:"options" binding:"required,min=2"`
	ExpiresIn int      `json:"expires_in"` // in minutes
}

func CreatePollHandler(c *gin.Context) {

	if err := c.ShouldBindJSON(&createPollInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	newPoll := models.Poll{
		ID:        uuid.New().String(),
		Question:  createPollInput.Question,
		Options:   createPollInput.Options,
		CreatedBy: userID.(string),
		UpdatedBy: userID.(string),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(createPollInput.ExpiresIn) * time.Minute),
		IsClosed:  false,
	}

	// Save to Redis (or DB in future phase)
	key := "poll:" + newPoll.ID

	err := redis.Rdb.HSet(c, key, map[string]interface{}{
		"id":         newPoll.ID,
		"question":   newPoll.Question,
		"created_by": newPoll.CreatedBy,
		"updated_by": newPoll.UpdatedBy,
		"created_at": newPoll.CreatedAt.Format(time.RFC3339),
		"expires_at": newPoll.ExpiresAt.Format(time.RFC3339),
		"is_closed":  false,
	}).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create poll"})
		return
	}

	// Store options separately
	for idx, opt := range newPoll.Options {
		redis.Rdb.HSet(c, key+":options", idx, opt)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poll created", "poll_id": newPoll.ID})
}

func ListPollsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	pattern := "poll:*"
	keys, err := redis.Rdb.Keys(ctx, pattern).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch polls"})
		return
	}
	polls := []models.Poll{}
	for _, key := range keys {
		if strings.Contains(key, ":options") {
			continue // skip options keys
		}

		data, err := redis.Rdb.HGetAll(ctx, key).Result()
		if err != nil || len(data) == 0 {
			continue
		}
		log.Println("--DATA--", data)
		createdAt, _ := time.Parse(time.RFC3339, data["created_at"])
		expiresAt, _ := time.Parse(time.RFC3339, data["expires_at"])
		isClosed := data["is_closed"] == "true"

		//check timing
		if isClosed || expiresAt.Before(time.Now()) {
			continue // skip expired or closed
		}

		optionsMap, _ := redis.Rdb.HGetAll(ctx, key+":options").Result()
		var options []string
		for i := 0; i < len(optionsMap); i++ {
			options = append(options, optionsMap[strconv.Itoa(i)])
		}

		poll := models.Poll{
			ID:        data["id"],
			Question:  data["question"],
			Options:   options,
			CreatedBy: data["created_by"],
			UpdatedBy: data["updated_by"],
			CreatedAt: createdAt,
			ExpiresAt: expiresAt,
			IsClosed:  false,
		}

		polls = append(polls, poll)
	}
	c.JSON(http.StatusOK, map[string]interface{}{"polls": polls})
}

func GetPollByID(c *gin.Context) {
	pollID := c.Param("id")
	if pollID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please pass id"})
		return
	}

	pollKey := fmt.Sprintf("poll:%s", pollID)
	pollOptionKey := fmt.Sprintf("poll:%s:options", pollID)

	type redisResult struct {
		data map[string]string
		err  error
	}

	// Channels to receive Redis results
	pollCh := make(chan redisResult)
	optionCh := make(chan redisResult)

	// Create a context with timeout for Redis operations
	ctx, cancel := context.WithTimeout(redis.Ctx, 3*time.Second)
	defer cancel()

	// Goroutine for poll data
	go func() {
		res, err := redis.Rdb.HGetAll(ctx, pollKey).Result()
		pollCh <- redisResult{data: res, err: err}
	}()

	// Goroutine for poll options
	go func() {
		res, err := redis.Rdb.HGetAll(ctx, pollOptionKey).Result()
		optionCh <- redisResult{data: res, err: err}
	}()

	// Collect results
	pollRes := <-pollCh
	optionRes := <-optionCh

	// Handle Redis errors
	if pollRes.err != nil || optionRes.err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch poll or options"})
		return
	}

	if len(pollRes.data) == 0 || len(optionRes.data) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll or options not found"})
		return
	}

	var data models.Poll
	mapstructure.Decode(pollRes.data, &data)

	// Parse time fields
	if t, err := time.Parse(time.RFC3339, pollRes.data["created_at"]); err == nil {
		data.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339, pollRes.data["expires_at"]); err == nil {
		data.ExpiresAt = t
	}
	//get is_closed flag
	data.IsClosed = true
	if pollRes.data["is_closed"] == "false" {
		data.IsClosed = false
	}
	// Convert map values to slice
	options := make([]string, 0, len(optionRes.data))
	for _, v := range optionRes.data {
		options = append(options, v)
	}
	data.Options = options

	c.JSON(http.StatusOK, gin.H{"data": data})
}

func ClosePoll(c *gin.Context) {
	user, exists := c.Get("userID")
	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can close polls"})
		return
	}
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	pollID := c.Param("id")
	pollKey := "poll:" + pollID

	existsCmd := redis.Rdb.Exists(context.Background(), pollKey)
	if existsCmd.Val() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
		return
	}
	err := redis.Rdb.HSet(context.Background(), pollKey, map[string]interface{}{"updated_by": user, "is_closed": "true"}).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close poll"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poll closed successfully"})
}
