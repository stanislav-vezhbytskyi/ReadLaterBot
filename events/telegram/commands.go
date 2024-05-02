package telegram

import (
	"ReadLaterBot/storage"
	"errors"
	"log"
	"net/url"
	"strings"
)

const (
	RndCmd   = "/rnd"
	HelpCmd  = "/help"
	StartCmd = "/start"
)

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new command %s from %s", text, username)

	if isAddCmd(text) {
		p.savePage(chatID, text, username)
	}

	switch text {
	case RndCmd:

		return p.sendRnd(chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID, username)
	case StartCmd:
		return p.sendHello(chatID, username)

	default:
		return p.tg.SendMessage(chatID, "sorry, I don't now this command")

	}
}

func (p *Processor) savePage(chatID int, pageURL string, username string) error {
	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
	}

	isExist, err := p.storage.IsExists(page)
	if err != nil {
		return err
	}
	if isExist {
		return p.tg.SendMessage(chatID, "this page already exists")
	}
	if err := p.storage.Save(page); err != nil {
		return err
	}
	return p.tg.SendMessage(chatID, "page saved successfully")
}

func (p *Processor) sendRnd(chatID int, username string) error {
	rndPage, err := p.storage.PickRandom(username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}

	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatID, storage.ErrNoSavedPages.Error())
	}

	return p.tg.SendMessage(chatID, rndPage.URL)

}

func (p *Processor) sendHelp(chatID int, username string) error {
	return p.tg.SendMessage(chatID, "I'm too lazy to write it now, sorry :)")
}

func (p *Processor) sendHello(chatID int, username string) error {
	return p.tg.SendMessage(chatID, "I'm too lazy to write it now, sorry :)")
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}
