package banking

import (
	"encoding/json"
	"log"
	"mmcvpn/account"
	"time"
)

type Bank struct{}

func (b Bank) StartMakePayments() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Ежедневное списание баланса остановлено")
			return
		case <-ticker.C:

			log.Println("Ежедневное списание баланса")

			query := account.DatabaseQuery{
				UserID:    0,
				QueryType: "getAccountsIDs",
				Query:     "*",
				ReplyChan: make(chan account.DatabaseAnswer),
			}

			queryChan <- query
			answer := <-query.ReplyChan

			if answer.Err != nil {
				log.Printf("Ошибка получения списка всех пользователей: %v", answer.Err)
				continue
			}

			var allUsers []string
			if err := json.Unmarshal([]byte(answer.Result), &allUsers); err != nil {
				log.Printf("Error unmarshaling keys: %v", err)
				continue
			}

			for _, tgID := range allUsers {
				query := account.DatabaseQuery{
					UserID:    0,
					QueryType: "getAccDB",
					Query:     tgID,
					ReplyChan: make(chan account.DatabaseAnswer),
				}

				queryChan <- query
				answer := <-query.ReplyChan

				var acc account.DBAccount
				err := json.Unmarshal([]byte(answer.Result), &acc)
				if err != nil {
					log.Printf("Ошибка парсинга данных акунта %s: %v", tgID, err)
					continue
				}

				decrQuery := account.DatabaseQuery{
					UserID:    acc.UserID,
					QueryType: "decrBalance",
					Query:     "3",
					ReplyChan: make(chan account.DatabaseAnswer),
				}

				queryChan <- decrQuery
				decrAnswer := <-decrQuery.ReplyChan
				if decrAnswer.Err != nil {
					log.Printf("Ошибка списания баланса у %s: %v", tgID, answer.Err)
					continue
				}
			}
		}
	}
}
