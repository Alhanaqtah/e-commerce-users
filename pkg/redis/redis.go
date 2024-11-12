package redis

import (
	"context"
	"fmt"

	"github.com/Alhanaqtah/auth/internal/config"

	"github.com/redis/go-redis/v9"
)

func New(cfg *config.Redis) (*redis.Client, error) {
	const op = "redis.New"

	c := redis.NewClient(
		&redis.Options{
			Addr:     cfg.Address,
			Password: cfg.Password,
			DB:       cfg.DB,
		},
	)

	if err := c.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to ping, %w", op, err)
	}

	return c, nil
}
