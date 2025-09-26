package banking

import (
	"log"
	"mmcvpn/dbaccount"
	"strconv"
	"time"
)

type Bank struct{}

func (b Bank) StartMakePayments() {
	ticker := time.NewTicker(30 * 24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Ежедневное списание баланса")

		allUsers := dbaccount.DBAccount{}.GetAccounts("*")

		for _, tgID := range allUsers {

			numericId, _ := strconv.ParseInt(tgID, 10, 64)
			accountToCharge := dbaccount.DBAccount{
				UserID: numericId,
			}
			newBalance := accountToCharge.DecrBalance(75)
			log.Println("Списано 75 рублей с пользователя: ", numericId, ". Новый баланс: ", newBalance)
		}
	}
}
