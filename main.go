package main

import (
	"log"
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

	certPath := "/home/mickey/solid-barnacle/fullchain.pem"
	webhook.Certificate = tgbotapi.FilePath(certPath)

	webhook.AllowedUpdates = []string{"message", "callback_query"}

	_, err = bot.Request(webhook)
	if err != nil {
		log.Fatal("Ошибка установки вебхука:", err)
	}
	log.Println("Вебхук установлен:", webhookURL)

	updates := bot.ListenForWebhook("/webhook")

	go func() {
		log.Println("Go-бот слушает на :8443 (HTTP)")
		if err := http.ListenAndServeTLS(":8443", certPath, "/home/mickey/solid-barnacle/privkey.pem", nil); err != nil {
			log.Fatal("Ошибка запуска HTTP-сервера:", err)
		}
	}()

	for update := range updates {
		log.Printf("Получено обновление: %+v", update)

		if update.Message != nil && update.Message.IsCommand() && update.Message.Command() == "start" {
			items := []Item{
				{1, "Яблоко"}, {2, "Банан"}, {3, "Вишня"},
			}

			var rows [][]tgbotapi.InlineKeyboardButton
			for _, item := range items {
				btn := tgbotapi.NewInlineKeyboardButtonData(
					item.Name,
					"item_"+strconv.Itoa(item.ID),
				)
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
			}

			keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите фрукт:")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		}

		if update.CallbackQuery != nil {
			callback := update.CallbackQuery
			data := callback.Data

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
