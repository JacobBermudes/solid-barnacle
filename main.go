package main

import (
	"context"
	"fmt"
	"log"
	"mmcvpn/account"
	"mmcvpn/msg"
	"os"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type mmcMsg interface {
	HomeMsg(username string, balance int64, tariff string, adblocker bool, active string) tgbotapi.MessageConfig
	VpnConnectMsg() tgbotapi.MessageConfig
	BalanceEditMsg() tgbotapi.MessageConfig
}

var messenger mmcMsg = msg.MessageCreator{}

type RedisReader interface {
	AccountInit(queryChan chan account.DatabaseQuery)
	GetUsername() string
	GetBalance() int64
	GetTariff() string
	GetAdblocker() bool
	GetUserID() int64
	GetActive() string
	GetSharedKey() string
	ToggleVpn() (bool, error)
	TopupAccount(int64) (int64, error)
	AddSharedKey(string) string
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

			var accountReader RedisReader = &account.RedisAccount{}

			if update.CallbackQuery != nil {

				callback := update.CallbackQuery

				accountReader.AccountInit(queryChan)

				log.Printf("Received callback: %s. From user %s", callback.Data, callback.From.UserName)

				msg, showHome := menuCallbackHandler(callback.Data, accountReader)

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

				if update.Message.IsCommand() {

					msg := commandHandler(update.Message.Command(), accountReader)

					if _, err := bot.Send(msg); err != nil {
						log.Printf("Ошибка отправки сообщения в чат: %v", err)
					}
				} else {

					if key_sender == accountReader.GetUserID() {
						accountReader.AddSharedKey(update.Message.Text)
						key_sender = 0
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ключ успешно добавлен!")

						if _, err := bot.Send(msg); err != nil {
							log.Printf("Ошибка отправки сообщения в чат: %v", err)
						}

						return
					}

					accountReader.AccountInit(queryChan)

					homeMsg := messenger.HomeMsg(accountReader.GetUsername(), accountReader.GetBalance(), accountReader.GetTariff(), accountReader.GetAdblocker(), accountReader.GetActive())
					homeMsg.ChatID = update.Message.Chat.ID

					if _, err := bot.Send(homeMsg); err != nil {
						log.Printf("Ошибка отправки сообщения в чат: %v", err)
					}
				}
			}
		}(update)

	}
}

func menuCallbackHandler(data string, acc RedisReader) (tgbotapi.MessageConfig, bool) {

	switch data {
	case "vpnConnect":
		return messenger.VpnConnectMsg(), false
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
	case "keys":

		takenKey := acc.GetSharedKey()

		return tgbotapi.NewMessage(0, takenKey), true
	}

	return tgbotapi.NewMessage(0, "Ошибка разбора команды. Пожалуйста обратитесь в поддержку."), true
}

func commandHandler(command string, acc RedisReader) tgbotapi.MessageConfig {
	switch command {
	case "addkey":
		key_sender = acc.GetUserID()
		return tgbotapi.NewMessage(0, "Введите ключ без vpn:// ")
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
		switch query.QueryType {
		case "addkey":
			keysDb.SAdd(ctx, "nonactive_keys", query.Query)
			query.ReplyChan <- account.DatabaseAnswer{
				Result: "Ключ успешно добавлен!",
				Err:    nil,
			}
			return
		case "getAccDB":
			result, err := accountDb.Get(ctx, query.Query).Result()
			query.ReplyChan <- account.DatabaseAnswer{
				Result: result,
				Err:    err,
			}
			return
		case "setAccDB":
			err := accountDb.Set(ctx, fmt.Sprintf("%d", query.UserID), query.Query, 0).Err()
			query.ReplyChan <- account.DatabaseAnswer{
				Result: "Запись успешно завершена!",
				Err:    err,
			}

		}
	}

}
