package main

import (
	"ReadLaterBot/clients/telegram-client"
	"ReadLaterBot/consumer/event-consumer"
	"ReadLaterBot/events/telegram"
	"ReadLaterBot/storage/sqlite"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/context"
	"log"
)

const (
	tgBotHost   = "api.telegram.org"
	storagePath = "data/sqlite/storage.db"
	batchSize   = 100
)

func main() {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		log.Fatal("can't connect to storage: ", err)
	}

	if err := storage.Init(context.Background()); err != nil {
		log.Fatal("can't init storage: ", err)
	}

	eventsProcessor := telegram.New(
		telegram_client.New(tgBotHost, mustToken()),
		storage,
	)
	log.Printf("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	log.Printf("consumer created successfully")
	if err := consumer.Start(); err != nil {
		log.Fatal(err.Error())
	}
}

func mustToken() string {
	token := flag.String(
		"tg-bot-token",
		"",
		"token for access to telegram bot",
	)
	flag.Parse()

	if *token == "" {
		log.Fatal("token is required")
	}

	return *token
}
