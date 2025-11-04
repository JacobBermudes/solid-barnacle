package main

import (
	"log"
	"mmcvpn/account"
	"mmcvpn/handlers"
	"mmcvpn/keys"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Item struct {
	ID   int
	Name string
}

func main() {
	token := os.Getenv("TG_API")
	if token == "" {
		log.Fatal("TG_API не установлен! Установите переменную окружения.")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Ошибка создания бота:", err)
	}

	bot.Debug = true
	log.Printf("Авторизован: @%s", bot.Self.UserName)

	webhookURL := "https://www.phunkao.fun:8443/webhook"
	webhook, _ := tgbotapi.NewWebhook(webhookURL)

	webhook.AllowedUpdates = []string{"message", "callback_query"}

	_, err = bot.Request(webhook)
	if err != nil {
		log.Fatal("Ошибка установки вебхука:", err)
	}
	log.Println("Вебхук установлен:", webhookURL)

	updates := bot.ListenForWebhook("/webhook")

	go func() {
		log.Println("Go-бот слушает на :8080 (HTTP)")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("Ошибка запуска HTTP-сервера:", err)
		}
	}()

	keySender := int64(0)

	for update := range updates {
		log.Printf("Получено обновление: %+v", update)

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
			keyStorage := keys.KeyStorage{
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

			callbackReslut := callbackHandler.HandleCallback()

			editMsg := tgbotapi.NewEditMessageTextAndMarkup(
				callback.Message.Chat.ID,
				callback.Message.MessageID,
				callbackReslut.Message.Text,
				callbackReslut.ReplyMarkup,
			)
			editMsg.ParseMode = callbackReslut.Message.ParseMode
			editMsg.DisableWebPagePreview = callbackReslut.Message.DisableWebPagePreview
			bot.Send(editMsg)

			if callbackReslut.NewMessage.Text != "" {
				newMsg := callbackReslut.NewMessage
				newMsg.ReplyMarkup = callbackReslut.ReplyMarkup
				bot.Send(newMsg)
			}

			bot.Request(tgbotapi.NewCallback(callback.ID, ""))

			continue
		}
	}
}
