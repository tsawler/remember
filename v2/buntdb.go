package remember

import (
	"github.com/tidwall/buntdb"
	"log"
	"strings"
	"time"
)

// BuntDBCache is the type for a BuntDB cache.
type BuntDBCache struct {
	Conn   *buntdb.DB
	Prefix string
}

// Has checks to see if the supplied key is in the cache and returns true if found, otherwise false.
func (b *BuntDBCache) Has(str string) bool {
	err := b.Conn.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get(str)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return false
	}
	return true
}

// Close closes the badger database.
func (b *BuntDBCache) Close() error {
	return b.Conn.Close()
}

// Get attempts to retrieve a value from the cache.
func (b *BuntDBCache) Get(str string) (any, error) {
	var fromCache string

	err := b.Conn.View(func(txn *buntdb.Tx) error {
		item, err := txn.Get(str)
		if err != nil {
			return err
		}

		fromCache = item
		return nil
	})
	if err != nil {
		return nil, err
	}

	decoded, err := decode(fromCache)
	if err != nil {
		return nil, err
	}

	item := decoded[str]

	return item, nil
}

// Set puts a value into BuntDB. The final parameter, expires, is optional.
func (b *BuntDBCache) Set(str string, value any, expires ...time.Duration) error {
	entry := CacheEntry{}

	entry[str] = value
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	var so *buntdb.SetOptions

	if len(expires) > 0 {
		so = &buntdb.SetOptions{Expires: true, TTL: expires[0]}
	} else {
		so = nil
	}

	err = b.Conn.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(str, string(encoded), so)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

// Forget removes an item from the cache, by key.
func (b *BuntDBCache) Forget(str string) error {
	err := b.Conn.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(str)
		return err
	})
	return err
}

// EmptyByMatch removes all entries in Redis which have the prefix match.
func (b *BuntDBCache) EmptyByMatch(str string) error {
	return b.emptyByMatch(str)
}

// Empty removes all entries in Badger.
func (b *BuntDBCache) Empty() error {
	return b.emptyByMatch("")
}

func (b *BuntDBCache) emptyByMatch(str string) error {
	var delkeys []string
	err := b.Conn.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			log.Println("key:", key)
			if strings.HasPrefix(key, str) {
				delkeys = append(delkeys, key)
			}
			return true
		})
		return err
	})

	err = b.Conn.Update(func(tx *buntdb.Tx) error {
		for _, k := range delkeys {
			if _, err = tx.Delete(k); err != nil {
				return err
			}
		}
		return err
	})

	return err
}

// GetInt is a convenience method which retrieves a value from the cache, converts it to an int, and returns it.
func (b *BuntDBCache) GetInt(key string) (int, error) {
	val, err := b.Get(key)
	if err != nil {
		return 0, err
	}

	return val.(int), nil
}

// GetString is a convenience method which retrieves a value from the cache and returns it as a string.
func (b *BuntDBCache) GetString(key string) (string, error) {
	s, err := b.Get(key)
	if err != nil {
		return "", err
	}
	return s.(string), nil
}

// GetTime retrieves a value from the cache by the specified key and returns it as time.Time.
func (b *BuntDBCache) GetTime(key string) (time.Time, error) {
	fromCache, err := b.Get(key)
	if err != nil {
		return time.Time{}, err
	}

	t := fromCache.(time.Time)
	return t, nil
}
