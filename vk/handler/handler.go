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

	ConfirmStart(message object.MessagesMessage)
	RejectStart(message object.MessagesMessage)

	MyDebtors(message object.MessagesMessage)
	MyDebts(message object.MessagesMessage)

	Stats(message object.MessagesMessage)

	DefaultError(message object.MessagesMessage)
}

type basicHandler struct {
	repo        repo.Debts
	vk          *api.VK
	notificator notificator.Notificator
}

func (h *basicHandler) Start(peer int) {
	h.SendStart(peer)
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
	}
	err = h.repo.SetStatus(uint(id), pkg.DebtStatusActive)
	if err != nil {
		log.Printf("ошибка подтверждения долга(%s): %s", message.Payload, err)
	}
	h.SendText(fmt.Sprintf("Долг #%d подтвержден. Отлично.", id), message.FromID)
	log.Printf("Долг #%d подтвержден.", id)
}

func (h *basicHandler) RejectStart(message object.MessagesMessage) {
	id, err := strconv.Atoi(message.Payload)
	if err != nil {
		log.Printf("ошибка конвертации id долга(%s): %s", message.Payload, err)
	}
	err = h.repo.SetStatus(uint(id), pkg.DebtStatusCanceled)
	if err != nil {
		log.Printf("ошибка отклонения долга(%s): %s", message.Payload, err)
	}
	h.SendText(fmt.Sprintf("Долг #%d отклонён. Отлично.", id), message.FromID)
	log.Printf("Долг #%d отклонён.", id)
}

func NewHandler(repo repo.Debts, vk *api.VK, notificator notificator.Notificator) Handler {
	rand.Seed(time.Now().Unix())
	return &basicHandler{repo: repo, vk: vk, notificator: notificator}
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
		h.notificator.SendNotify(debt, int(debt.LenderID))
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
		h.notificator.SendNotify(debt, int(debt.DebtorID))
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


