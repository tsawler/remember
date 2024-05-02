package remember

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
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

// UnmarshalBinary takes a value retrieved from the cache, which will be a JSON string.
// We unmarshal it into value, which must be a pointer. Any non-scalar type we want
// to store in the cache must implement the MarshalBinary method, i.e.:
//
//	 func (m Student) MarshalBinary() ([]byte, error) {
//		  return json.Marshal(m)
//	 }
func (c *Cache) UnmarshalBinary(data string, value any) error {
	return json.Unmarshal([]byte(data), value)
}

// Set puts a value into Redis. The final parameter, expires, is optional.
func (c *Cache) Set(key string, data any, expires ...time.Duration) error {
	ctx := context.Background()
	var expiration time.Duration
	if len(expires) > 0 {
		expiration = expires[0]
	}

	return c.Client.Set(ctx, key, data, expiration).Err()
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

// GetInt to retrieve a value from the cache, convert it to an int, and return it.
func (c *Cache) GetInt(key string) (int, error) {
	ctx := context.Background()

	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// GetFloat64 to retrieve a value from the cache, convert it to an float64, and return it.
func (c *Cache) GetFloat64(key string) (float64, error) {
	ctx := context.Background()

	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	i, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// GetFloat32 to retrieve a value from the cache, convert it to an float32, and return it.
func (c *Cache) GetFloat32(key string) (float64, error) {
	ctx := context.Background()

	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	i, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// Delete removes an item from the cache, by key.
func (c *Cache) Delete(key string) error {
	ctx := context.Background()
	return c.Client.Del(ctx, key).Err()
}

// Has checks to see if the supplied key is in the cache and returns true if found,
// otherwise false.
func (c *Cache) Has(key string) bool {
	ctx := context.Background()
	res, err := c.Client.Exists(ctx, key).Result()
	if res == 0 || err != nil {
		return false
	}

	return true
}
