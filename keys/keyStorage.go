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

func (k KeyStorage) GetKeysList(UserID int64) []string {

	keys := []string{}

	bindedKeys := db.SMembers(ctx, fmt.Sprintf("%d", UserID))

	if bindedKeys != nil {
		keys, _ = bindedKeys.Result()
	}

	return keys
}
