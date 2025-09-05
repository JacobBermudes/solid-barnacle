package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mmcvpn/account"
	"mmcvpn/msg"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type mmcMsg interface {
	HomeMsg(username string, balance int64, tariff string, adblocker bool, active string) tgbotapi.MessageConfig
	VpnConnectMsg(currentKeys []string) tgbotapi.MessageConfig
	BalanceEditMsg() tgbotapi.MessageConfig
}

var messenger mmcMsg = msg.MessageCreator{}

type RedisReader interface {
	AccountInit(queryChan chan account.DatabaseQuery) *account.RedisAccount
	GetUsername() string
	GetBalance() int64
	GetTariff() string
	GetAdblocker() bool
	GetUserID() int64
	GetActive() string
	GetSharedKey(queryChan chan account.DatabaseQuery) []string
	ToggleVpn() (bool, error)
	TopupAccount(int64) (int64, error)
	AddKey(queryChan chan account.DatabaseQuery) string
}

var key_sender int64

func main() {

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable not set")
	}
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	queryChan := make(chan account.DatabaseQuery, 100)
	go DBWorker(queryChan)

	for update := range updates {

		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		go func(update tgbotapi.Update) {

			var accountReader RedisReader = &account.RedisAccount{
				Userid:   update.SentFrom().ID,
				Username: update.SentFrom().UserName,
			}

			if update.CallbackQuery != nil {

				callback := update.CallbackQuery

				accountReader.AccountInit(queryChan)

				log.Printf("Received callback: %s. From user %s", callback.Data, callback.From.UserName)

				msg, showHome := menuCallbackHandler(callback.Data, accountReader, queryChan)

				msg.ChatID = callback.Message.Chat.ID

				releaseButton := tgbotapi.NewCallback(callback.ID, "")

				bot.Send(msg)

				if showHome {
					homeMsg := messenger.HomeMsg(accountReader.GetUsername(), accountReader.GetBalance(), accountReader.GetTariff(), accountReader.GetAdblocker(), accountReader.GetActive())
					homeMsg.ChatID = callback.Message.Chat.ID
					bot.Send(homeMsg)
					releaseButton.Text = "Вернулись в главное меню!"
				}

				bot.Request(releaseButton)
			} else if update.Message != nil {

				if key_sender == accountReader.GetUserID() {

					query := account.DatabaseQuery{
						UserID:    0,
						QueryType: "addkey",
						Query:     update.Message.Text,
						ReplyChan: make(chan account.DatabaseAnswer),
					}

					queryChan <- query
					answer := <-query.ReplyChan
					fmt.Printf("%s", answer.Result)
					key_sender = 0

					msg := tgbotapi.NewMessage(update.FromChat().ID, "Ключ успешно добавлен")
					msg.ChatID = update.FromChat().ID

					bot.Send(msg)
				}

				if update.Message.IsCommand() {

					msg := commandHandler(update.Message.Command(), accountReader, queryChan)
					msg.ChatID = update.FromChat().ID

					if _, err := bot.Send(msg); err != nil {
						log.Printf("Ошибка отправки сообщения в чат: %v", err)
					}
				} else {

					//accountReader.AccountInit(queryChan)

					homeMsg := tgbotapi.NewMessage(0, "Test")
					homeMsg.ChatID = update.FromChat().ID

					if _, err := bot.Send(homeMsg); err != nil {
						log.Printf("Ошибка отправки сообщения в чат: %v", err)
					}
				}
			}
		}(update)

	}
}

