package handlers

import (
	"fmt"
	"mmcvpn/account"
	"mmcvpn/keys"
	"mmcvpn/msg"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackHandler struct {
	Data            string
	ChatID          int64
	CallbackID      string
	ShowHome        bool
	InternalAccount account.InternalAccount
}

type CallbackResult struct {
	Message     tgbotapi.MessageConfig
	ReplyMarkup tgbotapi.InlineKeyboardMarkup
	NewMessage  tgbotapi.MessageConfig
}

func (c CallbackHandler) HandleCallback() CallbackResult {

	messenger := msg.MessageCreator{
		BotAddress: "https://t.me/mmcvpnbot",
		ChatID:     c.ChatID,
	}

	result := CallbackResult{
		Message:     tgbotapi.NewMessage(c.ChatID, "Ошибка разбора команды. Пожалуйста обратитесь в поддержку."),
		ReplyMarkup: tgbotapi.InlineKeyboardMarkup{},
	}

	switch c.Data {
	case "bindKey":
		msg := tgbotapi.NewMessage(c.ChatID, keys.KeyStorage{
			UserID: c.InternalAccount.Userid,
		}.BindRandomKey())
		msg.ParseMode = "Markdown"
		result.Message = msg
		result.NewMessage = messenger.VpnConnectMsg(c.InternalAccount.GetSharedKey())
		c.Data = "vpnConnect"
	case "homePage":
		result.Message = messenger.HomeMsg(c.InternalAccount.GetUsername(), c.InternalAccount.GetBalance(), c.InternalAccount.GetTariff(), c.InternalAccount.GetAdblocker(), c.InternalAccount.GetActive())
	case "vpnConnect":
		result.Message = messenger.VpnConnectMsg(c.InternalAccount.GetSharedKey())
	case "helpMenu":
		result.Message = messenger.HelpMenuMsg()
	case "paymentMenu":
		result.Message = messenger.PaymentMenuMsg(c.InternalAccount.GetUsername(), c.InternalAccount.GetBalance())
	case "updateBalance":
		result.Message = messenger.PaymentMenuMsg(c.InternalAccount.GetUsername(), c.InternalAccount.GetBalance())
	case "referral":
		result.Message = messenger.RefererMsg(fmt.Sprintf("%d", c.InternalAccount.GetUserID()))
	case "donate":
		result.Message = messenger.DonateMsg()
	case "help":
		result.Message = messenger.HelpMenuMsg()
	case "topup_fiat":
		topupSum := int64(100)
		sum := c.InternalAccount.TopupAccount(topupSum)
		result.Message = messenger.SuccessTopup(sum, topupSum)
		result.NewMessage = messenger.PaymentMenuMsg(c.InternalAccount.GetUsername(), sum)
		c.Data = "paymentMenu"
	case "topup_crypto":
		topupSum := int64(100)
		sum := c.InternalAccount.TopupAccount(topupSum)
		result.Message = messenger.SuccessTopup(sum, topupSum)
		result.NewMessage = messenger.PaymentMenuMsg(c.InternalAccount.GetUsername(), sum)
		c.Data = "paymentMenu"
	}

	result.ReplyMarkup = messenger.GetInlineKeyboardMarkup(c.Data, c.InternalAccount.GetUserID())
	result.Message.ReplyMarkup = result.ReplyMarkup

	return result
}

func (c CallbackHandler) ShowHomePage() bool {

	ActionsDontShowHome := []string{"homePage", "vpnConnect", "helpMenu", "paymentMenu", "updateBalance", "referral", "help"}

	actionsSet := make(map[string]bool, len(ActionsDontShowHome))
	for _, action := range ActionsDontShowHome {
		actionsSet[action] = true
	}

	_, exists := actionsSet[c.Data]

	return !exists
}
