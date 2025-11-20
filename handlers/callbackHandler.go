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

	switch c.Data {
	case "bindKey":
		result = append(result, c.User.BindRandomKey(), msg.VpnConnectMsg(c.User.GetBindedKeys()))
	case "homePage":
		UserData := c.User.GetAccount()
		result = []string{msg.HomeMsg(UserData.Username, UserData.Balance, UserData.Tariff, UserData.Active)}
	case "vpnConnect":
		result = []string{msg.VpnConnectMsg(c.User.GetBindedKeys())}
	case "helpMenu":
		result = []string{msg.HelpMenuMsg()}
	case "paymentMenu":
		UserData := c.User.GetAccount()
		result = []string{msg.PaymentMenuMsg(UserData.Username, UserData.Balance)}
	case "updateBalance":
		UserData := c.User.GetAccount()
		result = []string{msg.PaymentMenuMsg(UserData.Username, UserData.Balance)}
	case "referral":
		result = []string{msg.RefererMsg(fmt.Sprintf("%d", c.User.UserID), "https://t.me/mmcvpnbot")}
	case "donate":
		result = []string{msg.DonateMsg()}
	case "help":
		result = []string{msg.HelpMenuMsg()}
	case "topup_fiat", "topup_crypto":
		UserData := c.User.GetAccount()
		topupSum := int64(100)
		sum := c.User.TopupBalance(topupSum)
		result = append(result, msg.SuccessTopup(sum, topupSum))
		result = append(result, msg.PaymentMenuMsg(UserData.Username, sum))
	}

	return result
}
