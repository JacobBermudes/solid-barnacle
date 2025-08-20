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

func (r *RedisAccount) AccountInit(userid int64, username string) {

	dbaccount, err := rdb.Get(ctx, fmt.Sprintf("%d", userid)).Result()
	if err != nil {
		fmt.Println("Error fetching account from Redis:", err)
		fmt.Println("Creating new account")
		r = &RedisAccount{
			Userid:    userid,
			Username:  username,
			Balance:   0,
			Tariff:    "По братски",
			Adblocker: false,
			Active:    false,
			SharedKey: "",
		}
		accountData, err := json.Marshal(r)
		if err != nil {
			fmt.Println("Error marshaling account:", err)
			return
		}
		err = rdb.Set(ctx, fmt.Sprintf("%d", userid), accountData, 0).Err()
		if err != nil {
			fmt.Println("Error saving account to Redis:", err)
			return
		}
		return
	}
	err = json.Unmarshal([]byte(dbaccount), &r)
	if err != nil {
		fmt.Println("Error unmarshaling account:", err)
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
	if r.SharedKey == "" {
		return "Остутствует!"
	}
	return r.SharedKey
}

func (r *RedisAccount) GetActive() string {
	desc := "ошибка"
	if r.Active {
		desc = "включен"
	} else {
		desc = "выключен"
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
