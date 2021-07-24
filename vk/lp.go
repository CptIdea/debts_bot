package vk

import (
	"context"
	handler2 "debts_bot/vk/handler"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"
	"log"
	"math/rand"
	"time"
)

type Client struct {
	vk        *api.VK
	lp        *longpoll.LongPoll
	handler   handler2.Handler
	nameCache map[int]string
	groupID   int
}

func NewClient(vk *api.VK, handler handler2.Handler, groupID int) (*Client, error) {
	rand.Seed(time.Now().Unix())

	// Инициализируем longpoll
	lp, err := longpoll.NewLongPoll(vk, groupID)
	if err != nil {
		return nil, err
	}

	c := &Client{lp: lp, vk: vk, handler: handler, nameCache: make(map[int]string), groupID: groupID}

	// Событие нового сообщения
	lp.MessageNew(func(_ context.Context, obj events.MessageNewObject) {
		log.Printf("%d: %s, %s", obj.Message.PeerID, obj.Message.Text, obj.Message.Payload)

		switch obj.Message.Text {
		case "Начать", "Start":
			go c.handler.Start(obj.Message.FromID)

		case "Подтвердить":
			go c.handler.ConfirmStart(obj.Message)

		case "Отклонить":
			go c.handler.RejectStart(obj.Message)

		case "Мои долги":
			go c.handler.MyDebts(obj.Message)

		case "Мои должники":
			go c.handler.MyDebtors(obj.Message)

		case "Статистика":
			go c.handler.Stats(obj.Message)

		case "Отмена", "Назад", "Меню":
			go c.handler.Cancel(obj.Message.FromID)

		case "Новый долг":
			go c.handler.StartNewDebt(obj.Message)

		case "Закрыть":
			go c.handler.CloseDebt(obj.Message)

		default:
			go c.handler.HandleWithPage(obj.Message)
		}
	})

	return c, nil
}

func (c *Client) Start() error {
	// Запускаем Bots Longpoll
	log.Println("Start longpoll")
	if err := c.lp.Run(); err != nil {
		return err
	}
	return nil
}
