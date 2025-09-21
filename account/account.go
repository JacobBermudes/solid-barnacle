package account

import (
	"encoding/json"
	"fmt"
	"mmcvpn/dbaccount"
	"strconv"
	"strings"
	"sync"
)

type InternalAccount struct {
	Userid     int64    `json:"userid"`
	Username   string   `json:"username"`
	Balance    int64    `json:"balance"`
	Tariff     string   `json:"tariff"`
	Adblocker  bool     `json:"adblocker"`
	SharedKeys []string `json:"sharedkey"`
	Active     bool     `json:"active"`
}

type RedisAccount interface {
	GetAccountByID(userid string) dbaccount.DBAccount
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

func (r *InternalAccount) AccountInit(queryChan chan DatabaseQuery) *InternalAccount {

	DatabaseAccount := dbaccount.DBAccount{}
	DatabaseAccount = DatabaseAccount.GetAccountByID(fmt.Sprintf("%d", r.Userid))

	if DatabaseAccount.UserID == 0 { //Create new acc

		fmt.Println("New user has came up")

		DatabaseAccount.UserID = r.Userid
		DatabaseAccount.Username = r.Username
		DatabaseAccount.Tariff = "Стандартный"
		DatabaseAccount.Active = false

		accountData, err := json.Marshal(DatabaseAccount)
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

func (r *InternalAccount) GetUserID() int64 {
	return r.Userid
}

func (r *InternalAccount) GetUsername() string {
	return r.Username
}

func (r *InternalAccount) GetBalance() int64 {
	return r.Balance
}

func (r *InternalAccount) GetTariff() string {
	return r.Tariff
}

func (r *InternalAccount) GetAdblocker() bool {
	return r.Adblocker
}

func (r *InternalAccount) GetSharedKey(queryChan chan DatabaseQuery) []string {

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

func (r *InternalAccount) GetActive() string {
	desc := "ошибка"
	if r.Active {
		desc = "Активен"
	} else {
		desc = "Отключен"
	}
	return desc
}

func (r *InternalAccount) TopupAccount(sum int64, queryChan chan DatabaseQuery) (int64, error) {
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

func (r *InternalAccount) UpdateBalance(queryChan chan DatabaseQuery) int64 {
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

func (r *InternalAccount) ToggleVpn() (bool, error) {
	r.Active = !r.Active
	return r.Active, nil
}

func (r *InternalAccount) AddKey(queryChan chan DatabaseQuery) string {

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
