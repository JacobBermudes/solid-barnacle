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

				accountReader.AccountInit()

				callbackHandler := handlers.CallbackHandler{
					Data:       update.CallbackQuery.Data,
					ChatID:     update.CallbackQuery.Message.Chat.ID,
					CallbackID: update.CallbackQuery.ID,
					ShowHome:   false,
					InternalAccount: account.InternalAccount{
						Userid:   update.SentFrom().ID,
						Username: update.SentFrom().UserName,
					},
				}

				bot.Send(callbackHandler.Handle())

				releaseButton := tgbotapi.NewCallback(callbackHandler.CallbackID, "")
				if callbackHandler.ShowHomePage() {
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

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, result)
					bot.Send(msg)
				}

				if update.Message.IsCommand() {

					refArgs := update.Message.CommandArguments()
					if strings.HasPrefix(refArgs, "ref") {
						refID := strings.TrimPrefix(refArgs, "ref")
						if refID != fmt.Sprintf("%d", accountReader.GetUserID()) && !accountReader.AccountExist() {
							msg := messenger.ThanksMsg()
							msg.ChatID = update.Message.Chat.ID
							bot.Send(msg)

							friendID, _ := strconv.ParseInt(refID, 10, 64)
							referralBonus := account.ReferralBonus{
								FriendID: accountReader.GetUserID(),
								CallerID: friendID,
								Sum:      10,
							}
							msgBalance := tgbotapi.NewMessage(update.Message.Chat.ID, referralBonus.ApplyBonus())
							bot.Send(msgBalance)

						}

					}

					commandHandler := handlers.CommandHandler{
						ChatID:  update.Message.Chat.ID,
						Command: update.Message.Command(),
						InternalAccount: account.InternalAccount{
							Userid:   update.SentFrom().ID,
							Username: update.SentFrom().UserName,
						},
					}

					commandHandledMsg := commandHandler.Handle()
					bot.Send(commandHandledMsg)
				}

			}
		}(update)

	}
}
