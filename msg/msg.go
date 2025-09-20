package msg

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageCreator struct {
	BotAddress string
}

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
			tgbotapi.NewInlineKeyboardButtonData("Пополнить баланс банковской картой", "topup_fiat"),
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
		for i, vpnKey := range currentKeys {
			if vpnKey == "" {
				msg.Text = "Ключи подключения отсутствуют. Пожалуйста добавьте ключ для подключения к VPN."
				break
			}
			msg.Text = msg.Text + fmt.Sprintf("%d. Ключ подключения: ```%s```\n", i+1, vpnKey)
		}
		msg.Text = msg.Text + "\n\nДля быстрой настройки VPN-подключения скопируйте ключ подключения и вставьте в приложение!\n\n"
	}

	msg.Text = msg.Text + "Ссылки на приложение:\n	[IPhone/iPad](https://apps.apple.com/ru/app/defaultvpn/id6744725017)\n	[Android](https://play.google.com/store/apps/details?id=org.amnezia.vpn)\n	[Windows](https://github.com/amnezia-vpn/amnezia-client/releases/download/4.8.9.2/AmneziaVPN_4.8.9.2_windows_x64.exe)\n	[Linux](https://github.com/amnezia-vpn/amnezia-client/releases/download/4.8.9.2/AmneziaVPN_4.8.9.2_linux_x64.tar.zip)\n\n"

	msg.Text = msg.Text + "\nДля получения информации по подключению выберите вашу операционную систему: \n\n"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Добавить ключ", "addkey"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Помощь с подключением", "helpMenu"),
		),
	)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	return msg
}

func (m MessageCreator) HelpMenuMsg() tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(0, "Для получения дальнейшей помощи по настройке VPN-соединения, пожалуйста, выберите вашу операционную систему:\n\n")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("iPhone/iPad", "https://madjahead.com/help/ios"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Android", "https://madjahead.com/help/android"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Windows", "https://madjahead.com/help/windows"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Linux", "https://madjahead.com/help/linux"),
		),
	)
	return msg

}

func (m MessageCreator) RefererMsg(userid string) tgbotapi.MessageConfig {
	refLink := m.BotAddress + "?start=ref" + userid

	msg := tgbotapi.NewMessage(0, "💵Акция «Приведи друга»💵\n\nПриглашайте друзей и получайте бонусы на баланс!\n\nЗа каждого приглашенного друга вы и ваш друг получит 10 рублей на баланс для тестирования сервиса.\n\nДля участия в акции просто поделитесь своей уникальной ссылкой приглашения:\n\n`"+refLink+"`\n\nЧем больше друзей вы пригласите, тем больше бонусов получите! Акция действует без ограничений по количеству приглашенных друзей.\n\nСпасибо, что выбираете наш VPN-сервис! Вместе мы сделаем интернет безопаснее и доступнее для всех.")

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Главное меню", "homePage"),
			tgbotapi.NewInlineKeyboardButtonSwitch("Поделиться ссылкой", "Присоединяйся к использованию Madjahead VPN по моей ссылке и получи бонус! 🎁\nСсылка для подключения: "+refLink),
		),
	)

	msg.ParseMode = "Markdown"

	return msg
}
