package account

import (
	"fmt"
	"mmcvpn/dbaccount"
	"mmcvpn/keys"
	"time"
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

type ReferralBonus struct {
	CallerID int64
	FriendID int64
	Sum      int64
}

func (r ReferralBonus) ApplyBonus() string {

	DatabaseAccount := dbaccount.DBAccount{
		UserID: r.CallerID,
	}
	DatabaseAccount.IncrBalance(r.Sum)

	DatabaseAccount = dbaccount.DBAccount{
		UserID: r.FriendID,
	}
	DatabaseAccount.IncrBalance(r.Sum)

	return fmt.Sprintf("Начислено по %d рублей на баланс вам и вашему другу!", r.Sum)
}

func (r *InternalAccount) AccountInit() *InternalAccount {

	DatabaseAccount := dbaccount.DBAccount{
		UserID: r.Userid,
	}
	DatabaseAccount = DatabaseAccount.GetAccount()

	if DatabaseAccount.Created == "" { //Create new acc
		fmt.Println("New user has came up")

		DatabaseAccount.UserID = r.Userid
		DatabaseAccount.Username = r.Username
		DatabaseAccount.Tariff = "Стандартный"
		DatabaseAccount.Active = false
		DatabaseAccount.Created = time.Now().Format("2006-01-02")

		DatabaseAccount.SetAccount()

		r.SharedKeys = []string{}
		r.Tariff = DatabaseAccount.Tariff
		r.Active = DatabaseAccount.Active
		r.Adblocker = false
	} else {
		keysGetter := keys.KeyStorage{
			UserID: DatabaseAccount.UserID,
		}

		r.SharedKeys = keysGetter.GetKeysList()
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
	DatabaseAccount := dbaccount.DBAccount{
		UserID: r.Userid,
	}
	r.Balance = DatabaseAccount.GetBalance()

	return r.Balance
}

func (r *InternalAccount) GetTariff() string {
	return r.Tariff
}

func (r *InternalAccount) GetAdblocker() bool {
	return r.Adblocker
}

func (r *InternalAccount) GetSharedKey() []string {
	keysGetter := keys.KeyStorage{
		UserID: r.Userid,
	}

	r.SharedKeys = keysGetter.GetKeysList()
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

func (r *InternalAccount) AccountExist() bool {
	DatabaseAccount := dbaccount.DBAccount{
		UserID: r.Userid,
	}
	DatabaseAccount = DatabaseAccount.GetAccount()

	return DatabaseAccount.UserID != 0
}

func (r *InternalAccount) RefferalBonus(userid int64, sum int64) int64 {
	DatabaseAccount := dbaccount.DBAccount{
		UserID: userid,
	}
	result := DatabaseAccount.IncrBalance(sum)

	return result
}

func (r *InternalAccount) TopupAccount(sum int64) int64 {
	DatabaseAccount := dbaccount.DBAccount{
		UserID: r.Userid,
	}
	result := DatabaseAccount.IncrBalance(sum)
	return result
}
