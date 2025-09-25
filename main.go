package main

import (
	"fmt"
	"log"
	"mmcvpn/account"
	"mmcvpn/banking"
	"mmcvpn/handlers"
	"mmcvpn/keys"
	"mmcvpn/msg"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type mmcMsg interface {
	HomeMsg(username string, balance int64, tariff string, adblocker bool, active string) tgbotapi.MessageConfig
	VpnConnectMsg(currentKeys []string) tgbotapi.MessageConfig
	PaymentMenuMsg(username string, balance int64) tgbotapi.MessageConfig
	HelpMenuMsg() tgbotapi.MessageConfig
	RefererMsg(userid string) tgbotapi.MessageConfig
	DonateMsg() tgbotapi.MessageConfig
	ThanksMsg() tgbotapi.MessageConfig
}

var messenger mmcMsg = msg.MessageCreator{
	BotAddress: "https://t.me/mmcvpnbot",
}

type RedisReader interface {
	AccountInit() *account.InternalAccount
	GetUsername() string
	GetTariff() string
	GetAdblocker() bool
	GetUserID() int64
	GetActive() string
	GetSharedKey() []string
	GetBalance() int64
	AccountExist() bool
	RefferalBonus(userid int64, sum int64) int64
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

	go banking.Bank{}.StartMakePayments()

	for update := range updates {

		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		go func(update tgbotapi.Update) {

			var accountReader RedisReader = &account.InternalAccount{
				Userid:   update.SentFrom().ID,
				Username: update.SentFrom().UserName,
			}

			if update.CallbackQuery != nil {

				callbackHandler := handlers.CallbackHandler{
					Data:       update.CallbackQuery.Data,
					ChatID:     update.CallbackQuery.Message.Chat.ID,
					CallbackID: update.CallbackQuery.ID,
					ShowHome:   false,
				}

				accountReader.AccountInit()

				responseMsg := callbackHandler.Handle()
				bot.Send(responseMsg)

				releaseButton := tgbotapi.NewCallback(callbackHandler.CallbackID, "")
				if callbackHandler.ShowHome {
					homeMsg := messenger.HomeMsg(accountReader.GetUsername(), accountReader.GetBalance(), accountReader.GetTariff(), accountReader.GetAdblocker(), accountReader.GetActive())
					homeMsg.ChatID = callbackHandler.ChatID
					bot.Send(homeMsg)
					releaseButton.Text = "Вернулись в главное меню!"
				}
				bot.Request(releaseButton)
			} else if update.Message != nil {

				if key_sender == accountReader.GetUserID() {

					key_sender = 0

					result := keys.KeyStorage{
						Keys: []string{
							update.Message.Text,
						},
					}.AddKeyToStorage()

					msg := tgbotapi.NewMessage(update.FromChat().ID, result)
					msg.ChatID = update.FromChat().ID
					bot.Send(msg)
				}

				if update.Message.IsCommand() {

					refArgs := update.Message.CommandArguments()
					if strings.HasPrefix(refArgs, "ref") {
						refID := strings.TrimPrefix(refArgs, "ref")
						if refID != fmt.Sprintf("%d", accountReader.GetUserID()) && !accountReader.AccountExist() {
							msg := messenger.ThanksMsg()
							msg.ChatID = update.FromChat().ID
							bot.Send(msg)

							friendID, _ := strconv.ParseInt(refID, 10, 64)
							referralBonus := account.ReferralBonus{
								CallerID: accountReader.GetUserID(),
								FriendID: friendID,
								Sum:      10,
							}
							msgBalance := tgbotapi.NewMessage(update.FromChat().ID, referralBonus.ApplyBonus())
							msgBalance.ChatID = update.FromChat().ID
							bot.Send(msgBalance)

						}

					}

					msg := commandHandler(update.Message.Command(), accountReader, queryChan)
					msg.ChatID = update.FromChat().ID

					if _, err := bot.Send(msg); err != nil {
						log.Printf("Ошибка отправки сообщения в чат: %v", err)
					}
				}

			}
		}(update)

	}
}

func menuCallbackHandler(data string, acc RedisReader) (tgbotapi.MessageConfig, bool) {

	switch data {
	case "addkey":
		msg := tgbotapi.NewMessage(0, "")
		answer := acc.AddKey(queryChan)
		if answer == "Ключей как будто бы и нет..." || answer == "Максимильное количество ключей" {
			msg.Text = answer
		} else {
			msg.Text = fmt.Sprintf("Ключ ```%s``` >успешно привязан к аккаунту!", answer)
			msg.ParseMode = "Markdown"

		}
		return msg, true
	case "homePage":
		return messenger.HomeMsg(acc.GetUsername(), acc.GetBalance(), acc.GetTariff(), acc.GetAdblocker(), acc.GetActive()), false
	case "vpnConnect":
		return messenger.VpnConnectMsg(acc.GetSharedKey(queryChan)), false
	case "helpMenu":
		return messenger.HelpMenuMsg(), false
	case "paymentMenu":
		return messenger.PaymentMenuMsg(acc.GetUsername(), acc.UpdateBalance(queryChan)), false
	case "updateBalance":
		return messenger.PaymentMenuMsg(acc.GetUsername(), acc.UpdateBalance(queryChan)), false
	case "referral":
		return messenger.RefererMsg(fmt.Sprintf("%d", acc.GetUserID())), false
	case "donate":
		return messenger.DonateMsg(), true
	case "help":
		return messenger.HelpMenuMsg(), false
	case "topup_fiat":
		topupSum := int64(100)
		sum, err := acc.TopupAccount(topupSum, queryChan)
		if err != nil {
			return tgbotapi.NewMessage(0, "Ошибка пополнения баланса!"), true
		}

		return tgbotapi.NewMessage(0, fmt.Sprintf("Баланс успешно пополнен на %d рублей. Итого: %d", topupSum, sum)), true
	case "topup_crypto":
		topupSum := int64(100)
		sum, err := acc.TopupAccount(topupSum, queryChan)
		if err != nil {
			return tgbotapi.NewMessage(0, "Ошибка пополнения баланса!"), true
		}

		return tgbotapi.NewMessage(0, fmt.Sprintf("Баланс успешно пополнен на %d рублей. Итого: %d", topupSum, sum)), true
	}

	return tgbotapi.NewMessage(0, "Ошибка разбора команды. Пожалуйста обратитесь в поддержку."), true
}

func commandHandler(command string, acc RedisReader, queryChan chan account.DatabaseQuery) tgbotapi.MessageConfig {
	fmt.Printf("Received command: %s. From user %s", command, acc.GetUsername())
	switch command {
	case "addkey":
		key_sender = acc.GetUserID()
		return tgbotapi.NewMessage(0, "Ожидаем ключа включая VPN://")
	case "start":
		acc.AccountInit()
		return messenger.HomeMsg(acc.GetUsername(), acc.GetBalance(), acc.GetTariff(), acc.GetAdblocker(), acc.GetActive())
	case "connect":
		return messenger.VpnConnectMsg(acc.GetSharedKey(queryChan))
	}

	return tgbotapi.NewMessage(0, "Ошибка разбора команды.Обратитесь в поддержку")
}
