package handler

import (
	"debts_bot/pkg"
	"debts_bot/repo"
	"debts_bot/vk/notificator"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/object"
	"log"
	"math/rand"
	"strconv"
	"time"
)

type Handler interface {
	Start(peer int)
	Cancel(peer int)

	ConfirmStart(message object.MessagesMessage)
	RejectStart(message object.MessagesMessage)

	MyDebtors(message object.MessagesMessage)
	MyDebts(message object.MessagesMessage)

	Stats(message object.MessagesMessage)

	DefaultError(message object.MessagesMessage)

	StartNewDebt(message object.MessagesMessage)
	SetDebtor(message object.MessagesMessage)
	SetSum(message object.MessagesMessage)
	ConfirmNewDebt(message object.MessagesMessage)

	HandleWithPage(message object.MessagesMessage)

	CloseDebt(message object.MessagesMessage)

	DebtNotify(message object.MessagesMessage)
}

type basicHandler struct {
	repo        repo.Debts
	vk          *api.VK
	notificator notificator.Notificator
	groupID     int
}

func (h *basicHandler) Start(peer int) {
	h.SendStart(peer, "Привет! Если у тебя есть заявки, то сейчас их отправлю.")
	time.Sleep(3 * time.Second)
	h.SetPage(peer, "start")
	debts, err := h.repo.GetListByDebtorID(uint(peer))
	if err != nil {
		log.Printf("ошибка получения долгов из базы: %s", err)
	}
	for _, debt := range debts {
		if debt.Status != pkg.DebtStatusStartWaiting {
			continue
		}
		var initiator int
		if debt.LenderID == int64(peer) {
			initiator = int(debt.DebtorID)
		} else {
			initiator = int(debt.LenderID)
		}
		h.notificator.NewDebtNotify(debt, initiator)
	}
}

func (h *basicHandler) ConfirmStart(message object.MessagesMessage) {
	id, err := strconv.Atoi(message.Payload)
	if err != nil {
		log.Printf("ошибка конвертации id долга(%s): %s", message.Payload, err)
		return
	}
	err = h.repo.SetStatus(uint(id), pkg.DebtStatusActive)
	if err != nil {
		log.Printf("ошибка подтверждения долга(%s): %s", message.Payload, err)
		return
	}
	h.SendText(fmt.Sprintf("Долг #%d подтвержден. Отлично.", id), message.FromID)
	debt, err := h.repo.GetByDebtID(uint(id))
	if err != nil {
		log.Printf("ошибка получения долга из базы: %s", err)
		return
	}
	h.notificator.NewStatusNotify(debt, message.FromID)
	log.Printf("Долг #%d подтвержден.", id)
}

func (h *basicHandler) RejectStart(message object.MessagesMessage) {
	id, err := strconv.Atoi(message.Payload)
	if err != nil {
		log.Printf("ошибка конвертации id долга(%s): %s", message.Payload, err)
		return
	}
	err = h.repo.SetStatus(uint(id), pkg.DebtStatusCanceled)
	if err != nil {
		log.Printf("ошибка отклонения долга(%s): %s", message.Payload, err)
		return
	}
	h.SendText(fmt.Sprintf("Долг #%d отклонён. Отлично.", id), message.FromID)
	debt, err := h.repo.GetByDebtID(uint(id))
	if err != nil {
		log.Printf("ошибка получения долга из базы: %s", err)
		return
	}
	h.notificator.NewStatusNotify(debt, message.FromID)
	log.Printf("Долг #%d отклонён.", id)
}

func (h *basicHandler) DefaultError(message object.MessagesMessage) {
	h.SendText("Произошла какая-то мистическая ошибка при обработке запроса.", message.FromID)
	h.SetPage(message.FromID, "start")
}

func (h *basicHandler) MyDebts(message object.MessagesMessage) {
	debts, err := h.repo.GetActiveListByDebtorID(uint(message.FromID))
	if err != nil {
		log.Printf("ошибка получения долгов из базы: %s", err)
	}
	if len(debts) > 10 {
		h.SendText("У тебе очень много долгов, покажу только первые десять", message.FromID)
		debts = debts[:9]
	}
	for _, debt := range debts {
		h.SendTextWithKeyboard(h.notificator.GenMessageFromDebt(debt), message.FromID, h.GenCloseInlineKeyboard(int(debt.ID)))
	}
	if len(debts) == 0 {
		h.SendText("Тут пока ничего нет", message.FromID)
	}
}

