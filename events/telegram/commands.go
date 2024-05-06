package telegram

import (
	"ReadLaterBot/lib/e"
	"ReadLaterBot/storage"
	"ReadLaterBot/webparser"
	_ "ReadLaterBot/webparser"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"log"
	"net/url"
	"strconv"
	"strings"
)

const (
	RndCmd      = "/rnd"
	HelpCmd     = "/help"
	StartCmd    = "/start"
	GetAllPages = "/getall"
	RemovePage  = "/rm"
)

func (p *Processor) doCmd(text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	log.Printf("got new command '%s' from '%s", text, username)

	if isAddCmd(text) {
		return p.savePage(chatID, text, username)
	}

	if isRemoveCmd(text) {
		index, err := getIndex(text)
		if err != nil {
			err := p.tg.SendMessage(chatID, "incorrect index of page")
			if err != nil {
				return err
			}
			return nil
		} else {
			return p.removePage(index-1, username)
		}
	}

	switch text {
	case RndCmd:
		return p.sendRandom(chatID, username)
	case HelpCmd:
		return p.sendHelp(chatID)
	case StartCmd:
		return p.sendHello(chatID)
	case GetAllPages:
		return p.sendAllPages(chatID, username)
	default:
		return p.tg.SendMessage(chatID, MsgUnknownCommand)
	}
}

func (p *Processor) removePage(index int, username string) error {
	return p.storage.RemoveByIndex(context.Background(), username, index)
}

func (p *Processor) sendAllPages(chatID int, username string) (err error) {
	defer func() {
		err = e.WrapIfErr("can't do command: can't send all pages", err)
	}()

	pages, err := p.storage.PickAll(context.Background(), username)
	if err != nil {
		if errors.Is(err, storage.ErrNoSavedPages) {
			return p.tg.SendMessage(chatID, "msgNoSavedPages")
		}
		return err
	}

	var messageBuilder strings.Builder

	for i, page := range pages {
		title := webparser.GetTitle(page.URL)

		messageBuilder.WriteString(fmt.Sprintf("Page %d: %s: [[link](%s)]\n", i+1, title, page.URL))
	}

	// Send the combined message at once
	return p.tg.SendMessage(chatID, messageBuilder.String())
}

func (p *Processor) savePage(chatID int, pageURL string, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: save page", err) }()

	title := webparser.GetTitle(pageURL)

	page := &storage.Page{
		URL:      pageURL,
		UserName: username,
		Title:    title,
	}

	isExists, err := p.storage.IsExists(context.Background(), page)
	if err != nil {
		return err
	}
	if isExists {
		return p.tg.SendMessage(chatID, MsgAlreadyExists)
	}

	if err := p.storage.Save(context.Background(), page); err != nil {
		return err
	}

	if err := p.tg.SendMessage(chatID, MsgSaved); err != nil {
		return err
	}

	return nil
}

func (p *Processor) sendRandom(chatID int, username string) (err error) {
	defer func() { err = e.WrapIfErr("can't do command: can't send random", err) }()

	page, err := p.storage.PickRandom(context.Background(), username)
	if err != nil && !errors.Is(err, storage.ErrNoSavedPages) {
		return err
	}
	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(chatID, MsgNoSavedPages)
	}

	title := webparser.GetTitle(page.URL)

	if err := p.tg.SendMessage(chatID, fmt.Sprintf("%s: [[link](%s)]", title, page.URL)); err != nil {
		return err
	}

	return nil
	//return p.storage.Remove(page)
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, MsgHelp)
}

func (p *Processor) sendHello(chatID int) error {
	return p.tg.SendMessage(chatID, MsgHello)
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isRemoveCmd(text string) bool {
	log.Printf(text[:len(RemovePage)])
	return text[:len(RemovePage)] == RemovePage
}

func getIndex(text string) (int, error) {
	log.Printf(text[len(RemovePage):])
	num, err := strconv.Atoi(strings.ReplaceAll(text[len(RemovePage):], " ", ""))
	if err != nil {
		return 0, fmt.Errorf("Error converting string to int: %v\n", err)
	}
	return num, nil
}

func isURL(text string) bool {
	u, err := url.Parse(text)

	return err == nil && u.Host != ""
}
