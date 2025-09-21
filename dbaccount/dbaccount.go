package dbaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var rdbpass = os.Getenv("REDIS_PASS")
var ctx = context.Background()

var db = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	DB:       0,
	Password: rdbpass,
})

type DBAccount struct {
	UserID   int64  `json:"userid"`
	Username string `json:"username"`
	Tariff   string `json:"tariff"`
	Active   bool   `json:"active"`
}

func (d DBAccount) GetAccountByID(userid string) DBAccount {

	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if fmt.Sprintf("%d", d.UserID) != userid {
		accountDataQuery := db.Get(timeout, userid)

		if accountDataQuery != nil {
			accountData, _ := accountDataQuery.Result()
			json.Unmarshal([]byte(accountData), &d)
		}
	}

	return d
}
