package getapi

import (
	"encoding/json"
	"fmt"
	"mytelegrambot/pcg/structs"
	"net/http"
)

// GetVKUserInfo получает информацию о пользователе VK по его ID
// и токену VK. Возвращает указатель на структуру VKUserInfoResponse
// и ошибку, если что-то пошло не так.
func GetVKUserInfo(userID int, vkTocken string) (*structs.VKUserInfoResponse, error) {

	// Формирование URL для запроса к API VK
	url := fmt.Sprintf("https://api.vk.com/method/users.get?user_ids=%d&&fields=online&access_token=%s&v=5.131", userID, vkTocken)

	// Отправка GET-запроса к API VK
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Распаковка JSON-ответа в структуру VKUserInfoResponse
	var vkResponse structs.VKUserInfoResponse
	err = json.NewDecoder(response.Body).Decode(&vkResponse)
	if err != nil {
		return nil, err
	}

	return &vkResponse, nil
}
