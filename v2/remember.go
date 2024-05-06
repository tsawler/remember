// Package remember provides an easy way to implement a Redis, BuntDB or Badger cache in your Go application.

package remember

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/redis/go-redis/v9"
	"github.com/tidwall/buntdb"
	"github.com/tsawler/toolbox"
	"time"
)

// CacheInterface is the interface which anything providing cache functionality must satisfy.
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
	Close() error
}

// RedisCache is the type for a Redis-based cache.
type RedisCache struct {
	Conn         *redis.Client
	BadgerClient *badger.DB
	Prefix       string
}

// Options is the type used to configure a CacheInterface object.
type Options struct {
	Server     string // The server where Redis exists.
	Port       string // The port Redis is listening on.
	Password   string // The password for Redis.
	Prefix     string // A prefix to use for all keys for this client.
	DB         int    // Database. Specifying 0 (the default) means use the default database.
	BadgerPath string // The location for the badger database on disk.
	BuntDBPath string // The location for the BuntDB database on disk.
}

// CacheEntry is a map to hold values, so we can serialize them.
type CacheEntry map[string]any

// New is a factory method which returns an instance of a CacheInterface.
func New(cacheType string, o ...Options) (CacheInterface, error) {
	var ops Options
	if len(o) > 0 {
		ops = o[0]
	} else {
		switch cacheType {
		case "redis":
			ops = Options{
				Server:   "localhost",
				Port:     "6379",
				Password: "",
				Prefix:   "dev",
				DB:       0,
			}

		case "badger":
			ops = Options{
				BadgerPath: "./badger",
			}

		case "buntdb":
			ops = Options{
				BuntDBPath: ":memory:",
			}
		}

	}

	switch cacheType {
	case "redis":
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", ops.Server, ops.Port),
			Password: ops.Password,
			DB:       ops.DB,
		})
		return &RedisCache{
			Conn:   client,
			Prefix: ops.Prefix,
		}, nil

	case "badger":
		var t toolbox.Tools
		_ = t.CreateDirIfNotExist(ops.BadgerPath)
		client, err := badger.Open(badger.DefaultOptions(ops.BadgerPath))
		if err != nil {
			return nil, err
		}
		return &BadgerCache{
			Conn:   client,
			Prefix: ops.Prefix,
		}, nil

	case "buntdb":
		client, err := buntdb.Open(ops.BuntDBPath)
		if err != nil {
			return nil, err
		}
		return &BuntDBCache{
			Conn:   client,
			Prefix: ops.Prefix,
		}, nil

	default:
		return nil, errors.New("unsupported cache type")
	}
}

// Close closes the pool of redis connections
func (c *RedisCache) Close() error {
	return c.Conn.Close()
}

// Get attempts to retrieve a value from the cache.
func (c *RedisCache) Get(key string) (any, error) {
	ctx := context.Background()

	val, err := c.Conn.Get(ctx, fmt.Sprintf("%s:%s", c.Prefix, key)).Result()
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
func (c *RedisCache) Set(key string, data any, expires ...time.Duration) error {
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

	return c.Conn.Set(ctx, fmt.Sprintf("%s:%s", c.Prefix, key), string(encoded), expiration).Err()
}

// GetInt is a convenience method which retrieves a value from the cache, converts it to an int, and returns it.
func (c *RedisCache) GetInt(key string) (int, error) {
	val, err := c.Get(key)
	if err != nil {
		return 0, err
	}

	return val.(int), nil
}

// GetString is a convenience method which retrieves a value from the cache and returns it as a string.
func (c *RedisCache) GetString(key string) (string, error) {
	s, err := c.Get(key)
	if err != nil {
		return "", err
	}
	return s.(string), nil
}

// Forget removes an item from the cache, by key.
func (c *RedisCache) Forget(key string) error {
	ctx := context.Background()
	return c.Conn.Del(ctx, fmt.Sprintf("%s:%s", c.Prefix, key)).Err()
}

// Has checks to see if the supplied key is in the cache and returns true if found, otherwise false.
func (c *RedisCache) Has(key string) bool {
	ctx := context.Background()

	res, err := c.Conn.Exists(ctx, fmt.Sprintf("%s:%s", c.Prefix, key)).Result()
	if res == 0 || err != nil {
		return false
	}

	return true
}

// GetTime retrieves a value from the cache by the specified key and returns it as time.Time.
func (c *RedisCache) GetTime(key string) (time.Time, error) {
	fromCache, err := c.Get(key)
	if err != nil {
		return time.Time{}, err
	}

	t := fromCache.(time.Time)
	return t, nil
}

// EmptyByMatch removes all entries in Redis which have the prefix match.
func (c *RedisCache) EmptyByMatch(match string) error {
	ctx := context.Background()

	res, err := c.Conn.Keys(ctx, fmt.Sprintf("%s:%s*", c.Prefix, match)).Result()
	if err != nil {
		return err
	}

	for _, x := range res {
		err := c.Conn.Del(ctx, x).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// Empty removes all entries in Redis for a given client.
func (c *RedisCache) Empty() error {
	ctx := context.Background()

	res, err := c.Conn.Keys(ctx, fmt.Sprintf("%s:*", c.Prefix)).Result()
	if err != nil {
		return err
	}

	for _, x := range res {
		err := c.Conn.Del(ctx, x).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// encode serializes a CacheEntry for storage in the cache.
func encode(item CacheEntry) ([]byte, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(item)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// decode deserializes an item into a map[string]any.
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
