package dbaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

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

type DBAccount struct {
	UserID   int64  `json:"userid"`
	Username string `json:"username"`
	Tariff   string `json:"tariff"`
	Active   bool   `json:"active"`
	Created  string `json:"created"`
}

func (d DBAccount) GetAccount() DBAccount {

	accountDataQuery := acc_db.Get(ctx, fmt.Sprintf("%d", d.UserID))

	if accountDataQuery != nil {
		accountData, _ := accountDataQuery.Result()
		json.Unmarshal([]byte(accountData), &d)
	}

	return d
}

func (d DBAccount) SetAccountByID(userid string) DBAccount {

	accountData, _ := json.Marshal(d)

	acc_db.Set(ctx, userid, string(accountData), 0)

	return d
}

func (d DBAccount) GetBalance() int64 {

	var balanceNumeric int64 = 0
	balanceQuery, err := balance_db.Get(ctx, fmt.Sprintf("%d", d.UserID)).Int64()

	if err != redis.Nil {
		balanceNumeric = balanceQuery
	}

	return balanceNumeric
}

func (d DBAccount) IncrBalance(sum int64) int64 {
	return balance_db.IncrBy(ctx, fmt.Sprintf("%d", d.UserID), sum).Val()
}

func (d DBAccount) DecrBalance(sum int64) int64 {

	currentBalance := d.GetBalance()

	if currentBalance < sum {
		return 0
	}

	balanceNumeric := acc_db.DecrBy(ctx, fmt.Sprintf("%d", d.UserID), sum).Val()
	return balanceNumeric
}
