package main

import (
	"log"
	"net/http"
	"os"
	"speed-ball/handlers"
	core "speed-ball/internal/core/data"
	"speed-ball/internal/msg"

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

			User := core.User{
				UserID: update.Message.From.ID,
			}

			commandHandler := handlers.CommandHandler{
				Data:  update.Message.Command(),
				User:  User,
				Props: update.Message.CommandArguments(),
			}

			result_msgs := commandHandler.HandleCommand()

			for _, text := range result_msgs {
				newMsg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
				newMsg.ReplyMarkup = msg.GetInlineKeyboardMarkup(update.Message.Command(), User.UserID)
				newMsg.ParseMode = "Markdown"
				newMsg.DisableWebPagePreview = true
				bot.Send(newMsg)
			}

			if update.Message.Command() == "addkey" {
				keySender = update.Message.From.ID
			}
			continue

		}

		if update.Message != nil && keySender == update.Message.From.ID {

			User := core.User{
				UserID: update.Message.From.ID,
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, User.AddKey(update.Message.Text))
			msg.ParseMode = "Markdown"
			bot.Send(msg)

			keySender = 0
			continue
		}

		if update.CallbackQuery != nil {

			User := core.User{
				UserID: update.CallbackQuery.From.ID,
			}

			callback := update.CallbackQuery
			data := callback.Data

			callbackHandler := handlers.CallbackHandler{
				Data: data,
				User: User,
			}

			callbackResult := callbackHandler.HandleCallback()

			for i, text := range callbackResult {
				if i > 0 {
					newMsg := tgbotapi.NewMessage(callback.Message.Chat.ID, text)
					newMsg.ParseMode = "Markdown"
					newMsg.DisableWebPagePreview = true
					bot.Send(newMsg)
				} else {
					editMsg := tgbotapi.NewEditMessageTextAndMarkup(
						callback.Message.Chat.ID,
						callback.Message.MessageID,
						text,
						msg.GetInlineKeyboardMarkup(data, User.UserID))
					editMsg.ParseMode = "Markdown"
					editMsg.DisableWebPagePreview = true
					bot.Send(editMsg)
				}
			}

			bot.Request(tgbotapi.NewCallback(callback.ID, ""))

			continue
		}
	}
}
