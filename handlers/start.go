package handlers

import (
	"gopkg.in/telebot.v3"
)

func StartCommandHandler(c telebot.Context) error {
	return c.Send("Hi! To start just send url to page you want to get InstantView", &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
	})
}
