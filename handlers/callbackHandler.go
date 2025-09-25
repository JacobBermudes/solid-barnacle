package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackHandler struct {
	Data       string
	ChatID     int64
	CallbackID string
	ShowHome   bool
}

func (c CallbackHandler) Handle() tgbotapi.MessageConfig {

	switch c.Data {
	case "bindKey":
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
