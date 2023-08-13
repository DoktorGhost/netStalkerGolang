package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type users struct {
	name_user     string
	lastname_user string
	status        bool
}

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
	usersToWatch := make(map[int]users)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Text == "/start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Бот запущен!\n Команда /add ждет ID для добавления в список слежки\n Команда /del ждет ID для удаления из списка слежки")
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
					vkUserInfo, err := getVKUserInfo(userID, vkTocken)
					if err == nil && vkUserInfo.Response != nil && len(vkUserInfo.Response) > 0 {
						userToWatch := users{
							name_user:     vkUserInfo.Response[0].FirstName,
							lastname_user: vkUserInfo.Response[0].LastName,
							status:        true,
						}
						usersToWatch[userID] = userToWatch
					}
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Я буду следить за пользователем с ID %d\nИмя пользователя: %s\nФамилия пользователя: %s", userID, usersToWatch[userID].name_user, usersToWatch[userID].lastname_user))
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
					// Удаляем пользователя из мапы
					delete(usersToWatch, userID)

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Я больше не буду следить за пользователем с ID %d", userID))
					bot.Send(msg)

					waitingForDel = false // Завершаем режим ожидания ID
				}
			}
		}

	}
}

func getVKUserInfo(userID int, vkTocken string) (*VKUserInfoResponse, error) {

	url := fmt.Sprintf("https://api.vk.com/method/users.get?user_ids=%d&&fields=online&access_token=%s&v=5.131", userID, vkTocken)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var vkResponse VKUserInfoResponse
	err = json.NewDecoder(response.Body).Decode(&vkResponse)
	if err != nil {
		return nil, err
	}

	return &vkResponse, nil
}

type VKUserInfoResponse struct {
	Response []VKUser `json:"response"`
}

type VKUser struct {
	ID        int    `json:"id"`
	Online    int    `json:"online"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func valid(str string) bool {
	if str == "/del" || str == "/add" || str == "/start" {
		return false
	}
	return true
}
