package main

import (
	"encoding/json"
	"log"
	"mmcvpn/account"
	"mmcvpn/handlers"
	"net/http"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Api_req struct {
	Uid      int64  `json:"uid"`
	Username string `json:"username"`
}

type Api_resp struct {
	Username   string   `json:"username"`
	Balance    int64    `json:"balance"`
	Tariff     string   `json:"tariff"`
	SharedKeys []string `json:"sharedkey"`
	Active     string   `json:"active"`
}

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

		r := http.NewServeMux()
		r.HandleFunc("/webhook/api/init", func(w http.ResponseWriter, r *http.Request) {

			var req Api_req
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				http.Error(w, "Error parsing BODY "+err.Error(), http.StatusBadRequest)
				return
			}

			vpnacc := account.InternalAccount{
				Userid:   req.Uid,
				Username: req.Username,
			}

			vpnacc.AccountInit()

			resp := Api_resp{
				Username:   vpnacc.GetUsername(),
				Balance:    vpnacc.GetBalance(),
				Tariff:     vpnacc.GetTariff(),
				SharedKeys: vpnacc.GetSharedKey(),
				Active:     vpnacc.GetActive(),
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")

			json.NewEncoder(w).Encode(resp)
		})

		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("HTTP Server FAULT:", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(30 * 24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("Ежедневное списание баланса")

			allUsers := account.DBAccount{}.GetAccounts("*")

			for _, tgID := range allUsers {

				numericId, _ := strconv.ParseInt(tgID, 10, 64)
				accountToCharge := account.DBAccount{
					UserID: numericId,
				}
				newBalance := accountToCharge.DecrBalance(75)
				log.Println("Списано 75 рублей с пользователя: ", numericId, ". Новый баланс: ", newBalance)
			}
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
