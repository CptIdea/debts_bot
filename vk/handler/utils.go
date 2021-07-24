package handler

import (
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/object"
	"log"
	"math"
	"math/rand"
)

func (h *basicHandler) GetPage(userID int) string {
	request := params.NewStorageGetBuilder()
	request.Key("page")
	request.UserID(userID)

	response, err := h.vk.StorageGet(request.Params)
	if err != nil {
		log.Printf("получение page пользователя(%d): %s", userID, err)
	}
	if len(response) < 1 {
		log.Printf("не удалось получить page пользователя(%d)", userID)
	}

	return response[0].Value

}

func (h *basicHandler) SetPage(userID int, page string) {
	request := params.NewStorageSetBuilder()
	request.UserID(userID)
	request.Key("page")
	request.Value(page)

	response, err := h.vk.StorageSet(request.Params)
	if err != nil {
		log.Printf("ошибка при смены page пользователя(%d): %s", userID, err)
	}
	if response != 1 {
		log.Printf("не удалось сменить page пользователя(%d)", userID)
	}
}

func (h *basicHandler) SendText(text string, peer int) {
	request := params.NewMessagesSendBuilder()

	request.UserID(peer)
	request.RandomID(rand.Intn(math.MaxInt16))
	request.Message(text)

	_, err := h.vk.MessagesSend(request.Params)
	if err != nil {
		log.Printf("ошибка при отправке сообщения пользователю(%d): %s", peer, err)
	}
}

func (h *basicHandler) SendStart(peer int) {
	keyboard := object.NewMessagesKeyboard(false).AddRow()
	keyboard.AddTextButton("Мои долги", "", "primary")
	keyboard.AddTextButton("Мои должники", "", "primary")
	keyboard.AddRow()
	keyboard.AddTextButton("Статистика", "", "secondary")

	request := params.NewMessagesSendBuilder().Message("Привет! Бот нужен только для оповещений. Скорее всего ты здесь потому что кто-то хочет дать тебе в долг. Если у тебя есть заявки, то сейчас их отправлю.").Keyboard(keyboard).RandomID(rand.Intn(math.MaxInt16)).PeerID(peer)

	_, err := h.vk.MessagesSend(request.Params)
	if err != nil {
		log.Printf("ошибка при отправке сообщения пользователю(%d): %s", peer, err)
	}
}
