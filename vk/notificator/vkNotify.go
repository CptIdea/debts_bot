package notificator

import (
	"debts_bot/pkg"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/object"
	"log"
	"math"
	"math/rand"
	"time"
)

type vkNotificator struct {
	vk        *api.VK
	nameCache map[int]string
}

func NewVKNotificator(vk *api.VK) Notificator {
	return &vkNotificator{vk: vk, nameCache: make(map[int]string)}
}

func (v *vkNotificator) SendNotify(debt *pkg.Debt, initiator int) {
	if initiator == int(debt.DebtorID) {
		v.SendDebtMessage(debt, int(debt.LenderID))
	} else {
		v.SendDebtMessage(debt, int(debt.DebtorID))
	}
}

func (v *vkNotificator) NewStatusNotify(debt *pkg.Debt, initiator int) {
	if initiator == int(debt.DebtorID) {
		v.SendDebtUpdateMessage(debt, int(debt.LenderID))
	} else {
		v.SendDebtUpdateMessage(debt, int(debt.DebtorID))
	}
}

func (v *vkNotificator) NewDebtNotify(debt *pkg.Debt, initiator int) {
	if initiator == int(debt.DebtorID) {
		v.SendDebtConfirmationMessage(debt, int(debt.LenderID))
	} else {
		v.SendDebtConfirmationMessage(debt, int(debt.DebtorID))
	}
}

func (v *vkNotificator) ConfirmStopNotify(debt *pkg.Debt, initiator int) {
	if initiator == int(debt.DebtorID) {
		v.SendDebtConfirmationMessage(debt, int(debt.LenderID))
	} else {
		v.SendDebtConfirmationMessage(debt, int(debt.DebtorID))
	}
}

func (v *vkNotificator) GenYesNoKeyboard(debtId uint) *object.MessagesKeyboard {
	ans := object.NewMessagesKeyboard(false)
	ans.Inline = true
	ans.AddRow()
	ans.AddTextButton("Подтвердить", debtId, "positive")
	ans.AddTextButton("Отклонить", debtId, "negative")

	return ans
}

func (v *vkNotificator) GenMessageFromDebt(debt *pkg.Debt) string {
	ans := ""
	ans += fmt.Sprintf("Долг #%d\n", debt.ID)
	ans += fmt.Sprintf("Кредитор: %s\n", v.GetNameById(int(debt.LenderID)))
	ans += fmt.Sprintf("Должник: %s\n", v.GetNameById(int(debt.DebtorID)))
	ans += fmt.Sprintf("Сумма долга: %d %s\n", debt.Sum, debt.Currency)
	ans += fmt.Sprintf("Статус: %s\n", debt.Status)
	ans += fmt.Sprintf("Дата создания: %s", debt.CreatedAt.Format(time.Stamp))

	return ans
}

func (v *vkNotificator) GetNameById(id int) string {
	if len(v.nameCache) > 1000 {
		v.nameCache = make(map[int]string)
	}
	if name, ok := v.nameCache[id]; ok {
		return name
	} else {
		response, err := v.vk.UsersGet(params.NewUsersGetBuilder().UserIDs([]string{fmt.Sprint(id)}).Params)
		if err != nil {
			log.Printf("ошибка получения пользователя(%d): %s", id, err)
			return "UNKNOWN"
		}
		if len(response) < 1 {
			log.Printf("не удалось получить пользователя пользователя(%d)", id)
			return "UNKNOWN"
		}
		v.nameCache[id] = fmt.Sprintf("%s %s", response[0].FirstName, response[0].LastName)
		return v.nameCache[id]
	}
}

func (v *vkNotificator) SendDebtConfirmationMessage(debt *pkg.Debt, sendTo int) {
	request := params.NewMessagesSendBuilder()
	request.Keyboard(v.GenYesNoKeyboard(debt.ID))
	request.RandomID(rand.Intn(math.MaxInt16))
	request.Message(v.GenMessageFromDebt(debt))
	request.PeerID(sendTo)

	_, err := v.vk.MessagesSend(request.Params)
	if err != nil {
		log.Printf("ошибка при отправке сообщения пользователю(%d): %s", sendTo, err)
	}
}

func (v *vkNotificator) SendDebtUpdateMessage(debt *pkg.Debt, sendTo int) {
	request := params.NewMessagesSendBuilder()
	request.RandomID(rand.Intn(math.MaxInt16))
	request.Message(fmt.Sprintf("Обновление статуса долга\n\n%s", v.GenMessageFromDebt(debt)))
	request.PeerID(sendTo)

	_, err := v.vk.MessagesSend(request.Params)
	if err != nil {
		log.Printf("ошибка при отправке сообщения пользователю(%d): %s", sendTo, err)
	}
}

func (v *vkNotificator) SendDebtMessage(debt *pkg.Debt, sendTo int) {
	request := params.NewMessagesSendBuilder()
	request.RandomID(rand.Intn(math.MaxInt16))
	request.Message(v.GenMessageFromDebt(debt))
	request.PeerID(sendTo)

	_, err := v.vk.MessagesSend(request.Params)
	if err != nil {
		log.Printf("ошибка при отправке сообщения пользователю(%d): %s", sendTo, err)
	}
}
