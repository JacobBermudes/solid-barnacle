package main

import (
	"log"
	"mmcvpn/account"
	"mmcvpn/handlers"
	"net/http"
	"os"
	"strconv"
	"strings"

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

			bot.Send(callbackReslut.Message)

			if strings.HasPrefix(data, "item_") {
				idStr := strings.TrimPrefix(data, "item_")
				id, _ := strconv.Atoi(idStr)

				fruits := map[int]string{1: "Яблоко", 2: "Банан", 3: "Вишня"}
				name := fruits[id]
				if name == "" {
					name = "Неизвестно"
				}

				edit := tgbotapi.NewEditMessageText(
					callback.Message.Chat.ID,
					callback.Message.MessageID,
					"Вы выбрали: *"+name+"* (ID: "+strconv.Itoa(id)+")",
				)
				edit.ParseMode = "Markdown"
				bot.Send(edit)

				bot.Request(tgbotapi.NewCallback(callback.ID, ""))
			}
		}
	}
}
