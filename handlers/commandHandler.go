package handlers

import (
	core "speed-ball/internal/core/data"
	"speed-ball/internal/msg"
)

type CommandHandler struct {
	Data string
	User core.User
}

func (c CommandHandler) HandleCommand() []string {

	var result []string

	switch c.Data {
	case "addkey":
		result = append(result, "Ожидаем ключа включая VPN://")
	case "start":

		User := core.User{
			UserID: c.User.UserID,
		}
		UserData := User.GetAccount()

		result = append(result, msg.HomeMsg(UserData.Username, UserData.Balance, UserData.Tariff, "Активен"))
	}

	return result
}
