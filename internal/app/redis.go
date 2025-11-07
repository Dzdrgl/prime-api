package app

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis(addr, pass string, db int) (*redis.Client, error) {
	rc := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DB:           db,
		MinIdleConns: 5,
		PoolSize:     50,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rc.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return rc, nil
}
