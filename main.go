package main

import (
	"debts_bot/repo"
	"debts_bot/vk"
	handler2 "debts_bot/vk/handler"
	notificator2 "debts_bot/vk/notificator"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
)

func main() {

	vkClient := api.NewVK(os.Getenv("TOKEN"))
	groupID, err := strconv.Atoi(os.Getenv("GROUP_ID"))
	if err != nil {
		log.Fatalf("ошибка получения id группы: %s", err)
	}
	adminID, _ := strconv.Atoi(os.Getenv("ADMIN_ID"))
	if adminID != 0 {
		vkClient.MessagesSend(params.NewMessagesSendBuilder().Message(fmt.Sprintf("Произошёл запуск. Сообщу если что-то пойдет не так.")).PeerID(adminID).Params)
	}

	dsn := "host=localhost user=mvp password=mvp dbname=debt_control port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("ошибка подключения базы данных: %s", err)
	}

	debtsRepo, err := repo.NewDebts(db)
	if err != nil {
		log.Fatalf("ошибка создания репозитория: %s", err)
	}

	notificator := notificator2.NewVKNotificator(vkClient)

	handler := handler2.NewHandler(debtsRepo, vkClient, notificator, groupID)

	client, err := vk.NewClient(vkClient, handler, groupID)
	if err != nil {
		if adminID != 0 {
			vkClient.MessagesSend(params.NewMessagesSendBuilder().Message(fmt.Sprintf("ОШИБКА. Ошибка создания клиента %s", err)).PeerID(adminID).Params)
		}
		log.Fatalf("ошибка создания клиента:%s", err)
	}

	err = client.Start()
	if err != nil {
		if adminID != 0 {
			vkClient.MessagesSend(params.NewMessagesSendBuilder().Message(fmt.Sprintf("ОШИБКА. Ошибка работы клиента %s", err)).PeerID(adminID).Params)
		}
		log.Fatalf("ошибка работы клиента:%s", err)
	}
}
