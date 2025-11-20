package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

type DB_user struct {
	UserID   int64  `json:"userid"`
	Username string `json:"username"`
	Tariff   string `json:"tariff"`
	Active   bool   `json:"active"`
	Created  string `json:"created"`
}

var rdbpass = os.Getenv("REDIS_PASS")
var ctx = context.Background()

var acc_db = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	DB:       0,
	Password: rdbpass,
})

var balance_db = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	DB:       2,
	Password: rdbpass,
})

var keys_db = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	DB:       1,
	Password: rdbpass,
})

func (d DB_user) GetAccount() DB_user {

	accountDataQuery := acc_db.Get(ctx, fmt.Sprintf("%d", d.UserID))

	fmt.Printf("Getting account for user ID: %d\n", d.UserID)
	fmt.Printf("Redis GET result: %+v\n", accountDataQuery)

	if accountDataQuery != nil {
		accountData, _ := accountDataQuery.Result()
		json.Unmarshal([]byte(accountData), &d)
	}

	return d
}

func (d DB_user) SetAccount(setString string) DB_user {

	acc_db.Set(ctx, fmt.Sprintf("%d", d.UserID), setString, 0)
	return d
}

func (d DB_user) TopupBalance(sum int64) int64 {
	return balance_db.IncrBy(ctx, fmt.Sprintf("%d", d.UserID), sum).Val()
}

func (d DB_user) GetBindedKeys() []string {

	keys, _ := keys_db.SMembers(ctx, fmt.Sprintf("%d", d.UserID)).Result()
	return keys
}

func (d DB_user) BindRandomKey() string {

	bindedKey := keys_db.SPop(ctx, "ready_keys").Val()
	keys_db.SAdd(ctx, fmt.Sprintf("%d", d.UserID), bindedKey)
	return bindedKey
}

func AddKey(key string) bool {
	result := keys_db.SAdd(ctx, "ready_keys", key)
	return result.Err() != redis.Nil
}

func GetFreeKeys() int64 {

	count, err := keys_db.SCard(ctx, "ready_keys").Result()
	if err != nil {
		return 0
	}
	return count
}
