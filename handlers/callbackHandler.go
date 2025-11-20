package handlers

import (
	"fmt"
	core "speed-ball/internal/core/data"
	"speed-ball/internal/msg"
)

type CallbackHandler struct {
	Data string
	User core.User
}

func (c CallbackHandler) HandleCallback() []string {

	result := []string{}

	User := core.User{
		UserID: c.User.UserID,
	}

	switch c.Data {
	case "bindKey":
		result = append(result, User.BindRandomKey(), msg.VpnConnectMsg(User.GetBindedKeys()))
	case "homePage":
		UserData := User.GetAccount()
		result = []string{msg.HomeMsg(UserData.Username, UserData.Balance, UserData.Tariff, UserData.Active)}
	case "vpnConnect":
		result = []string{msg.VpnConnectMsg(User.GetBindedKeys())}
	case "helpMenu":
		result = []string{msg.HelpMenuMsg()}
	case "paymentMenu":
		UserData := User.GetAccount()
		result = []string{msg.PaymentMenuMsg(UserData.Username, UserData.Balance)}
	case "updateBalance":
		UserData := User.GetAccount()
		result = []string{msg.PaymentMenuMsg(UserData.Username, UserData.Balance)}
	case "referral":
		result = []string{msg.RefererMsg(fmt.Sprintf("%d", User.UserID), "https://t.me/mmcvpnbot")}
	case "donate":
		result = []string{msg.DonateMsg()}
	case "help":
		result = []string{msg.HelpMenuMsg()}
	case "topup_fiat", "topup_crypto":
		UserData := User.GetAccount()
		topupSum := int64(100)
		sum := User.TopupBalance(topupSum)
		result = append(result, msg.SuccessTopup(sum, topupSum))
		result = append(result, msg.PaymentMenuMsg(UserData.Username, sum))
	}

	return result
}
