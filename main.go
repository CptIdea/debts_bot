package main

import (
	"debts_bot/repo"
	"debts_bot/vk"
	handler2 "debts_bot/vk/handler"
	notificator2 "debts_bot/vk/notificator"
	"github.com/SevereCloud/vksdk/v2/api"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

func main() {

	dsn := "host=localhost user=mvp password=mvp dbname=debt_control port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("ошибка подключения базы данных: %s", err)
	}

	debtsRepo, err := repo.NewDebts(db)
	if err != nil {
		log.Fatalf("ошибка создания репозитория: %s", err)
	}

	vkClient := api.NewVK(os.Getenv("TOKEN"))

	notificator := notificator2.NewVKNotificator(vkClient)

	handler := handler2.NewHandler(debtsRepo, vkClient,notificator)



	client, err := vk.NewClient(vkClient, handler)
	if err != nil {
		log.Fatalf("ошибка создания клиента:%s", err)
	}

	err = client.Start()
	if err != nil {
		log.Fatalf("ошибка работы клиента:%s", err)
	}
}
