package handler

import (
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/object"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
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
func (h *basicHandler) SendTextWithKeyboard(text string, peer int, keyboard *object.MessagesKeyboard) {
	request := params.NewMessagesSendBuilder()

	request.UserID(peer)
	request.RandomID(rand.Intn(math.MaxInt16))
	request.Message(text)
	request.Keyboard(keyboard)

	_, err := h.vk.MessagesSend(request.Params)
	if err != nil {
		log.Printf("ошибка при отправке сообщения пользователю(%d): %s", peer, err)
	}
}

func (h *basicHandler) SendStart(peer int, text string) {
	keyboard := object.NewMessagesKeyboard(false).AddRow()
	keyboard.AddTextButton("Мои долги", "", "primary")
	keyboard.AddTextButton("Мои должники", "", "primary")
	keyboard.AddRow()
	keyboard.AddTextButton("Статистика", "", "secondary")
	keyboard.AddRow()
	keyboard.AddTextButton("Новый долг", "", "secondary")

	request := params.NewMessagesSendBuilder().Message(text).Keyboard(keyboard).RandomID(rand.Intn(math.MaxInt16)).PeerID(peer)

	_, err := h.vk.MessagesSend(request.Params)
	if err != nil {
		log.Printf("ошибка при отправке сообщения пользователю(%d): %s", peer, err)
	}
}

func (h *basicHandler) SetValue(userID int, key string, value string) {
	request := params.NewStorageSetBuilder()
	request.UserID(userID)
	request.Key(key)
	request.Value(value)

	response, err := h.vk.StorageSet(request.Params)
	if err != nil {
		log.Printf("ошибка при смене переменной(%s) пользователя(%d): %s", key, userID, err)
	}
	if response != 1 {
		log.Printf("не удалось сменить переменную(%s) пользователя(%d)", key, userID)
	}
}

func (h *basicHandler) GetValue(userID int, key string) string {
	request := params.NewStorageGetBuilder()
	request.UserID(userID)
	request.Key(key)

	response, err := h.vk.StorageGet(request.Params)
	if err != nil {
		log.Printf("ошибка при получении переменной(%s) пользователя(%d): %s", key, userID, err)
	}
	if len(response) < 1 {
		log.Printf("ошибка при получении переменной(%s) пользователя(%d)", key, userID)
	}
	return response[0].Value
}

func (h *basicHandler) GetIdByShortName(name string) int {
	users, err := h.vk.UsersGet(params.NewUsersGetBuilder().UserIDs([]string{name}).Params)
	if err != nil {
		log.Printf("ошибка получения пользователя(%s): %s", name, err)
		return 0
	}
	if len(users) < 1 {
		log.Printf("не удалось получить пользователя пользователя(%s)", name)
		return 0
	}

	return users[0].ID
}

func (h *basicHandler) CanSendMessageTo(id int) bool {
	p := params.NewMessagesIsMessagesFromGroupAllowedBuilder().UserID(id).GroupID(h.groupID).Params

	response, err := h.vk.MessagesIsMessagesFromGroupAllowed(p)
	if err != nil {
		log.Printf("ошибка получения информации о доступе к отправке сообщений: %s", err)
		return false
	}
	return bool(response.IsAllowed)
}

func (h *basicHandler) GenDebtorsKeyboard(peer int) *object.MessagesKeyboard {
	kb := object.NewMessagesKeyboard(true)
	list := strings.Split(h.GetValue(peer, "debtors"), " ")
	i := 0
	for _, s := range list {
		id, _ := strconv.Atoi(s)
		if id == 0 {
			continue
		}
		kb.AddRow()
		kb.AddTextButton(h.notificator.GetNameById(id), s, "secondary")
		i += 1
		if i > 8 {
			break
		}
	}
	kb.AddRow()
	kb.AddTextButton("Отмена", "", "negative")
	return kb
}

func (h *basicHandler) GenSumKeyboard() *object.MessagesKeyboard {
	kb := object.NewMessagesKeyboard(true)
	kb.AddRow()
	kb.AddTextButton("50", "", "secondary")
	kb.AddTextButton("100", "", "secondary")
	kb.AddTextButton("150", "", "secondary")
	kb.AddRow()
	kb.AddTextButton("300", "", "secondary")
	kb.AddTextButton("500", "", "secondary")
	kb.AddTextButton("1000", "", "secondary")
	kb.AddRow()
	kb.AddTextButton("Отмена", "", "negative")

	return kb
}

func (h *basicHandler) GenCreateCancelKeyboard() *object.MessagesKeyboard {
	kb := object.NewMessagesKeyboard(true)
	kb.AddRow()
	kb.AddTextButton("Да", "", "secondary")
	kb.AddTextButton("Отмена", "", "negative")

	return kb
}

func (h *basicHandler) GenCloseInlineKeyboard(debtId int) *object.MessagesKeyboard {
	kb := object.NewMessagesKeyboard(false)
	kb.Inline = true
	kb.AddRow()
	kb.AddTextButton("Закрыть", debtId, "secondary")
	kb.AddTextButton("Напомнить", debtId, "secondary")

	return kb
}
