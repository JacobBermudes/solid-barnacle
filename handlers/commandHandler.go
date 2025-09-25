package handlers

import (
	"mmcvpn/account"
	"mmcvpn/msg"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandHandler struct {
	ChatID          int64
	Command         string
	InternalAccount account.InternalAccount
}

func (c CommandHandler) Handle() tgbotapi.MessageConfig {

	messenger := msg.MessageCreator{
		BotAddress: "https://t.me/mmcvpnbot",
		ChatID:     c.ChatID,
	}

	switch c.Command {
	case "addkey":
		return tgbotapi.NewMessage(c.ChatID, "Ожидаем ключа включая VPN://")
	case "start":
		c.InternalAccount.AccountInit()
		return messenger.HomeMsg(c.InternalAccount.GetUsername(), c.InternalAccount.GetBalance(), c.InternalAccount.GetTariff(), c.InternalAccount.GetAdblocker(), c.InternalAccount.GetActive())
	case "connect":
		return messenger.VpnConnectMsg(c.InternalAccount.GetSharedKey())
	}

	return tgbotapi.NewMessage(0, "Ошибка разбора команды.Обратитесь в поддержку")
}
