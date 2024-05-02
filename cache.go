package remember

import (
	"fmt"
	"github.com/redis/go-redis/v9"
)

type Cacher struct {
}

func New(server, port, password string, db int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", server, port),
		Password: password,
	})

	return client
}
