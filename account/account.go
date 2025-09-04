package account

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type RedisAccount struct {
	Userid     int64    `json:"userid"`
	Username   string   `json:"username"`
	Balance    int64    `json:"balance"`
	Tariff     string   `json:"tariff"`
	Adblocker  bool     `json:"adblocker"`
	SharedKeys []string `json:"sharedkey"`
	Active     bool     `json:"active"`
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

	if answer.Err != nil && len(answer.Result) == 0 {
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

		r.Balance = 0
		r.Tariff = newAcc.Tariff
		r.Adblocker = false
		r.SharedKeys = []string{}
		r.Active = newAcc.Active

		return
	} else {
		err := json.Unmarshal([]byte(answer.Result), &r)
		if err != nil {
			fmt.Println("Ошибка парсинга данных из бд")
		}

		query = DatabaseQuery{
			UserID:    r.Userid,
			QueryType: "getBalance",
			Query:     "",
			ReplyChan: make(chan DatabaseAnswer),
		}
		queryChan <- query
		answer = <-query.ReplyChan
		balance, err := strconv.ParseInt(answer.Result, 10, 64)
		if err != nil {
			fmt.Println("Ошибка преобразования баланса:", err)
			r.Balance = 0
		} else {
			r.Balance = balance
		}

		query = DatabaseQuery{
			UserID:    r.Userid,
			QueryType: "getKeysList",
			Query:     fmt.Sprintf("%d", r.Userid),
			ReplyChan: make(chan DatabaseAnswer),
		}
		queryChan <- query
		answer = <-query.ReplyChan

		r.SharedKeys = strings.Split(answer.Result, ",")
		r.Adblocker = false
		r.Active = len(strings.Split(answer.Result, ",")) == 0
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

func (r *RedisAccount) GetSharedKey(queryChan chan DatabaseQuery) []string {

	if !r.Active {
		r.Active = true
	}

	if len(r.SharedKeys) == 0 {
		query := DatabaseQuery{
			UserID:    r.Userid,
			QueryType: "getKeysList",
			Query:     fmt.Sprintf("%d", r.Userid),
			ReplyChan: make(chan DatabaseAnswer),
		}
		queryChan <- query
		answer := <-query.ReplyChan

		if len(strings.Split(answer.Result, ",")) == 0 {
			r.SharedKeys = r.addKey(queryChan)
		} else {
			r.SharedKeys = strings.Split(answer.Result, ",")
		}
	}

	return r.SharedKeys
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
	return sum, nil
}

func (r *RedisAccount) ToggleVpn() (bool, error) {
	r.Active = !r.Active
	return r.Active, nil
}

func (r *RedisAccount) addKey(queryChan chan DatabaseQuery) []string {

	query := DatabaseQuery{
		UserID:    r.Userid,
		QueryType: "pickupKey",
		Query:     fmt.Sprintf("%d", r.Userid),
		ReplyChan: make(chan DatabaseAnswer),
	}

	queryChan <- query
	answer := <-query.ReplyChan

	r.SharedKeys = append(r.SharedKeys, answer.Result)

	return r.SharedKeys
}
