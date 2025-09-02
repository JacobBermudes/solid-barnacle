package account

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

var rdbpass = os.Getenv("REDIS_PASS")
var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	DB:       0,
	Password: rdbpass,
})
var ctx = context.Background()

type RedisAccount struct {
	Userid    int64  `json:"userid"`
	Username  string `json:"username"`
	Balance   int64  `json:"balance"`
	Tariff    string `json:"tariff"`
	Adblocker bool   `json:"adblocker"`
	SharedKey string `json:"sharedkey"`
	Active    bool   `json:"active"`
}

type DBAccount struct {
	UserID   int64  `json:"userid"`
	Username string `json:"username"`
	Tariff   string `json:"tariff"`
	Active   bool   `json:"active"`
}

type DatabaseQuery struct {
	UserID    int64
	QueryType string
	Query     string
	ReplyChan chan DatabaseAnswer
}

type DatabaseAnswer struct {
	Result string
	Err    error
}

func (r *RedisAccount) AccountInit(queryChan chan DatabaseQuery) {

	query := DatabaseQuery{
		UserID:    r.Userid,
		QueryType: "getAccDB",
		Query:     fmt.Sprintf("%d", r.Userid),
		ReplyChan: make(chan DatabaseAnswer),
	}
	queryChan <- query

	answer := <-query.ReplyChan

	if answer.Err != nil {
		fmt.Println("New user!")
		newAcc := DBAccount{
			UserID:   r.Userid,
			Username: r.Username,
			Tariff:   "Бесплатный",
			Active:   false,
		}
		accountData, err := json.Marshal(newAcc)
		if err != nil {
			fmt.Println("Error marshaling account:", err)
			return
		}

		query = DatabaseQuery{
			UserID:    r.Userid,
			QueryType: "setAccDB",
			Query:     string(accountData),
			ReplyChan: make(chan DatabaseAnswer),
		}
		queryChan <- query
		answer = <-query.ReplyChan

		if answer.Err != nil {
			fmt.Println("Error saving account to Redis:", err)
		}
		return
	} else {

	}
}

func (r *RedisAccount) GetUserID() int64 {
	return r.Userid
}

func (r *RedisAccount) GetUsername() string {
	return r.Username
}

func (r *RedisAccount) GetBalance() int64 {
	return r.Balance
}

func (r *RedisAccount) GetTariff() string {
	return r.Tariff
}

func (r *RedisAccount) GetAdblocker() bool {
	return r.Adblocker
}

func (r *RedisAccount) GetSharedKey() string {

	if !r.Active {
		r.Active = true
	}

	if r.SharedKey == "" {
		var keysdb = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			DB:       1,
			Password: rdbpass,
		})

		inactiveKeys, err := keysdb.SMembers(ctx, "inactive_keys").Result()
		if err != nil {
			fmt.Println("Error read keys data...")
			return "Ошибка получения ключ доступа к VPN..."
		}

		if len(inactiveKeys) == 0 {
			fmt.Println("Empty keys storage...")
			return "Ошибка получения ключ доступа к VPN..."
		}

		ts := keysdb.TxPipeline()
		ts.SRem(ctx, "inactive_keys", inactiveKeys[0])
		ts.SAdd(ctx, "active_keys", inactiveKeys[0])
		_, err = ts.Exec(ctx)
		if err != nil {
			fmt.Println("Error to save keys for account...")
			return "Ошибка получения ключ доступа к VPN..."
		}

		r.SharedKey = inactiveKeys[0]
		saveAccountData(r)
	}

	return r.SharedKey
}

func (r *RedisAccount) GetActive() string {
	desc := "ошибка"
	if r.Active {
		desc = "Активен"
	} else {
		desc = "Отключен"
	}
	return desc
}

func (r *RedisAccount) TopupAccount(sum int64) (int64, error) {
	r.Balance += sum
	err := saveAccountData(r)
	return sum, err
}

func (r *RedisAccount) ToggleVpn() (bool, error) {
	r.Active = !r.Active
	err := saveAccountData(r)
	return r.Active, err
}

func (r *RedisAccount) AddSharedKey(key string) string {

	var keysdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		DB:       1,
		Password: rdbpass,
	})
	keysdb.SAdd(ctx, "active_keys", key)
	return key
}

func saveAccountData(r *RedisAccount) error {
	accountData, err := json.Marshal(r)

	if err != nil {
		fmt.Println("Error marshaling account:", err)
		return err
	}

	err = rdb.Set(ctx, fmt.Sprintf("%d", r.Userid), accountData, 0).Err()
	if err != nil {
		fmt.Println("Error to save account data to DB!: ", err)
	}

	return err
}
