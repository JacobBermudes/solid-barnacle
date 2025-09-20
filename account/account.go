package account

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
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

func (r *RedisAccount) AccountInit(queryChan chan DatabaseQuery) *RedisAccount {

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
			return r
		}

		var wg sync.WaitGroup
		wg.Add(1)
		query := DatabaseQuery{
			UserID:    r.Userid,
			QueryType: "setAccDB",
			Query:     string(accountData),
			ReplyChan: make(chan DatabaseAnswer),
		}
		go func(dbquery DatabaseQuery) {
			queryChan <- query
			wg.Done()
		}(query)

		wg.Wait()

		answer := <-query.ReplyChan
		if answer.Err != nil {
			fmt.Println("Error saving account to Redis:", err)
		}

		queryBalance := DatabaseQuery{
			UserID:    r.Userid,
			QueryType: "getBalance",
			Query:     fmt.Sprintf("%d", r.Userid),
			ReplyChan: make(chan DatabaseAnswer),
		}
		queryChan <- queryBalance
		answerBalance := <-queryBalance.ReplyChan

		if answerBalance.Result == "" {
			answerBalance.Result = "0"
		}

		balance, err := strconv.ParseInt(answerBalance.Result, 10, 64)
		if err != nil {
			fmt.Println("Ошибка преобразования баланса:", err)
			r.Balance = 0
		} else {
			r.Balance = balance
		}

		r.Tariff = newAcc.Tariff
		r.Adblocker = false
		r.SharedKeys = []string{}
		r.Active = newAcc.Active

		return r
	} else {
		currDbAcc := DBAccount{}
		err := json.Unmarshal([]byte(answer.Result), &currDbAcc)
		if err != nil {
			fmt.Println("Ошибка парсинга данных из бд")
		}

		r.Tariff = currDbAcc.Tariff
		r.Active = currDbAcc.Active

		query := DatabaseQuery{
			UserID:    r.Userid,
			QueryType: "getBalance",
			Query:     fmt.Sprintf("%d", r.Userid),
			ReplyChan: make(chan DatabaseAnswer),
		}
		queryChan <- query
		answer := <-query.ReplyChan

		if answer.Result == "" {
			answer.Result = "0"
		}

		balance, err := strconv.ParseInt(answer.Result, 10, 64)
		if err != nil {
			fmt.Println("Ошибка преобразования баланса:", err)
			r.Balance = 0
		} else {
			r.Balance = balance
		}

		queryKeyList := DatabaseQuery{
			UserID:    r.Userid,
			QueryType: "getKeysList",
			Query:     fmt.Sprintf("%d", r.Userid),
			ReplyChan: make(chan DatabaseAnswer),
		}
		queryChan <- queryKeyList
		answergetKeyList := <-queryKeyList.ReplyChan

		r.SharedKeys = strings.Split(answergetKeyList.Result, ",")
		r.Adblocker = false
		r.Active = len(strings.Split(answergetKeyList.Result, ",")) >= 0
		return r
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
			r.SharedKeys = []string{r.AddKey(queryChan)}
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

func (r *RedisAccount) TopupAccount(sum int64, queryChan chan DatabaseQuery) (int64, error) {
	r.Balance += sum
	query := DatabaseQuery{
		UserID:    r.Userid,
		QueryType: "topupBalance",
		Query:     fmt.Sprintf("%d", sum),
		ReplyChan: make(chan DatabaseAnswer),
	}

	queryChan <- query
	answer := <-query.ReplyChan

	if answer.Err != nil {
		return 0, answer.Err
	}

	curBalance, err := strconv.ParseInt(answer.Result, 10, 64)
	if err != nil {
		return 0, err
	}

	return curBalance, nil
}

func (r *RedisAccount) UpdateBalance(queryChan chan DatabaseQuery) int64 {
	query := DatabaseQuery{
		UserID:    r.Userid,
		QueryType: "getBalance",
		Query:     fmt.Sprintf("%d", r.Userid),
		ReplyChan: make(chan DatabaseAnswer),
	}

	queryChan <- query
	balance := <-query.ReplyChan

	if balance.Err != nil {
		return 0
	}

	curBalance, err := strconv.ParseInt(balance.Result, 10, 64)
	if err != nil {
		return 0
	}

	r.Balance = curBalance

	return curBalance
}

func (r *RedisAccount) ToggleVpn() (bool, error) {
	r.Active = !r.Active
	return r.Active, nil
}

func (r *RedisAccount) AddKey(queryChan chan DatabaseQuery) string {

	query := DatabaseQuery{
		UserID:    r.Userid,
		QueryType: "pickupKey",
		Query:     fmt.Sprintf("%d", r.Userid),
		ReplyChan: make(chan DatabaseAnswer),
	}

	if len(r.GetSharedKey(queryChan)) == 2 {
		return "Максимильное количество ключей"
	}

	queryChan <- query
	answer := <-query.ReplyChan

	r.SharedKeys = append(r.SharedKeys, answer.Result)
	return answer.Result
}
