package msg

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageCreator struct{}

func (m MessageCreator) HomeMsg(username string, balance int64, tariff string, adblocker bool, active string) tgbotapi.MessageConfig {

	msg := tgbotapi.NewMessage(0, "")

	msg.Text = "Пользователь " + username + "!\n\n" +
		"Твой баланс: " + fmt.Sprintf("%d", balance) + "\n\n" +
		"Тариф: " + tariff + "\n" +
		"Ключ доступа VPN: " + active + "\n\n" +
		"Блокировка рекламы: " + fmt.Sprintf("%t", adblocker) + "\n"

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Управление VPN", "settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏦 Баланс", "balance"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧰 Инструкция по подключению", "instruction"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🍿 Youtube без рекламы", "adblocker"),
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

func (m MessageCreator) SettingsMsg() tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(0, "")
	msg.Text = "Это настройки вашего VPN. Вы можете получить ключ подключения или отключить VPN.\n\n"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Подключить/Отключить VPN", "toggleVpn"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔑 Ключи доступа", "keys"),
		),
	)

	return msg
}
