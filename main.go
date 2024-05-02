package ReadLaterBot

import (
	"ReadLaterBot/clients/telegram"
	"flag"
	"log"
)

const (
	tgBotHost = "api.telegram.org"
)

func main() {
	tgClient := telegram.New(tgBotHost, mustToken())

}
func mustToken() string {
	token := flag.String("token",
		"",
		"token for access to the telegram bot")
	flag.Parse()

	if *token == "" {
		log.Fatal("token is required")
	}

	return *token
}
