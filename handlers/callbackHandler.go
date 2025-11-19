package handlers

import (
	"fmt"
	core "speed-ball/internal/core/data"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackHandler struct {
	Data string
	User core.User
}

type CallbackResult struct {
	Message     tgbotapi.MessageConfig
	ReplyMarkup tgbotapi.InlineKeyboardMarkup
	NewMessage  tgbotapi.MessageConfig
}

func (c CallbackHandler) HandleCallback() string {

	switch c.Data {
	case "bindKey":
		BindedKey := core.User{
			UserID: c.User.UserID,
		}.BindRandomKey()
		msg := tgbotapi.NewMessage(c.ChatID, BindedKey)
		msg.ParseMode = "Markdown"
		result.Message = msg
		text := messenger.VpnConnectMsg(c.InternalAccount.GetSharedKey())
		result.NewMessage = tgbotapi.NewMessage(c.ChatID, text)
		result.NewMessage.ParseMode = "Markdown"
		c.Data = "vpnConnect"
	case "homePage":
		text := messenger.HomeMsg(c.InternalAccount.GetUsername(), c.InternalAccount.GetBalance(), c.InternalAccount.GetTariff(), c.InternalAccount.GetAdblocker(), c.InternalAccount.GetActive())
		result.Message = tgbotapi.NewMessage(c.ChatID, text)
	case "vpnConnect":
		text := messenger.VpnConnectMsg(c.InternalAccount.GetSharedKey())
		result.Message = tgbotapi.NewMessage(c.ChatID, text)
	case "helpMenu":
		result.Message = tgbotapi.NewMessage(c.ChatID, messenger.HelpMenuMsg())
	case "paymentMenu":
		result.Message = tgbotapi.NewMessage(c.ChatID, messenger.PaymentMenuMsg(c.InternalAccount.GetUsername(), c.InternalAccount.GetBalance()))
	case "updateBalance":
		result.Message = tgbotapi.NewMessage(c.ChatID, messenger.PaymentMenuMsg(c.InternalAccount.GetUsername(), c.InternalAccount.GetBalance()))
	case "referral":
		result.Message = tgbotapi.NewMessage(c.ChatID, messenger.RefererMsg(fmt.Sprintf("%d", c.InternalAccount.GetUserID())))
	case "donate":
		result.Message = tgbotapi.NewMessage(c.ChatID, messenger.DonateMsg())
	case "help":
		result.Message = tgbotapi.NewMessage(c.ChatID, messenger.HelpMenuMsg())
	case "topup_fiat":
		topupSum := int64(100)
		sum := c.InternalAccount.TopupAccount(topupSum)
		result.Message = tgbotapi.NewMessage(c.ChatID, messenger.SuccessTopup(sum, topupSum))
		result.NewMessage = tgbotapi.NewMessage(c.ChatID, messenger.PaymentMenuMsg(c.InternalAccount.GetUsername(), sum))
		result.NewMessage.ParseMode = "Markdown"
		c.Data = "paymentMenu"
	case "topup_crypto":
		topupSum := int64(100)
		sum := c.InternalAccount.TopupAccount(topupSum)
		result.Message = tgbotapi.NewMessage(c.ChatID, messenger.SuccessTopup(sum, topupSum))
		result.NewMessage = tgbotapi.NewMessage(c.ChatID, messenger.PaymentMenuMsg(c.InternalAccount.GetUsername(), sum))
		result.NewMessage.ParseMode = "Markdown"
		c.Data = "paymentMenu"
	}

	result.ReplyMarkup = messenger.GetInlineKeyboardMarkup(c.Data, c.InternalAccount.GetUserID())
	result.Message.ParseMode = "Markdown"

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
