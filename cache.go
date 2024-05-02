package remember

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// Cache is the main type for this package.
type Cache struct {
	Client *redis.Client
}

// New is a factory method which returns an instance of Cache.
func New(server, port, password string, db int) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", server, port),
		Password: password,
		DB:       db,
	})

	return &Cache{Client: client}
}

// Set puts a value into Redis. The final parameter, expires, is optional.
func (c *Cache) Set(key string, data any, expires ...time.Duration) error {
	ctx := context.Background()
	var expiration time.Duration
	if len(expires) > 0 {
		expiration = expires[0]
	}

	err := c.Client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

// Get attempts to retrieve a value from the cache.
func (c *Cache) Get(key string) (any, error) {
	ctx := context.Background()

	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	return val, nil
}
