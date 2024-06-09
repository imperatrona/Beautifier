package handlers

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/imperatrona/beautifier/app"
	"gopkg.in/telebot.v3"
)

func TextHandler(app *app.App) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		text := c.Text()
		urls := []string{}

		for _, entity := range c.Message().Entities {
			if entity.Type == telebot.EntityURL {
				urls = append(urls, text[entity.Offset:entity.Offset+entity.Length])
			}
		}

		if len(urls) < 1 {
			return c.Send("no links provided")
		}

		c.Notify(telebot.Typing)

		var response string

		for _, href := range urls {

			link, err := url.Parse(href)
			if err != nil {
				log.Println(err)
				continue
			}

			log.Println(link)

			document, err := app.Fetch(href)
			if err != nil {
				log.Println(err)
				continue
			}
			defer document.Close()

			log.Println("doc")

			article, err := app.Parse(document, *link)
			if err != nil {
				log.Println(err)
				continue
			}

			log.Println("article", article.Title)

			posts, err := app.Publish(article, *link)
			if err != nil {
				log.Println(err)
				continue
			}

			log.Println(len(posts))

			if len(posts) > 0 {
				response += fmt.Sprintf("> [%s](%s)\n", posts[0].Title, posts[0].URL)
			}
		}

		if len(strings.TrimSpace(response)) > 0 {
			return c.Reply(strings.TrimSpace(response), &telebot.SendOptions{
				ParseMode:             telebot.ModeMarkdown,
				DisableWebPagePreview: false,
			})
		}

		return nil
	}
}
