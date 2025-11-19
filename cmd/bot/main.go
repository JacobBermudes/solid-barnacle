package bot

import (
	"log"
	"net/http"
	"os"
	"speed-ball/account"
	"speed-ball/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Bot create FAIL:", err)
	}

	bot.Debug = true
	log.Printf("Auth as: @%s", bot.Self.UserName)

	webhookURL := "https://www.phunkao.fun:8443/webhook"
	webhook, _ := tgbotapi.NewWebhook(webhookURL)

	webhook.AllowedUpdates = []string{"message", "callback_query"}

	_, err = bot.Request(webhook)
	if err != nil {
		log.Fatal("Setting webhook FAIL:", err)
	}
	log.Println("Webhook setted:", webhookURL)

	updates := bot.ListenForWebhook("/webhook")

	go func() {
		log.Println("Go back listening :8080 (HTTP)")

		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("HTTP WebHook-Server FAULT:", err)
		}
	}()

	keySender := int64(0)

	for update := range updates {
		log.Printf("Get update: %+v", update)

		if update.Message != nil && update.Message.IsCommand() {

			commandHandler := handlers.CommandHandler{
				ChatID:          update.Message.Chat.ID,
				Command:         update.Message.Command(),
				InternalAccount: account.InternalAccount{Userid: update.Message.From.ID, Username: update.Message.From.UserName},
			}

			commandResult := commandHandler.HandleCommand()

			bot.Send(commandResult.Message)

			if update.Message.Command() == "addkey" {
				keySender = update.Message.From.ID
			}
			continue

		}

		if update.Message != nil && keySender == update.Message.From.ID {
			keyStorage := account.KeyStorage{
				UserID: keySender,
				Keys:   []string{update.Message.Text},
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, keyStorage.AddKeyToStorage())
			msg.ParseMode = "Markdown"
			bot.Send(msg)

			keySender = 0
			continue
		}

		if update.CallbackQuery != nil {
			callback := update.CallbackQuery
			data := callback.Data

			callbackHandler := handlers.CallbackHandler{
				Data:            data,
				ChatID:          callback.Message.Chat.ID,
				CallbackID:      callback.ID,
				InternalAccount: account.InternalAccount{Userid: callback.From.ID, Username: callback.From.UserName},
			}

			callbackResult := callbackHandler.HandleCallback()

			editMsg := tgbotapi.NewEditMessageTextAndMarkup(
				callback.Message.Chat.ID,
				callback.Message.MessageID,
				callbackResult.Message.Text,
				callbackResult.ReplyMarkup,
			)
			if callbackResult.NewMessage.Text != "" {
				newMsg := callbackResult.NewMessage
				newMsg.ReplyMarkup = callbackResult.ReplyMarkup
				bot.Send(newMsg)
				editMsg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
				}
			}
			editMsg.ParseMode = callbackResult.Message.ParseMode
			editMsg.DisableWebPagePreview = callbackResult.Message.DisableWebPagePreview
			bot.Send(editMsg)

			bot.Request(tgbotapi.NewCallback(callback.ID, ""))

			continue
		}
	}
}
