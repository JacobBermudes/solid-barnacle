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

type CommandResult struct {
	Message tgbotapi.MessageConfig
}

func (c CommandHandler) HandleCommand() CommandResult {

	messenger := msg.MessageCreator{
		BotAddress: "https://t.me/mmcvpnbot",
		ChatID:     c.ChatID,
	}

	result := CommandResult{
		Message: tgbotapi.NewMessage(c.ChatID, "Неизвестная команда.Обратитесь в поддержку"),
	}

	switch c.Command {
	case "addkey":
		result.Message = tgbotapi.NewMessage(c.ChatID, "Ожидаем ключа включая VPN://")
	case "start":
		c.InternalAccount.AccountInit()
		result.Message = messenger.HomeMsg(c.InternalAccount.GetUsername(), c.InternalAccount.GetBalance(), c.InternalAccount.GetTariff(), c.InternalAccount.GetAdblocker(), c.InternalAccount.GetActive())
	case "connect":
		result.Message = messenger.VpnConnectMsg(c.InternalAccount.GetSharedKey())
	}

	return result
}
