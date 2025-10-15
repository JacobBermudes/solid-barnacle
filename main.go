package main

import (
	"log"
	"mmcvpn/account"
	"mmcvpn/banking"
	"mmcvpn/handlers"
	"mmcvpn/msg"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/telebot.v4"
	tb "gopkg.in/telebot.v4"
)

type mmcMsg interface {
	HomeMsg_del(username string, balance int64, tariff string, adblocker bool, active string) tgbotapi.MessageConfig
	HomeMsg(username string, balance int64, tariff string, adblocker bool, active string) string
	VpnConnectMsg_del(currentKeys []string) tgbotapi.MessageConfig
	VpnConnectMsg(currentKeys []string) string
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

func startTelebotWebAppHandler(token string) {
	pref := tb.Settings{
		Token: token,
		URL:   "https://api.telegram.org",
		Poller: &tb.Webhook{
			Listen: ":443",
			Endpoint: &tb.WebhookEndpoint{
				PublicURL: "https://phunkao.fun/webhook"},
			TLS: &tb.WebhookTLS{
				Cert: "/home/mickey/phunkao.fun.pem",
				Key:  "/home/mickey/phunkao.fun-key.pem",
			},
		},
	}
	bot, err := tb.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	bot.Handle("/start", func(c telebot.Context) error {
		keyboard := &telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				{
					telebot.InlineButton{
						Unique: "vpnConnect",
						Text:   "⚙️ Подключение к VPN",
					},
				},
				{
					telebot.InlineButton{
						Unique: "paymentMenu",
						Text:   "💲 Внесение оплаты за VPN",
					},
				},
				{
					telebot.InlineButton{
						Unique: "referral",
						Text:   "💵 Акция «Приведи друга»",
					},
				},
				{
					telebot.InlineButton{
						Unique: "donate",
						Text:   "💸 Пожертвовать",
					},
				},
				{
					telebot.InlineButton{
						Unique: "help",
						Text:   "💬 Помощь",
					},
				},
			},
		}

		accountData := &account.InternalAccount{
			Userid:   c.Sender().ID,
			Username: c.Sender().Username,
		}

		if accountData.AccountExist() {
			accountData.AccountInit()
			msgText := messenger.HomeMsg(accountData.GetUsername(), accountData.GetBalance(), accountData.GetTariff(), accountData.GetAdblocker(), accountData.GetActive())
			return c.Send(msgText, keyboard)
		} else {
			accountData.AccountInit()
			accountData.TopupAccount(10)
			// accountData.RefferalBonus()
			msgText := messenger.HomeMsg(accountData.GetUsername(), accountData.GetBalance(), accountData.GetTariff(), accountData.GetAdblocker(), accountData.GetActive())
			return c.Send(msgText, keyboard)
		}
	},
	)

	bot.Handle(tb.OnCallback, func(c telebot.Context) error {
		var accountReader RedisReader = &account.InternalAccount{
			Userid:   c.Sender().ID,
			Username: c.Sender().Username,
		}

		accountReader.AccountInit()

		callbackHandler := handlers.CallbackHandler{
			Data:       c.Callback().Unique,
			ChatID:     c.Message().Chat.ID,
			CallbackID: c.Callback().ID,
			ShowHome:   false,
			InternalAccount: account.InternalAccount{
				Userid:   c.Sender().ID,
				Username: c.Sender().Username,
			},
		}

		return c.Send(callbackHandler.Handle())
	},
	)

	//OLD CODE

	// if update.Message == nil && update.CallbackQuery == nil {
	// 	return nil
	// }

	// go func(update tgbotapi.Update) {

	// 	var accountReader RedisReader = &account.InternalAccount{
	// 		Userid:   update.SentFrom().ID,
	// 		Username: update.SentFrom().UserName,
	// 	}

	// 	if update.CallbackQuery != nil {

	// 		accountReader.AccountInit()

	// 		callbackHandler := handlers.CallbackHandler{
	// 			Data:       update.CallbackQuery.Data,
	// 			ChatID:     update.CallbackQuery.Message.Chat.ID,
	// 			CallbackID: update.CallbackQuery.ID,
	// 			ShowHome:   false,
	// 			InternalAccount: account.InternalAccount{
	// 				Userid:   update.SentFrom().ID,
	// 				Username: update.SentFrom().UserName,
	// 			},
	// 		}

	// 		bot.Send(callbackHandler.Handle())

	// 		releaseButton := tgbotapi.NewCallback(callbackHandler.CallbackID, "")
	// 		if callbackHandler.ShowHomePage() {
	// 			homeMsg := messenger.HomeMsg(accountReader.GetUsername(), accountReader.GetBalance(), accountReader.GetTariff(), accountReader.GetAdblocker(), accountReader.GetActive())
	// 			homeMsg.ChatID = callbackHandler.ChatID
	// 			bot.Send(homeMsg)
	// 			releaseButton.Text = "Вернулись в главное меню!"
	// 		}
	// 		bot.Request(releaseButton)
	// 	} else if update.Message != nil {

	// 		if key_sender == accountReader.GetUserID() {

	// 			key_sender = 0

	// 			result := keys.KeyStorage{
	// 				Keys: []string{
	// 					update.Message.Text,
	// 				},
	// 			}.AddKeyToStorage()

	// 			msg := tgbotapi.NewMessage(update.Message.Chat.ID, result)
	// 			bot.Send(msg)
	// 		}

	// 		if update.Message.IsCommand() {

	// 			refArgs := update.Message.CommandArguments()
	// 			if strings.HasPrefix(refArgs, "ref") {
	// 				refID := strings.TrimPrefix(refArgs, "ref")
	// 				if refID != fmt.Sprintf("%d", accountReader.GetUserID()) && !accountReader.AccountExist() {
	// 					msg := messenger.ThanksMsg()
	// 					msg.ChatID = update.Message.Chat.ID
	// 					bot.Send(msg)

	// 					friendID, _ := strconv.ParseInt(refID, 10, 64)
	// 					referralBonus := account.ReferralBonus{
	// 						FriendID: accountReader.GetUserID(),
	// 						CallerID: friendID,
	// 						Sum:      10,
	// 					}
	// 					msgBalance := tgbotapi.NewMessage(update.Message.Chat.ID, referralBonus.ApplyBonus())
	// 					bot.Send(msgBalance)

	// 				}

	// 			}

	// 			commandHandler := handlers.CommandHandler{
	// 				ChatID:  update.Message.Chat.ID,
	// 				Command: update.Message.Command(),
	// 				InternalAccount: account.InternalAccount{
	// 					Userid:   update.SentFrom().ID,
	// 					Username: update.SentFrom().UserName,
	// 				},
	// 			}

	// 			if update.Message.Command() == "addkey" {
	// 				key_sender = update.SentFrom().ID
	// 			}

	// 			commandHandledMsg := commandHandler.Handle()
	// 			bot.Send(commandHandledMsg)
	// 			return
	// 		}
	// 	}
	// }(update)

	go bot.Start()
}

