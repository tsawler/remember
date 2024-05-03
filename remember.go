//Package remember provides an easy way to implement a Redis cache in your Go application.

package remember

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// CacheInterface is the interface which anything providing
// cache functionality must satisfy.
type CacheInterface interface {
	Empty() error
	EmptyByMatch(match string) error
	Forget(key string) error
	Get(key string) (any, error)
	GetInt(key string) (int, error)
	GetString(key string) (string, error)
	GetTime(key string) (time.Time, error)
	Has(key string) bool
	Set(key string, data any, expires ...time.Duration) error
}

// Cache is the main type for this package.
type Cache struct {
	Client *redis.Client
	Prefix string
}

// Options is the type used to configure a Cache object.
type Options struct {
	Server   string // The server where Redis exists.
	Port     string // The port Redis is listening on.
	Password string // The password for Redis.
	Prefix   string // A prefix to use for all keys for this client.
	DB       int    // Database. Specifying 0 (the default) means use the default database.
}

// CacheEntry is a map to hold values, so we can serialize them.
type CacheEntry map[string]interface{}

// New is a factory method which returns an instance of *Cache,
// which satisfies the CacheInterface, for future development.
func New(o ...Options) CacheInterface {
	var ops Options
	if len(o) > 0 {
		ops = o[0]
	} else {
		ops = Options{
			Server:   "localhost",
			Port:     "6379",
			Password: "",
			Prefix:   "dev",
			DB:       0,
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", ops.Server, ops.Port),
		Password: ops.Password,
		DB:       ops.DB,
	})

	return &Cache{Client: client, Prefix: ops.Prefix}
}

// Get attempts to retrieve a value from the cache.
func (c *Cache) Get(key string) (any, error) {
	ctx := context.Background()

	val, err := c.Client.Get(ctx, fmt.Sprintf("%s:%s", c.Prefix, key)).Result()
	if err != nil {
		return nil, err
	}

	decoded, err := decode(val)
	if err != nil {
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

	return c.Client.Set(ctx, fmt.Sprintf("%s:%s", c.Prefix, key), string(encoded), expiration).Err()
}

// GetInt is a convenience method which retrieves a value from the cache, converts it to an int, and returns it.
func (c *Cache) GetInt(key string) (int, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}

	return val.(int), nil
}

// GetString is a convenience method which retrieves a value from the cache and returns it as a string.
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
	return c.Client.Del(ctx, fmt.Sprintf("%s:%s", c.Prefix, key)).Err()
}

// Has checks to see if the supplied key is in the cache and returns true if found,
// otherwise false.
func (c *Cache) Has(key string) bool {
	ctx := context.Background()

	res, err := c.Client.Exists(ctx, fmt.Sprintf("%s:%s", c.Prefix, key)).Result()
	if res == 0 || err != nil {
		return false
	}

	return true
}

// GetTime retrieves a value from the cache by the specified key and returns it as time.Time.
func (c *Cache) GetTime(key string) (time.Time, error) {
	fromCache, err := c.Get(key)
	if err != nil {
		return time.Time{}, err
	}

	t := fromCache.(time.Time)
	return t, nil
}

// EmptyByMatch removes all entries in Redis which have the prefix match.
func (c *Cache) EmptyByMatch(match string) error {
	ctx := context.Background()

	res, err := c.Client.Keys(ctx, fmt.Sprintf("%s:%s*", c.Prefix, match)).Result()
	if err != nil {
		return err
	}

	for _, x := range res {
		err := c.Client.Del(ctx, x).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// Empty removes all entries in Redis for a given client.
func (c *Cache) Empty() error {
	ctx := context.Background()

	res, err := c.Client.Keys(ctx, fmt.Sprintf("%s:*", c.Prefix)).Result()
	if err != nil {
		return err
	}

	for _, x := range res {
		err := c.Client.Del(ctx, x).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// encode serializes a CacheEntry for storage in Redis.
func encode(item CacheEntry) ([]byte, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(item)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// decode deserializes an item into a map[string]interface{}.
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
