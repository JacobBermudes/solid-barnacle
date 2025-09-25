package keys

import (
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

var rdbpass = os.Getenv("REDIS_PASS")
var ctx = context.Background()

var db = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	DB:       1,
	Password: rdbpass,
})

type KeyStorage struct {
	UserID int64
	Keys   []string
}

func (k KeyStorage) GetKeysList() []string {
	keys, _ := db.SMembers(ctx, fmt.Sprintf("%d", k.UserID)).Result()
	return keys
}

func (k KeyStorage) AddKeyToStorage() string {
	for key := range k.Keys {
		db.SAdd(ctx, "ready_keys", key)
	}
	return fmt.Sprintf("Добавлено %d ключей", len(k.Keys))
}
