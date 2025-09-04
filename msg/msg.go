package msg

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageCreator struct{}

func (m MessageCreator) HomeMsg(username string, balance int64, tariff string, adblocker bool, active string) tgbotapi.MessageConfig {

	msg := tgbotapi.NewMessage(0, "")

	msg.Text = "Бот управления доступом Madjahead VPN" + "\n\n" +
		"Пользователь " + username + "\n\n" +
		"Твой баланс: " + fmt.Sprintf("%d", balance) + "\n\n" +
		"Тариф: " + tariff + "\n" +
		"Статус доступа к VPN: " + active + "\n\n" +
		"Блокировка рекламы: " + fmt.Sprintf("%t", adblocker) + "\n"

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Подключение к VPN", "vpnConnect"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏦 Баланс", "balance"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧰 Инструкция по подключению", "instruction"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🤝 Пригласить друга", "referral"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💸 Пожертвовать", "donate"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 Помощь", "help"),
		),
	)

	return msg
}

func (m MessageCreator) BalanceEditMsg() tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(0, "Это меню управления балансом. Вы можете пополнить баланс несколькими способами\n\n")

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пополнить баланс через эквайринг", "topup_fiat"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Пополнить криптовалютой", "topup_crypto"),
		),
	)

	return msg
}

func (m MessageCreator) VpnConnectMsg(currentKeys []string) tgbotapi.MessageConfig {

	msg := tgbotapi.NewMessage(0, "")
	if len(currentKeys) == 0 {
		msg.Text = ""
	} else {
		for _, vpnKey := range currentKeys {
			msg.Text = msg.Text + fmt.Sprintf("\n\nКлюч подключения `%s`\n\n", vpnKey)
		}
		msg.Text = msg.Text + "\n\nДля быстрого подключения скопируйте ключ подключения и вставьте в приложение Amnezia VPN!\n"
	}
	msg.ParseMode = "Markdown"
	msg.Text = msg.Text + "\n\nДля более получения информации по подключению выберите операционную систему: \n\n"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Android", "androidHelpMsg"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("I'Os", "iosHelpMsg"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Windows", "windowsHelpMsg"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Linux", "linuxHelpMsg"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить ключ", "addkey"),
		),
	)

	return msg
}
