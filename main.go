package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	// Вставь свой токен, полученный от @BotFather
	bot, err := tgbotapi.NewBotAPI("YOUR_BOT_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Text == "/start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я буду оповещать тебя, когда кто-то появится в сети.")
			bot.Send(msg)
		}

		// Обработка события появления в сети
		if update.Message.NewChatMembers != nil {
			for _, user := range *update.Message.NewChatMembers {
				if user.IsBot {
					continue
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Пользователь "+user.FirstName+" появился в сети!")
				bot.Send(msg)
			}
		}
	}
}
