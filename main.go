package main

import (
	"fmt"
	"log"
	getapi "mytelegrambot/pcg/getApi"
	"mytelegrambot/pcg/structs"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {

	vkTocken := VkToken
	botTocken := BotToken

	bot, err := tgbotapi.NewBotAPI(botTocken)

	if err != nil {
		log.Fatal(err)
	}

	var waitingForDel, waitingForAdd bool

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	usersToWatch := make(map[int]structs.Users)
	usersToWatchMutex := sync.RWMutex{}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Text == "/start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Бот запущен!\n Команда /add ждет ID для добавления в список слежки\n Команда /del ждет ID для удаления из списка слежки\n Команда /list отобразит всех, за кем мы следим")
			bot.Send(msg)
		}

		if update.Message.Text == "/add" {
			if waitingForDel {
				waitingForDel = false
			}
			waitingForAdd = true
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введи ID человека, за кем мы будем следить")
			bot.Send(msg)
		}

		if waitingForAdd {
			if valid(update.Message.Text) {
				userID, err := strconv.Atoi(strings.TrimSpace(update.Message.Text))
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Некорректный ID. Попробуй еще раз.")
					bot.Send(msg)
				} else {
					vkUserInfo, err := getapi.GetVKUserInfo(userID, vkTocken)
					if err == nil && vkUserInfo.Response != nil && len(vkUserInfo.Response) > 0 {
						userToWatch := structs.Users{
							Name_user:     vkUserInfo.Response[0].FirstName,
							Lastname_user: vkUserInfo.Response[0].LastName,
							Status:        true,
							ChatID:        update.Message.Chat.ID,
							UserID:        userID,
						}
						usersToWatch[userID] = userToWatch
					}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Я буду следить за пользователем с ID %d\nИмя пользователя: %s\nФамилия пользователя: %s", userID, usersToWatch[userID].Name_user, usersToWatch[userID].Lastname_user))
					bot.Send(msg)
					waitingForAdd = false

				}

			}
		}

		if update.Message.Text == "/del" {
			if waitingForAdd {
				waitingForAdd = false
			}
			waitingForDel = true
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Введите ID пользователя, за которым вы хотите перестать следить:")
			bot.Send(msg)
		}

		if waitingForDel {
			if valid(update.Message.Text) {
				userID, err := strconv.Atoi(strings.TrimSpace(update.Message.Text))
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Некорректный ID. Попробуй еще раз.")
					bot.Send(msg)
				} else {

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Я больше не буду следить за пользователем %s %s", usersToWatch[userID].Name_user, usersToWatch[userID].Lastname_user))
					bot.Send(msg)
					// Удаляем пользователя из мапы
					delete(usersToWatch, userID)

					waitingForDel = false // Завершаем режим ожидания ID
				}
			}
		}

		if update.Message.Text == "/list" {
			usersToWatchMutex.RLock() // Захватываем мьютекс для чтения мапы
			var userList string
			for _, user := range usersToWatch {
				userList += fmt.Sprintf("ID: %d, Имя: %s, Фамилия: %s\n", user.UserID, user.Name_user, user.Lastname_user)
			}
			usersToWatchMutex.RUnlock() // Освобождаем мьютекс

			if userList != "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Слежу за следующими пользователями:\n"+userList)
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "В данный момент нет пользователей, за которыми ведется слежка.")
				bot.Send(msg)
			}
		}

		go func() {
			for {
				usersToCheck := make(map[int]structs.Users)

				usersToWatchMutex.RLock() // Захватываем мьютекс для чтения мапы
				for id, user := range usersToWatch {
					usersToCheck[id] = user
				}
				usersToWatchMutex.RUnlock() // Освобождаем мьютекс

				for id, user := range usersToCheck {
					go func(userID int, user structs.Users) {
						vkUserInfo, err := getapi.GetVKUserInfo(userID, vkTocken)
						if err != nil {
							log.Printf("Error getting user info for ID %d: %v", userID, err)
							return
						}

						if vkUserInfo.Response != nil && len(vkUserInfo.Response) > 0 && vkUserInfo.Response[0].Online == 1 {
							usersToWatchMutex.Lock() // Захватываем мьютекс для изменения мапы

							usersToWatchMutex.Unlock() // Освобождаем мьютекс

							msg := tgbotapi.NewMessage(user.ChatID, fmt.Sprintf("Пользователь %s %s появился в сети!", usersToWatch[userID].Name_user, usersToWatch[userID].Lastname_user))
							bot.Send(msg)
							delete(usersToWatch, userID)
						}
					}(id, user)
				}

				time.Sleep(time.Minute) // Пауза между проверками
			}
		}()

	}
}

func valid(str string) bool {
	if str == "/del" || str == "/add" || str == "/start" || str == "/list" {
		return false
	}
	return true
}
