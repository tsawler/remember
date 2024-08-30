package remember

import (
	"github.com/dgraph-io/badger/v3"
	"time"
)

// BadgerCache is the type for a Badger database cache.
type BadgerCache struct {
	Conn   *badger.DB
	Prefix string
}

// Has checks for existence of item in cache.
func (b *BadgerCache) Has(str string) bool {
	_, err := b.Get(str)
	if err != nil {
		return false
	}
	return true
}

// Close closes the badger database.
func (b *BadgerCache) Close() error {
	return b.Conn.Close()
}

// Get attempts to retrieve a value from the cache.
func (b *BadgerCache) Get(str string) (any, error) {
	var fromCache []byte

	err := b.Conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(str))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			fromCache = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	decoded, err := decode(string(fromCache))
	if err != nil {
		return nil, err
	}

	item := decoded[str]

	return item, nil
}

// Set puts a value into Badger. The final parameter, expires, is optional.
func (b *BadgerCache) Set(str string, value any, expires ...time.Duration) error {
	entry := CacheEntry{}

	entry[str] = value
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	if len(expires) > 0 {
		err = b.Conn.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(str), encoded).WithTTL(expires[0])
			err = txn.SetEntry(e)
			return err
		})
	} else {
		err = b.Conn.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry([]byte(str), encoded)
			err = txn.SetEntry(e)
			return err
		})
	}

	return nil
}

// Forget removes an item from the cache, by key.
func (b *BadgerCache) Forget(str string) error {
	err := b.Conn.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(str))
		return err
	})

	return err
}

// EmptyByMatch removes all entries in Redis which have the prefix match.
func (b *BadgerCache) EmptyByMatch(str string) error {
	return b.emptyByMatch(str)
}

// Empty removes all entries in Badger.
func (b *BadgerCache) Empty() error {
	return b.emptyByMatch("")
}

func (b *BadgerCache) emptyByMatch(str string) error {
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := b.Conn.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	collectSize := 100000

	err := b.Conn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0

		for it.Seek([]byte(str)); it.ValidForPrefix([]byte(str)); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					return err
				}
			}
		}

		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

// GetInt is a convenience method which retrieves a value from the cache, converts it to an int, and returns it.
func (b *BadgerCache) GetInt(key string) (int, error) {
	val, err := b.Get(key)
	if err != nil {
		return 0, err
	}

	return val.(int), nil
}

// GetString is a convenience method which retrieves a value from the cache and returns it as a string.
func (b *BadgerCache) GetString(key string) (string, error) {
	s, err := b.Get(key)
	if err != nil {
		return "", err
	}
	return s.(string), nil
}

// GetTime retrieves a value from the cache by the specified key and returns it as time.Time.
func (b *BadgerCache) GetTime(key string) (time.Time, error) {
	fromCache, err := b.Get(key)
	if err != nil {
		return time.Time{}, err
	}

	t := fromCache.(time.Time)
	return t, nil
}
