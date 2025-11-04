package msg

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageCreator struct {
	BotAddress string
	ChatID     int64
}

func (m MessageCreator) HomeMsg(username string, balance int64, tariff string, adblocker bool, active string) tgbotapi.MessageConfig {

	msg := tgbotapi.NewMessage(m.ChatID, "")

	msg.Text = "Бот управления доступом SurfBoost VPN" + "\n\n" +
		"Пользователь " + username + "!\n\n" +
		"Твой баланс: " + fmt.Sprintf("%d", balance) + "\n" +
		"Тариф: " + tariff + "\n" +
		"Статус доступа к VPN: " + active + "\n"

	msg.ReplyMarkup = m.GetInlineKeyboardMarkup("homePage", 0)
	return msg
}

func (m MessageCreator) PaymentMenuMsg(username string, balance int64) tgbotapi.MessageConfig {

	walletData := fmt.Sprintf("Уважаемый %s!\nВаш баланс: %d рублей.\n\n", username, balance)

	msg := tgbotapi.NewMessage(m.ChatID, "💲 *Внесение оплаты!*\n\n"+
		walletData+
		"- Ваш баланс привязан к аккаунту и хранится в нашей базе данных.\n"+
		"- Пополнить баланс можно быстро и удобно через *СБП* (Систему быстрых платежей) — просто выберите нужную сумму и следуйте инструкциям.\n"+
		"- Списание с баланса происходит автоматически согласно вашему тарифу.\n\n"+
		"🔍 Хотите пополнить баланс или изменить тариф? Используйте кнопки ниже! 👇")
	msg.ParseMode = "Markdown"
	return msg
}

func (m MessageCreator) VpnConnectMsg(currentKeys []string) tgbotapi.MessageConfig {

	msg := tgbotapi.NewMessage(m.ChatID, "")
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

	msg.Text = msg.Text + "\nСсылки на приложение:\n	[IPhone/iPad](https://apps.apple.com/ru/app/defaultvpn/id6744725017) \n	[Android](https://play.google.com/store/apps/details?id=org.amnezia.vpn) \n	[Windows](https://github.com/amnezia-vpn/amnezia-client/releases/download/4.8.9.2/AmneziaVPN_4.8.9.2_windows_x64.exe) \n	[Linux](https://github.com/amnezia-vpn/amnezia-client/releases/download/4.8.9.2/AmneziaVPN_4.8.9.2_linux_x64.tar.zip)"

	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	return msg
}

func (m MessageCreator) HelpMenuMsg() tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(m.ChatID, "Для получения дальнейшей помощи по настройке VPN-соединения, пожалуйста, выберите вашу операционную систему:\n\n")
}

func (m MessageCreator) RefererMsg(userid string) tgbotapi.MessageConfig {
	refLink := m.BotAddress + "?start=ref" + userid

	msg := tgbotapi.NewMessage(m.ChatID, "💵Акция «Приведи друга»💵\n\nПриглашайте друзей и получайте бонусы на баланс!\n\nЗа каждого приглашенного друга вы и ваш друг получит 10 рублей на баланс для тестирования сервиса.\n\nДля участия в акции просто поделитесь своей уникальной ссылкой приглашения:\n\n`"+refLink+"`\n\nЧем больше друзей вы пригласите, тем больше бонусов получите! Акция действует без ограничений по количеству приглашенных друзей.\n\nСпасибо, что выбираете наш VPN-сервис! Вместе мы сделаем интернет безопаснее и доступнее для всех.")

	msg.ParseMode = "Markdown"

	return msg
}

func (m MessageCreator) GetInlineKeyboardMarkup(reqData string, uid int64) tgbotapi.InlineKeyboardMarkup {
	switch reqData {
	case "homePage":
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⚙️ Подключение к VPN", "vpnConnect"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💲 Внесение оплаты за VPN", "paymentMenu"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💵 Акция «Приведи друга»", "referral"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💸 Пожертвовать", "donate"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💬 Помощь", "help"),
			),
		)
		return kb
	case "paymentMenu":
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Пополнить банковской картой", "topup_fiat"),
				tgbotapi.NewInlineKeyboardButtonData("Пополнить криптовалютой", "topup_crypto"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Проверить поступления", "updateBalance"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Изменить тариф", "changeTariff"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Главное меню", "homePage"),
			),
		)
	case "vpnConnect":
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Добавить ключ", "bindKey"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Помощь с подключением", "helpMenu"),
				tgbotapi.NewInlineKeyboardButtonData("Главное меню", "homePage"),
			),
		)
	case "help":
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("iPhone/iPad", "https://surfboost.com/help/ios"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("Android", "https://surfboost.com/help/android"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("Windows", "https://surfboost.com/help/windows"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("Linux", "https://surfboost.com/help/linux"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Главное меню", "homePage"),
			),
		)
	case "referral":
		refLink := m.BotAddress + "?start=ref" + fmt.Sprintf("%d", uid)
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Главное меню", "homePage"),
				tgbotapi.NewInlineKeyboardButtonSwitch("Поделиться ссылкой", "Присоединяйся к использованию SurfBoost VPN по моей ссылке и получи бонус! 🎁\nСсылка для подключения: "+refLink),
			),
		)
	}

	return tgbotapi.NewInlineKeyboardMarkup()
}

func (m MessageCreator) SuccessTopup(sum int64, topupSum int64) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(m.ChatID, fmt.Sprintf("Баланс успешно пополнен на %d рублей. Итого: %d", topupSum, sum))
}

func (m MessageCreator) DonateMsg() tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(m.ChatID, "Если вам нравится наш VPN-сервис и вы хотите поддержать его развитие финансово, поддержка принимается по СБП на ТБанк :) ")
}

func (m MessageCreator) ThanksMsg() tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(m.ChatID, "Спасибо за регистрацию по реферальной ссылке!")
}
