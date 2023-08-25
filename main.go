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

	usersToWatchMap := make(map[int]map[int]structs.VKUser)

	usersToWatchMutex := sync.RWMutex{}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Text == "/start" {
			user := update.Message.From
			firstName := user.FirstName
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Привет, %s!\n Бот запущен!\n Команда /add ждет ID для добавления в список слежки\n Команда /del ждет ID для удаления из списка слежки\n Команда /list отобразит всех, за кем мы следим", firstName))
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
			user := update.Message.From
			userID := user.ID
			if usersToWatchMap[userID] == nil {
				usersToWatchMap[userID] = make(map[int]structs.VKUser)
			}
			if valid(update.Message.Text) {
				vkUserID, err := strconv.Atoi(strings.TrimSpace(update.Message.Text))
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Некорректный ID. Попробуй еще раз.")
					bot.Send(msg)
				} else {
					vkUserInfo, err := getapi.GetVKUserInfo(vkUserID, vkTocken)
					if err == nil && vkUserInfo.Response != nil && len(vkUserInfo.Response) > 0 {
						userToWatch := structs.VKUser{
							FirstName: vkUserInfo.Response[0].FirstName,
							LastName:  vkUserInfo.Response[0].LastName,
							ID:        vkUserID,
						}
						usersToWatchMutex.Lock()
						usersToWatchMap[userID][vkUserID] = userToWatch
						usersToWatchMutex.Unlock()
					}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Я буду следить за пользователем с ID %d\nИмя пользователя: %s\nФамилия пользователя: %s", vkUserID, usersToWatchMap[userID][vkUserID].FirstName, usersToWatchMap[userID][vkUserID].LastName))
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
			user := update.Message.From
			userID := user.ID
			if usersToWatchMap[userID] == nil {
				usersToWatchMap[userID] = make(map[int]structs.VKUser)
			}
			if valid(update.Message.Text) {
				userIDVk, err := strconv.Atoi(strings.TrimSpace(update.Message.Text))
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Некорректный ID. Попробуй еще раз.")
					bot.Send(msg)
				} else {
					if _, ok := usersToWatchMap[userID][userIDVk]; ok {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Я больше не буду следить за пользователем %s %s", usersToWatchMap[userID][userIDVk].FirstName, usersToWatchMap[userID][userIDVk].LastName))
						bot.Send(msg)
						// Удаляем пользователя из мапы
						usersToWatchMutex.Lock()
						delete(usersToWatchMap[userID], userIDVk)
						waitingForDel = false
						usersToWatchMutex.Unlock()
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Я и не следил за пользователем с ID %d", userIDVk))
						bot.Send(msg)
					}

				}
			}
		}

		if update.Message.Text == "/list" {
			user := update.Message.From
			userID := user.ID
			usersToWatchMutex.RLock() // Захватываем мьютекс для чтения мапы
			if usersToWatchMap[userID] == nil {
				usersToWatchMap[userID] = make(map[int]structs.VKUser)
			}
			var userList string
			for _, user := range usersToWatchMap[userID] {
				userList += fmt.Sprintf("ID: %d, Имя: %s, Фамилия: %s\n", user.ID, user.FirstName, user.LastName)
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

		go func(update tgbotapi.Update) {
			for {
				usersToCheck := make(map[int]structs.VKUser)

				usersToWatchMutex.RLock() // Захватываем мьютекс для чтения мапы
				user := update.Message.From
				userID := user.ID
				for id, user := range usersToWatchMap[userID] {
					usersToCheck[id] = user
				}
				usersToWatchMutex.RUnlock() // Освобождаем мьютекс

				for id, user := range usersToCheck {
					go func(userIDVk int, user structs.VKUser) {
						vkUserInfo, err := getapi.GetVKUserInfo(userIDVk, vkTocken)
						if err != nil {
							log.Printf("Error getting user info for ID %d: %v", userIDVk, err)
							return
						}

						if vkUserInfo.Response != nil && len(vkUserInfo.Response) > 0 && vkUserInfo.Response[0].Online == 1 {
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Пользователь %s %s появился в сети!", usersToWatchMap[userID][userIDVk].FirstName, usersToWatchMap[userID][userIDVk].LastName))
							bot.Send(msg)
							usersToWatchMutex.RLock()
							delete(usersToWatchMap[userID], userIDVk)
							usersToWatchMutex.RUnlock() // Освобождаем мьютекс
						}
					}(id, user)
				}

				time.Sleep(time.Minute) // Пауза между проверками
			}
		}(update)

	}
}

func valid(str string) bool {
	if str == "/del" || str == "/add" || str == "/start" || str == "/list" {
		return false
	}
	return true
}
