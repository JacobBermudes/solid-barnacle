package msg

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageCreator struct{}

func (m MessageCreator) HomeMsg(username string, balance int64, tariff string, adblocker bool, active string) tgbotapi.MessageConfig {

	msg := tgbotapi.NewMessage(0, "")

	msg.Text = "Бот управления доступом Madjahead VPN" + "\n\n" +
		"Пользователь " + username + "!\n\n" +
		"Твой баланс: " + fmt.Sprintf("%d", balance) + "\n" +
		"Тариф: " + tariff + "\n" +
		"Статус доступа к VPN: " + active + "\n"

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Подключение к VPN", "vpnConnect"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏦 Баланс", "balance"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💵Акция «Приведи друга»💵", "referral"),
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
		msg.Text = "Ключи подключения отсутствуют. Пожалуйста добавьте ключ для подключения к VPN."
	} else {
		for _, vpnKey := range currentKeys {
			msg.Text = msg.Text + fmt.Sprintf("\nКлюч подключения: `%s`", vpnKey)
		}
		msg.Text = msg.Text + "\n\nДля быстрого подключения скопируйте ключ подключения и вставьте в приложении!\n\n"
	}

	msg.Text = msg.Text + "Ссылки на приложение:\n	[IPhone/iPad](https://apps.apple.com/ru/app/defaultvpn/id6744725017)\n	[Android](https://play.google.com/store/apps/details?id=org.amnezia.vpn)\n	[Windows](https://github.com/amnezia-vpn/amnezia-client/releases/download/4.8.9.2/AmneziaVPN_4.8.9.2_windows_x64.exe)\n	[Linux](https://github.com/amnezia-vpn/amnezia-client/releases/download/4.8.9.2/AmneziaVPN_4.8.9.2_linux_x64.tar.zip)\n\n"

	msg.Text = msg.Text + "\nДля получения информации по подключению выберите вашу операционную систему: \n\n"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить ключ", "addkey"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Android", "androidHelpMsg"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("I'Os", "iosHelpMsg"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("I'Os (устаревший способ)", "iosHelpMsg"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Windows", "windowsHelpMsg"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Linux", "linuxHelpMsg"),
		),
	)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	return msg
}
