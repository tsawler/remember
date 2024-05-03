package remember

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

// Cache is the main type for this package.
type Cache struct {
	Client *redis.Client
}

// CacheEntry is a map to hold values, so we can serialize them
type CacheEntry map[string]interface{}

// New is a factory method which returns an instance of Cache.
func New(server, port, password string, db int) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", server, port),
		Password: password,
		DB:       db,
	})

	return &Cache{Client: client}
}

// Get attempts to retrieve a value from the cache.
func (c *Cache) Get(key string) (any, error) {
	ctx := context.Background()

	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	decoded, err := decode(val)
	if err != nil {
		fmt.Println("Error decoding", err)
		return nil, err
	}
	item := decoded[key]
	return item, nil
}

// Set puts a value into Redis. The final parameter, expires, is optional.
func (c *Cache) Set(key string, data any, expires ...time.Duration) error {
	ctx := context.Background()

	var expiration time.Duration
	if len(expires) > 0 {
		expiration = expires[0]
	}

	entry := CacheEntry{}
	entry[key] = data
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	return c.Client.Set(ctx, key, string(encoded), expiration).Err()
}

// GetInt to retrieve a value from the cache, convert it to an int, and return it.
func (c *Cache) GetInt(key string) (int, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}

	return val.(int), nil
}

// GetFloat64 to retrieve a value from the cache, convert it to a float64, and return it.
func (c *Cache) GetFloat64(key string) (float64, error) {
	val, err := c.GetString(key)
	if err != nil {
		return 0, err
	}

	i, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// GetFloat32 to retrieve a value from the cache, convert it to a float32, and return it.
func (c *Cache) GetFloat32(key string) (float64, error) {
	val, err := c.GetString(key)
	if err != nil {
		return 0, err
	}

	i, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// GetString to retrieve a value from the cache and return it as a string.
func (c *Cache) GetString(key string) (string, error) {
	s, err := c.Get(key)
	if err != nil {
		return "", err
	}
	return s.(string), nil
}

// Forget removes an item from the cache, by key.
func (c *Cache) Forget(key string) error {
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

// GetTime retrieves a value from the cache by the specified key,
// and returns it as time.Time.
func (c *Cache) GetTime(key string) (time.Time, error) {
	fromCache, err := c.Get(key)
	if err != nil {
		return time.Time{}, err
	}

	t := fromCache.(time.Time)
	return t, nil
}

// Encode serializes item, from a map[string]interface{}
func encode(item CacheEntry) ([]byte, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(item)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Decode unserializes item into a map[string]interface{}
func decode(str string) (CacheEntry, error) {
	item := CacheEntry{}
	b := bytes.Buffer{}
	b.Write([]byte(str))
	d := gob.NewDecoder(&b)
	err := d.Decode(&item)
	if err != nil {
		return nil, err
	}
	return item, nil
}
