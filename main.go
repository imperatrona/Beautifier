package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imperatrona/beautifier/app"
	"github.com/imperatrona/beautifier/handlers"
	"github.com/joho/godotenv"
	"gitlab.com/toby3d/telegraph"

	"gopkg.in/telebot.v3"
)

func main() {

	// parse env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	telegraphToken := os.Getenv("TELEGRAPH_TOKEN")
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	if telegraphToken == "" || telegramToken == "" {
		log.Fatal("Required env variables was not provided")
	}

	// init app

	transport, err := parseProxy(os.Getenv("PROXY"))
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}

	if transport != nil {
		client.Transport = transport
	}

	app := app.CreateApp(
		&telegraph.Account{
			AccessToken: telegraphToken,
		},
		client,
		&app.Author{
			Name: "Beautifier Reborn",
			URL:  "https://t.me/beautifierrebornbot",
		},
	)

	// start bot

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  telegramToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	bot.Handle("/start", handlers.StartCommandHandler)
	bot.Handle("/help", handlers.HelpCommandHandler)

	bot.Handle(telebot.OnText, handlers.TextHandler(app))

	go func() {
		bot.Start()
	}()

	log.Println("Server started...")

	// catch exit
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	bot.Stop()
	log.Println("Server stopped...")

}
