package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type User struct {
	UserID   int64  `json:"userid"`
	Username string `json:"username"`
	Tariff   string `json:"tariff"`
	Active   string `json:"active"`
	Balance  int64  `json:"balance"`
	Created  string `json:"created"`
}

func (d User) AccountExist() bool {

	DataUser := DB_user{
		UserID: d.UserID,
	}
	UserData := DataUser.GetAccount()

	return UserData.Created != ""
}

func (d User) SetAccount() User {

	d.Created = time.Now().Format("2006-01-02 15:04:05")

	accountData, _ := json.Marshal(d)
	DataUser := DB_user{
		UserID: d.UserID,
	}
	DataUser.SetAccount(string(accountData))

	json.Unmarshal(accountData, &d)

	return d
}

func (d User) GetAccount() User {

	DataUser := DB_user{
		UserID: d.UserID,
	}
	UserData := DataUser.GetAccount()

	stringData, _ := json.Marshal(UserData)
	json.Unmarshal(stringData, &d)

	return d
}

func (d User) RefBonus(sum int64) int64 {

	UserWallet := DB_user{
		UserID: d.UserID,
	}

	return UserWallet.TopupBalance(10)
}

func (d User) TopupBalance(sum int64) int64 {

	UserWallet := DB_user{
		UserID: d.UserID,
	}

	return UserWallet.TopupBalance(10)
}

func (d User) BindRandomKey() string {

	KeyHolder := DB_user{
		UserID: d.UserID,
	}

	freeKeys := GetFreeKeys()
	if freeKeys == 0 {
		return "Ключей как будто бы и нет..."
	}

	if len(d.GetBindedKeys()) == 2 {
		return "Максимильное количество ключей для одного пользователя — 2. Удалите ненужные ключи перед добавлением новых."
	}

	bindedKey := KeyHolder.BindRandomKey()
	return fmt.Sprintf("Ключ ```%s``` успешно привязан к вашему аккаунту!", bindedKey)
}

func (d User) GetBindedKeys() []string {

	KeyHolder := DB_user{
		UserID: d.UserID,
	}

	holdedKeys := KeyHolder.GetBindedKeys()

	return holdedKeys
}

func (d User) AddKey(key string) string {
	return "Ключ успешно добавлен в общий пул доступных ключей для раздачи пользователям! На данный момент в пуле доступно " + fmt.Sprintf("%d", GetFreeKeys()) + " ключей."
}

//OLD SHIT

func (d User) GetBalance() int64 {

	var balanceNumeric int64 = 0
	balanceQuery, err := balance_db.Get(ctx, fmt.Sprintf("%d", d.UserID)).Int64()

	if err != redis.Nil {
		balanceNumeric = balanceQuery
	}

	return balanceNumeric
}

func (d User) IncrBalance(sum int64) int64 {
	return balance_db.IncrBy(ctx, fmt.Sprintf("%d", d.UserID), sum).Val()
}

func (d User) DecrBalance(sum int64) int64 {

	currentBalance := d.GetBalance()

	if currentBalance < sum {
		return 0
	}

	balanceNumeric := balance_db.DecrBy(ctx, fmt.Sprintf("%d", d.UserID), sum).Val()
	return balanceNumeric
}
