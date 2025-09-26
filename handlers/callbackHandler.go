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

func (c CallbackHandler) Handle() tgbotapi.MessageConfig {

	messenger := msg.MessageCreator{
		BotAddress: "https://t.me/mmcvpnbot",
		ChatID:     c.ChatID,
	}

	switch c.Data {
	case "bindKey":
		msg := tgbotapi.NewMessage(c.ChatID, keys.KeyStorage{
			UserID: c.InternalAccount.Userid,
		}.BindRandomKey())
		msg.ParseMode = "Markdown"
		return msg
	case "homePage":
		return messenger.HomeMsg(c.InternalAccount.GetUsername(), c.InternalAccount.GetBalance(), c.InternalAccount.GetTariff(), c.InternalAccount.GetAdblocker(), c.InternalAccount.GetActive())
	case "vpnConnect":
		return messenger.VpnConnectMsg(c.InternalAccount.GetSharedKey())
	case "helpMenu":
		return messenger.HelpMenuMsg()
	case "paymentMenu":
		return messenger.PaymentMenuMsg(c.InternalAccount.GetUsername(), c.InternalAccount.GetBalance())
	case "updateBalance":
		return messenger.PaymentMenuMsg(c.InternalAccount.GetUsername(), c.InternalAccount.GetBalance())
	case "referral":
		return messenger.RefererMsg(fmt.Sprintf("%d", c.InternalAccount.GetUserID()))
	case "donate":
		return messenger.DonateMsg()
	case "help":
		return messenger.HelpMenuMsg()
	case "topup_fiat":
		topupSum := int64(100)
		sum := c.InternalAccount.TopupAccount(topupSum)
		return messenger.SuccessTopup(sum, topupSum)
	case "topup_crypto":
		topupSum := int64(100)
		sum := c.InternalAccount.TopupAccount(topupSum)
		return messenger.SuccessTopup(sum, topupSum)
	}

	return tgbotapi.NewMessage(0, "Ошибка разбора команды. Пожалуйста обратитесь в поддержку.")
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
