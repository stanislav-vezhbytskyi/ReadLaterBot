package telegram

import (
	"ReadLaterBot/clients/telegram"
	"ReadLaterBot/events"
	"ReadLaterBot/storage"
	"errors"
)

type Processor struct {
	tg      *telegram.Client
	offset  int
	storage storage.Storage
}

type Meta struct {
	ChatID   int
	Username string
}

func New(client *telegram.Client, storage storage.Storage) *Processor {
	return &Processor{
		tg:      client,
		storage: storage,
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, err
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	for _, u := range updates {
		res = append(res, event(&u))
	}
	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}

func (p *Processor) Process(event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(&event)
	default:
		return errors.New("can't process message")
	}
}

func (p *Processor) processMessage(event *events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return errors.New("can't process message")
	}

	if err := p.doCmd(event.Text, meta.ChatID, meta.Username); err != nil {
		return err
	}
	return nil
}

func meta(event *events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, errors.New("can't get meta")
	}

	return res, nil
}

func event(update *telegram.Update) events.Event {
	updType := fetchType(update)

	res := events.Event{
		Type: fetchType(update),
		Text: fetchText(update),
	}
	if updType == events.Message {
		res.Meta = Meta{
			ChatID:   update.Message.Chat.ID,
			Username: update.Message.Text,
		}
	}

	return res
}

func fetchText(update *telegram.Update) string {
	if update.Message == nil {
		return ""
	}

	return update.Message.Text
}

func fetchType(update *telegram.Update) events.Type {
	if update.Message == nil {
		return events.Unknown
	}

	return events.Message
}
