package controller

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"

	"github.com/saxenaaman628/redis-voting-system/internal/models"
	"github.com/saxenaaman628/redis-voting-system/internal/redis"
	redishandler "github.com/saxenaaman628/redis-voting-system/internal/redisHandler"
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
		createdAt, _ := time.Parse(time.RFC3339, data["created_at"])
		expiresAt, _ := time.Parse(time.RFC3339, data["expires_at"])
		isClosed := data["is_closed"] == "true"

		//check timing
		// if isClosed || expiresAt.Before(time.Now()) {
		// 	continue // skip expired or closed
		// }

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
			IsClosed:  isClosed,
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

func DeletePollHandler(c *gin.Context) {
	userID, validUser := c.Get("userID")
	isAdmin, _ := c.Get("role")
	// log.Println("userID-----", userID, "isAdmin----", isAdmin)
	if !validUser {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	pollID := c.Param("id")

	key := "poll:" + pollID
	optionsKey := fmt.Sprintf("%s:options", key)

	// Check if poll exists
	exists, err := redis.Rdb.Exists(context.Background(), key).Result()
	if err != nil || exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
		return
	}

	// Get poll details to verify creator
	pollData, err := redis.Rdb.HGetAll(context.Background(), key).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch poll"})
		return
	}

	createdBy := pollData["created_by"]
	if userID != createdBy && isAdmin != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to delete this poll"})
		return
	}

	// Delete poll
	_, err = redis.Rdb.Del(context.Background(), key, optionsKey).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete poll"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poll deleted successfully"})
}

func SearchPollsHandler(c *gin.Context) {
	createdBy := c.Query("created_by")
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")
	isClosedStr := c.Query("is_closed")

	var fromDate, toDate time.Time
	var err error
	if fromDateStr != "" {
		fromDate, err = time.Parse(time.RFC3339, fromDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_date format"})
			return
		}
	}
	if toDateStr != "" {
		toDate, err = time.Parse(time.RFC3339, toDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_date format"})
			return
		}
	}

	var isClosedFilter *bool
	if isClosedStr != "" {
		isClosed, err := strconv.ParseBool(isClosedStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid is_closed value"})
			return
		}
		isClosedFilter = &isClosed
	}

	// Fetch all polls from Redis

	allPolls, err := redishandler.GetAllPolls(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch polls"})
		return
	}

	// Apply filters
	filtered := make([]models.Poll, 0)
	for _, poll := range allPolls {
		if createdBy != "" && poll.CreatedBy != createdBy {
			continue
		}
		if !fromDate.IsZero() && poll.CreatedAt.Before(fromDate) {
			continue
		}
		if !toDate.IsZero() && poll.CreatedAt.After(toDate) {
			continue
		}
		if isClosedFilter != nil && poll.IsClosed != *isClosedFilter {
			continue
		}
		filtered = append(filtered, poll)
	}

	c.JSON(http.StatusOK, filtered)
}

func GetPollsWithVotes(c *gin.Context) {
	pattern := "poll:*"
	keys, err := redis.Rdb.Keys(c, pattern).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch poll keys"})
		return
	}
	polls := []models.Poll{}

	for _, key := range keys {
		// Ignore non-poll hashes like "poll:<id>:votes"
		if strings.Contains(key, ":voters") || strings.Contains(key, ":options") || strings.Contains(key, ":votes") {
			continue
		}
		// Get poll hash
		pollData, err := redis.Rdb.HGetAll(c, key).Result()
		if err != nil || len(pollData) == 0 {
			continue
		}

		// Convert to struct
		var poll models.Poll
		mapstructure.Decode(pollData, &poll)

		// Fetch vote counts
		voteKey := "poll:" + poll.ID + ":votes"
		voteData, err := redis.Rdb.HGetAll(c, voteKey).Result()
		if err != nil {
			voteData = map[string]string{}
		}

		votes := make(map[string]int64)
		for option, count := range voteData {
			intVal, _ := strconv.ParseInt(count, 10, 64)
			votes[option] = intVal
		}

		polls = append(polls, models.Poll{
			ID:       poll.ID,
			Question: poll.Question,
			// Options:   poll.Options,
			Votes: votes,
		})
	}

	c.JSON(http.StatusOK, polls)
}
