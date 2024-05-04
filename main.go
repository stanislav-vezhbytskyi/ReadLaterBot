package main

import (
	"ReadLaterBot/clients/telegram-client"
	"ReadLaterBot/consumer/event-consumer"
	"ReadLaterBot/events/telegram"
	"ReadLaterBot/storage/files"
	"flag"
	"log"
)

const (
	tgBotHost   = "api.telegram.org"
	storagePath = "storage"
	batchSize   = 100
)

func main() {
	eventsProcessor := telegram.New(
		telegram_client.New(tgBotHost, mustToken()),
		files.New(storagePath),
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
