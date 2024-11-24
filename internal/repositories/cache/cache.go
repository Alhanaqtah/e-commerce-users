package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const valBlacklisted = "blacklisted"

type Cache struct {
	rc *redis.Client
}

func New(rc *redis.Client) *Cache {
	return &Cache{
		rc: rc,
	}
}

func (c *Cache) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	const op = "repositories.cache.InBlacklist"

	exists, err := c.rc.Exists(ctx, token).Result()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists > 0, nil
}

func (c *Cache) AddToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	const op = "repositories.cache.AddToBlacklist"

	if _, err := c.rc.Set(ctx, token, valBlacklisted, ttl).Result(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
