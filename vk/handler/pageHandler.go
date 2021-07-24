package handler

import (
	"debts_bot/pkg"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/object"
	"log"
	"strconv"
	"strings"
)

func (h *basicHandler) HandleWithPage(message object.MessagesMessage) {
	switch h.GetPage(message.FromID) {
	case "setDebtor":
		h.SetDebtor(message)
	case "setSum":
		h.SetSum(message)
	case "createDebt":
		h.ConfirmNewDebt(message)
	}
}

// StartNewDebt запускает создание нового долга
func (h *basicHandler) StartNewDebt(message object.MessagesMessage) {
	h.SetPage(message.FromID, "setDebtor")
	h.SendTextWithKeyboard("Отправь ссылку на пользователя, которому хочешь занять", message.FromID, h.GenDebtorsKeyboard(message.FromID))
	// todo клавиатура с предыдущими должниками
}

// SetDebtor устанавливает должника при создании нового долга
func (h *basicHandler) SetDebtor(message object.MessagesMessage) {
	var user string
	if message.Payload != "" {
		user = strings.ReplaceAll(message.Payload, "\"", "")
	} else {
		path := strings.Split(message.Text, "/")
		if len(path) < 2 {
			log.Printf("ошибка парсинга ссылки на пользователя")
			h.SendText("Не получилось найти пользователя, попробуй ещё раз", message.FromID)
			return
		}
		user = strings.TrimPrefix(path[len(path)-1], "id")
	}

	id, err := strconv.Atoi(user)
	if err != nil {
		id = h.GetIdByShortName(user)
		user = fmt.Sprint(id)
	}

	h.SetValue(message.FromID, "debtorID", user)
	debtors := h.GetValue(message.FromID, "debtors")
	if !strings.Contains(debtors, user) {
		h.SetValue(message.FromID, "debtors", fmt.Sprintf("%s %s", user, debtors))
	}

	h.SendTextWithKeyboard(fmt.Sprintf("Должник - %s\nТеперь установим сумму", h.notificator.GetNameById(id)), message.FromID, h.GenSumKeyboard())
	h.SetPage(message.FromID, "setSum")
}

// SetSum устанаваливает сумму при создании нового долга
func (h *basicHandler) SetSum(message object.MessagesMessage) {
	sum, err := strconv.Atoi(strings.Split(message.Text, " ")[0])
	if err != nil {
		log.Printf("ошибка парсинга суммы")
		h.SendText("Не получилось определить сумму, попробуй ещё раз", message.FromID)
		return
	}
	h.SendTextWithKeyboard(fmt.Sprintf("Сумма - %d\nСоздаём?", sum), message.FromID, h.GenCreateCancelKeyboard())
	h.SetValue(message.FromID, "sum", fmt.Sprint(sum))
	h.SetPage(message.FromID, "createDebt")
}

func (h *basicHandler) ConfirmNewDebt(message object.MessagesMessage) {
	switch strings.ToLower(message.Text) {
	case "да", "создать":
		debtorID, err := strconv.Atoi(h.GetValue(message.FromID, "debtorID"))
		if err != nil {
			log.Printf("ошибка парсинга должника")
			h.DefaultError(message)
			return
		}
		sum, err := strconv.Atoi(h.GetValue(message.FromID, "sum"))
		if err != nil {
			log.Printf("ошибка парсинга суммы")
			h.DefaultError(message)
			return
		}
		newDebt := &pkg.Debt{
			LenderID: int64(message.FromID),
			DebtorID: int64(debtorID),
			AuthorID: int64(message.FromID),
			Sum:      int64(sum),
		}
		err = h.repo.Save(newDebt)
		if err != nil {
			h.DefaultError(message)
			log.Printf("ошибка сохранения долга: %s", err)
		}
		h.notificator.NewDebtNotify(newDebt, message.FromID)
		if !h.CanSendMessageTo(debtorID) {
			h.SendStart(message.FromID, "У меня нет доступа к отправке сообщений должнику.\nПерешли ему это сообщение и попроси его написать мне.\nhttps://vk.me/debt_control")
		}
		h.SendStart(message.FromID, "Отправлен запрос на создание долга.")
	default:
		h.Cancel(message.FromID)
	}
}
