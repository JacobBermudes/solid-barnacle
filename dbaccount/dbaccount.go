package dbaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var rdbpass = os.Getenv("REDIS_PASS")
var ctx = context.Background()

var acc_db = redis.NewClient(&redis.Options{
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
		accountDataQuery := acc_db.Get(timeout, userid)

		if accountDataQuery != nil {
			accountData, _ := accountDataQuery.Result()
			json.Unmarshal([]byte(accountData), &d)
		}
	}

	return d
}

func (d DBAccount) SetAccountByID(userid string) DBAccount {

	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	accountData, _ := json.Marshal(d)

	acc_db.Set(timeout, userid, string(accountData), 0)

	return d
}

func (d DBAccount) GetBalance() int64 {

	var balanceNumeric int64 = 0
	balanceQuery := acc_db.Get(ctx, fmt.Sprintf("%d", d.UserID))

	if balanceQuery != nil {
		balance, _ := balanceQuery.Result()
		balanceNumeric, _ = strconv.ParseInt(balance, 10, 64)
	}

	return balanceNumeric
}
