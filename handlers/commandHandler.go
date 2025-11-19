package handlers

import (
	core "speed-ball/internal/core/data"
	"speed-ball/internal/msg"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandHandler struct {
	Data string
	User core.User
}

type CommandResult struct {
	Message tgbotapi.MessageConfig
}

func (c CommandHandler) HandleCommand() string {

	switch c.Data {
	case "addkey":
		return "Ожидаем ключа включая VPN://"
	case "start":

		User := core.User{
			UserID: c.User.UserID,
		}
		UserData := User.GetAccount()

		return msg.HomeMsg(UserData.Username, UserData.Balance, UserData.Tariff, "Активен")
	default:
		return "Неизвестная команда.Обратитесь в поддержку"
	}
}
