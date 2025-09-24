package account

import (
	"fmt"
	"mmcvpn/dbaccount"
	"mmcvpn/keys"
	"strings"
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

func (r *InternalAccount) AccountInit() *InternalAccount {

	DatabaseAccount := dbaccount.DBAccount{}
	DatabaseAccount = DatabaseAccount.GetAccountByID(fmt.Sprintf("%d", r.Userid))

	if DatabaseAccount.UserID == 0 { //Create new acc
		fmt.Println("New user has came up")

		DatabaseAccount.UserID = r.Userid
		DatabaseAccount.Username = r.Username
		DatabaseAccount.Tariff = "Стандартный"
		DatabaseAccount.Active = false

		DatabaseAccount.SetAccountByID(fmt.Sprintf("%d", r.Userid))

		r.SharedKeys = []string{}
		r.Tariff = DatabaseAccount.Tariff
		r.Active = DatabaseAccount.Active
		r.Adblocker = false
	} else {
		keysGetter := keys.KeyStorage{
			UserID: DatabaseAccount.UserID,
		}

		r.SharedKeys = keysGetter.GetKeysList(r.Userid)
		r.Tariff = DatabaseAccount.Tariff
		r.Active = DatabaseAccount.Active
		r.Adblocker = false
	}

	r.Balance = DatabaseAccount.GetBalance()

	return r
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