func main() {

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable not set")
	}
	// bot, err := tgbotapi.NewBotAPI(botToken)
	// if err != nil {
	// 	log.Fatalf("Failed to create bot: %v", err)
	// }
	// bot.Debug = false
	// log.Printf("Authorized on account %s", bot.Self.UserName)
	// u := tgbotapi.NewUpdate(0)
	// u.Timeout = 60
	// updates := bot.GetUpdatesChan(u)

	go banking.Bank{}.StartMakePayments()

	startTelebotWebAppHandler(botToken)

	// for update := range updates {

	// 	if update.Message == nil && update.CallbackQuery == nil {
	// 		continue
	// 	}

	// 	go func(update tgbotapi.Update) {

	// 		var accountReader RedisReader = &account.InternalAccount{
	// 			Userid:   update.SentFrom().ID,
	// 			Username: update.SentFrom().UserName,
	// 		}

	// 		if update.CallbackQuery != nil {

	// 			accountReader.AccountInit()

	// 			callbackHandler := handlers.CallbackHandler{
	// 				Data:       update.CallbackQuery.Data,
	// 				ChatID:     update.CallbackQuery.Message.Chat.ID,
	// 				CallbackID: update.CallbackQuery.ID,
	// 				ShowHome:   false,
	// 				InternalAccount: account.InternalAccount{
	// 					Userid:   update.SentFrom().ID,
	// 					Username: update.SentFrom().UserName,
	// 				},
	// 			}

	// 			bot.Send(callbackHandler.Handle())

	// 			releaseButton := tgbotapi.NewCallback(callbackHandler.CallbackID, "")
	// 			if callbackHandler.ShowHomePage() {
	// 				homeMsg := messenger.HomeMsg(accountReader.GetUsername(), accountReader.GetBalance(), accountReader.GetTariff(), accountReader.GetAdblocker(), accountReader.GetActive())
	// 				homeMsg.ChatID = callbackHandler.ChatID
	// 				bot.Send(homeMsg)
	// 				releaseButton.Text = "Вернулись в главное меню!"
	// 			}
	// 			bot.Request(releaseButton)
	// 		} else if update.Message != nil {

	// 			if key_sender == accountReader.GetUserID() {

	// 				key_sender = 0

	// 				result := keys.KeyStorage{
	// 					Keys: []string{
	// 						update.Message.Text,
	// 					},
	// 				}.AddKeyToStorage()

	// 				msg := tgbotapi.NewMessage(update.Message.Chat.ID, result)
	// 				bot.Send(msg)
	// 			}

	// 			if update.Message.IsCommand() {

	// 				refArgs := update.Message.CommandArguments()
	// 				if strings.HasPrefix(refArgs, "ref") {
	// 					refID := strings.TrimPrefix(refArgs, "ref")
	// 					if refID != fmt.Sprintf("%d", accountReader.GetUserID()) && !accountReader.AccountExist() {
	// 						msg := messenger.ThanksMsg()
	// 						msg.ChatID = update.Message.Chat.ID
	// 						bot.Send(msg)

	// 						friendID, _ := strconv.ParseInt(refID, 10, 64)
	// 						referralBonus := account.ReferralBonus{
	// 							FriendID: accountReader.GetUserID(),
	// 							CallerID: friendID,
	// 							Sum:      10,
	// 						}
	// 						msgBalance := tgbotapi.NewMessage(update.Message.Chat.ID, referralBonus.ApplyBonus())
	// 						bot.Send(msgBalance)

	// 					}

	// 				}

	// 				commandHandler := handlers.CommandHandler{
	// 					ChatID:  update.Message.Chat.ID,
	// 					Command: update.Message.Command(),
	// 					InternalAccount: account.InternalAccount{
	// 						Userid:   update.SentFrom().ID,
	// 						Username: update.SentFrom().UserName,
	// 					},
	// 				}

	// 				if update.Message.Command() == "addkey" {
	// 					key_sender = update.SentFrom().ID
	// 				}

	// 				commandHandledMsg := commandHandler.Handle()
	// 				bot.Send(commandHandledMsg)
	// 				return
	// 			}
	// 		}
	// 	}(update)

	// }
}