func (h *basicHandler) MyDebtors(message object.MessagesMessage) {
	debts, err := h.repo.GetActiveListByLenderID(uint(message.FromID))
	if err != nil {
		log.Printf("ошибка получения долгов из базы: %s", err)
	}
	if len(debts) > 10 {
		h.SendText("У тебе очень много должников, покажу только первые десять", message.FromID)
		debts = debts[:9]
	}
	for _, debt := range debts {
		h.SendTextWithKeyboard(h.notificator.GenMessageFromDebt(debt), message.FromID, h.GenCloseInlineKeyboard(int(debt.ID)))
	}
	if len(debts) == 0 {
		h.SendText("Тут пока ничего нет", message.FromID)
	}
}

func (h *basicHandler) Stats(message object.MessagesMessage) {
	debts, err := h.repo.GetActiveListByLenderID(uint(message.FromID))
	if err != nil {
		log.Printf("ошибка получения долгов из базы: %s", err)
	}
	var plus int64
	for _, debt := range debts {
		plus += debt.Sum
	}

	credits, err := h.repo.GetActiveListByDebtorID(uint(message.FromID))
	if err != nil {
		log.Printf("ошибка получения долгов из базы: %s", err)
	}
	var minus int64
	for _, debt := range credits {
		minus += debt.Sum
	}

	h.SendText(fmt.Sprintf("Ты должен: %d\nТебе должны: %d", minus, plus), message.FromID)
}

func (h *basicHandler) Cancel(peer int) {
	h.SetPage(peer, "start")
	h.SendStart(peer, "Отмена.")
}

func (h *basicHandler) CloseDebt(message object.MessagesMessage) {
	debtID, err := strconv.Atoi(message.Payload)
	if err != nil {
		log.Printf("ошибка парсинга ид долга")
		h.DefaultError(message)
		return
	}

	debt, err := h.repo.GetByDebtID(uint(debtID))
	if err != nil {
		log.Printf("ошибка получения долга из базы: %s", err)
	}

	if debt.LenderID == int64(message.FromID) {
		err := h.repo.SetStatus(debt.ID, pkg.DebtStatusClosed)
		if err != nil {
			log.Printf("ошибка смены статуса: %s", err)
		}
		debt.Status = pkg.DebtStatusClosed
		h.notificator.SendNotify(debt, message.FromID)
		h.SendStart(message.FromID, fmt.Sprintf("Долг #%d закрыт", debtID))
	} else {
		debt.Status = pkg.DebtStatusStopWaiting
		h.SendTextWithKeyboard(fmt.Sprintf("Должник предлагает закрыть долг\n\n %s", h.notificator.GenMessageFromDebt(debt)), int(debt.LenderID), h.GenCloseInlineKeyboard(debtID))
		h.SendText(fmt.Sprintf("Кредитору отправлен запрос на закрытие долга #%d", debtID), message.FromID)
	}

}

func (h *basicHandler) DebtNotify(message object.MessagesMessage) {
	id, err := strconv.Atoi(message.Payload)
	if err != nil {
		log.Printf("ошибка конвертации id долга(%s): %s", message.Payload, err)
		h.DefaultError(message)
		return
	}
	debt, err := h.repo.GetByDebtID(uint(id))
	if err != nil {
		log.Printf("ошибка получения долга(%s): %s", message.Payload, err)
		h.DefaultError(message)
		return
	}
	if debt.LastNotify.Sub(time.Now()) < 30*time.Minute {
		h.SendText("Напоминать можно раз в 30 минут", message.FromID)
	}

	debt.LastNotify = time.Now()
	err = h.repo.Save(debt)
	if err != nil {
		log.Printf("ошибка сохранения долга(%s): %s", message.Payload, err)
	}

	h.SendText("Напомнили о долге", message.FromID)
	h.notificator.SendNotify(debt, message.FromID)
}

func NewHandler(repo repo.Debts, vk *api.VK, notificator notificator.Notificator, groupID int) Handler {
	rand.Seed(time.Now().Unix())
	return &basicHandler{repo: repo, vk: vk, notificator: notificator, groupID: groupID}
}
