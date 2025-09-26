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
	return fmt.Sprintf("Добавлено %d свободных ключей", len(k.Keys))
}

func (k KeyStorage) BindRandomKey() string {
	key, err := db.SPop(ctx, "ready_keys").Result()
	if err == redis.Nil {
		return "Ключей как будто бы и нет..."
	} else if err != nil {
		return "Ошибка при получении ключа"
	}

	currentKeys := db.SMembers(ctx, fmt.Sprintf("%d", k.UserID)).Val()
	if len(currentKeys) > 2 {
		db.SAdd(ctx, "ready_keys", key)
		return "Максимильное количество ключей"
	}

	db.SAdd(ctx, fmt.Sprintf("%d", k.UserID), key)
	return fmt.Sprintf("Ключ ```%s``` успешно привязан к вашему аккаунту!", key)
}