func menuCallbackHandler(data string, acc RedisReader, queryChan chan account.DatabaseQuery) (tgbotapi.MessageConfig, bool) {

	switch data {
	case "addkey":
		msg := tgbotapi.NewMessage(0, "")
		answer := acc.AddKey(queryChan)
		if answer == "Ключей как будто бы и нет..." || answer == "Максимильное количество ключей" {
			msg.Text = answer
		} else {
			msg.Text = fmt.Sprintf("Ключ `%s` >успешно привязан к аккаунту!", answer)
			msg.ParseMode = "Markdown"

		}
		return msg, true
	case "vpnConnect":
		return messenger.VpnConnectMsg(acc.GetSharedKey(queryChan)), false
	case "toggleVpn":

		_, err := acc.ToggleVpn()
		if err != nil {
			return tgbotapi.NewMessage(0, "Ошибка изменения статуса активности VPN!"), true
		}

		return tgbotapi.NewMessage(0, fmt.Sprintf("VPN успешно %s.", acc.GetActive())), true
	case "balance":

		return messenger.BalanceEditMsg(), false
	case "topup_fiat":

		sum, err := acc.TopupAccount(100)
		if err != nil {
			return tgbotapi.NewMessage(0, "Ошибка пополнения баланса!"), true
		}

		return tgbotapi.NewMessage(0, fmt.Sprintf("Баланс успешно пополнен на %d рублей.", sum)), true
	case "topup_crypto":

		sum, err := acc.TopupAccount(100)
		if err != nil {
			return tgbotapi.NewMessage(0, "Ошибка пополнения баланса!"), true
		}

		return tgbotapi.NewMessage(0, fmt.Sprintf("Баланс успешно пополнен на %d рублей.", sum)), true
	}

	return tgbotapi.NewMessage(0, "Ошибка разбора команды. Пожалуйста обратитесь в поддержку."), true
}

func commandHandler(command string, acc RedisReader, queryChan chan account.DatabaseQuery) tgbotapi.MessageConfig {
	switch command {
	case "addkey":
		key_sender = acc.GetUserID()
		return tgbotapi.NewMessage(0, "Ожидаем ключа включая VPN://")
	case "start":
		acc.AccountInit(queryChan)
		return messenger.HomeMsg(acc.GetUsername(), acc.GetBalance(), acc.GetTariff(), acc.GetAdblocker(), acc.GetActive())
	}

	return tgbotapi.NewMessage(0, "Ошибка разбора команды.Обратитесь в поддержку")
}

func DBWorker(queryChan <-chan account.DatabaseQuery) {

	ctx := context.Background()
	var rdbpass = os.Getenv("REDIS_PASS")

	var accountDb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		DB:       0,
		Password: rdbpass,
	})

	var keysDb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		DB:       1,
		Password: rdbpass,
	})

	var balanceDb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		DB:       2,
		Password: rdbpass,
	})

	for query := range queryChan {
		fmt.Printf("\n\n%s", query.QueryType)
		switch query.QueryType {
		case "addkey":
			keysDb.SAdd(ctx, "ready_keys", query.Query)
			query.ReplyChan <- account.DatabaseAnswer{
				Result: "Ключ успешно добавлен!",
				Err:    nil,
			}
		case "getAccDB":
			result, err := accountDb.Get(ctx, query.Query).Result()
			query.ReplyChan <- account.DatabaseAnswer{
				Result: result,
				Err:    err,
			}
		case "setAccDB":
			accountDb.Set(ctx, fmt.Sprintf("%d", query.UserID), query.Query, 0)
			query.ReplyChan <- account.DatabaseAnswer{
				Result: "Запись успешно завершена!",
				Err:    nil,
			}
		case "pickupKey":
			freeKeys, err := keysDb.SMembers(ctx, "ready_keys").Result()
			if err != nil || len(freeKeys) == 0 {
				query.ReplyChan <- account.DatabaseAnswer{
					Result: "Ключей как будто бы и нет...",
					Err:    errors.New(""),
				}
				continue
			}

			ts := keysDb.TxPipeline()
			ts.SRem(ctx, "ready_keys", freeKeys[0])
			ts.SAdd(ctx, query.Query, freeKeys[0])
			_, err = ts.Exec(ctx)
			if err != nil {
				fmt.Println("Ошибка присвоения ключа пользователю...")
			}

			query.ReplyChan <- account.DatabaseAnswer{
				Result: freeKeys[0],
				Err:    err,
			}
		case "getKeysList":
			bindedKeys, err := keysDb.SMembers(ctx, query.Query).Result()

			if err != nil {
				fmt.Println("Ошибка чтения ключей юзегра их бахы!")
				query.ReplyChan <- account.DatabaseAnswer{
					Result: "",
					Err:    err,
				}
			}

			query.ReplyChan <- account.DatabaseAnswer{
				Result: strings.Join(bindedKeys, ","),
				Err:    nil,
			}
		case "getBalance":
			balance, err := balanceDb.Get(ctx, query.Query).Result()
			query.ReplyChan <- account.DatabaseAnswer{
				Result: balance,
				Err:    err,
			}
		}
	}

}
