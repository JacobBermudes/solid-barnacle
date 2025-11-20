package handlers

import (
	core "speed-ball/internal/core/data"
	"speed-ball/internal/msg"
	"strconv"
	"strings"
)

type CommandHandler struct {
	Data  string
	User  core.User
	Props string
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

		if !User.AccountExist() {

			User = User.SetAccount()

			User.RefBonus(100)

			inviteMakerID, _ := strconv.ParseInt(strings.TrimPrefix(c.Props, "ref"), 10, 64)
			inviteMaker := core.User{
				UserID: inviteMakerID,
			}
			inviteMaker.RefBonus(100)
		}

		UserData := User.GetAccount()

		result = append(result, msg.HomeMsg(UserData.Username, UserData.Balance, UserData.Tariff, "Активен"))
	}

	return result
}
