package main

import (
	"fmt"
	"log"
	"mmcvpn/account"
	"mmcvpn/msg"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type mmcMsg interface {
	HomeMsg(username string, balance int64, tariff string, adblocker bool, active string) tgbotapi.MessageConfig
	SettingsMsg() tgbotapi.MessageConfig
	BalanceEditMsg() tgbotapi.MessageConfig
}

var messenger mmcMsg = msg.MessageCreator{}

type RedisReader interface {
	AccountInit(userid int64, username string)
	GetUsername() string
	GetBalance() int64
	GetTariff() string
	GetAdblocker() bool
	GetUserID() int64
	GetActive() string
	GetSharedKey() string
	ToggleVpn() (bool, error)
	TopupAccount(int64) (int64, error)
}

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

	for update := range updates {

		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		go func(update tgbotapi.Update) {

			var accountReader RedisReader = &account.RedisAccount{}

			if update.CallbackQuery != nil {

				callback := update.CallbackQuery

				accountReader.AccountInit(callback.From.ID, callback.From.UserName)

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

				message := update.Message
				accountReader.AccountInit(message.From.ID, message.From.UserName)

				homeMsg := messenger.HomeMsg(accountReader.GetUsername(), accountReader.GetBalance(), accountReader.GetTariff(), accountReader.GetAdblocker(), accountReader.GetActive())
				homeMsg.ChatID = update.Message.Chat.ID

				if _, err := bot.Send(homeMsg); err != nil {
					log.Printf("Failed to send message: %v", err)
				}
			}
		}(update)

	}
}

func menuCallbackHandler(data string, acc RedisReader) (tgbotapi.MessageConfig, bool) {

	switch data {
	case "settings":
		return messenger.SettingsMsg(), false
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

		return tgbotapi.NewMessage(0, acc.GetSharedKey()), true
	}

	return tgbotapi.NewMessage(0, "Ошибка разбора команды. Пожалуйста обратитесь в поддержку."), true
}
