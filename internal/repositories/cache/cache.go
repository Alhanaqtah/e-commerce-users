package cache

import (
	"context"
	"fmt"
	"time"

	"e-commerce-users/internal/repositories"

	"github.com/redis/go-redis/v9"
)

const valBlacklisted = "blacklisted"

type Cache struct {
	rc     *redis.Client
	prefix string
}

func New(rc *redis.Client, prefix string) *Cache {
	return &Cache{
		rc:     rc,
		prefix: fmt.Sprintf("%s_", prefix),
	}
}

func (c *Cache) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	const op = "repositories.cache.InBlacklist"

	exists, err := c.rc.Exists(ctx, c.prefix+token).Result()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists > 0, nil
}

func (c *Cache) AddToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	const op = "repositories.cache.AddToBlacklist"

	if _, err := c.rc.Set(ctx, c.prefix+token, valBlacklisted, ttl).Result(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Cache) SetConfirmationCode(ctx context.Context, email, code string, ttl time.Duration) error {
	const op = "repositories.cache.SetConfirmationCode"

	if _, err := c.rc.Set(ctx, c.prefix+email, code, ttl).Result(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Cache) GetConfirmationCode(ctx context.Context, email string) (string, error) {
	const op = "repositories.cache.SetConfirmationCode"

	code, err := c.rc.Get(ctx, c.prefix+email).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("%s: %w", op, repositories.ErrNotFound)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return code, nil
}

func (c *Cache) RemoveConfirmationCode(ctx context.Context, email string) error {
	const op = "repositories.cache.RemoveConfirmationCode"

	_, err := c.rc.Del(ctx, c.prefix+email).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("%s: %w", op, repositories.ErrNotFound)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
