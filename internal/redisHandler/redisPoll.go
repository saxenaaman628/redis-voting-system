package redishandler

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/saxenaaman628/redis-voting-system/internal/models"
	"github.com/saxenaaman628/redis-voting-system/internal/redis"
)

func GetAllPolls(ctx context.Context) ([]models.Poll, error) {
	var cursor uint64
	var keys []string
	var err error
	allPolls := make([]models.Poll, 0)
	r := redis.Rdb
	for {
		keys, cursor, err = r.Scan(ctx, cursor, "poll:*", 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			data, err := r.HGetAll(ctx, key).Result()
			if err != nil || len(data) == 0 {
				continue
			}

			var poll models.Poll
			err = mapstructure.Decode(data, &poll)
			if err != nil {
				continue
			}
			allPolls = append(allPolls, poll)
		}

		if cursor == 0 {
			break
		}
	}

	return allPolls, nil
}
