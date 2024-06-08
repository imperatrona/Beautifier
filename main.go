package main

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/imperatrona/beautifier/app"
	"github.com/joho/godotenv"
	"gitlab.com/toby3d/telegraph"
	"golang.org/x/net/proxy"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	telegraphToken := os.Getenv("TELEGRAPH_TOKEN")
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	if telegraphToken == "" || telegramToken == "" {
		log.Fatal("Required env variables was not provided")
	}

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

	url := "https://medium.com/@julienetienne/stop-using-localstorage-64a6d6805da8"

	document, err := app.Fetch(url)
	if err != nil {
		log.Fatal(err)
	}
	defer document.Close()

	article, err := app.Parse(document, url)
	if err != nil {
		log.Fatal(err)
	}

	posts, err := app.Publish(article)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(posts)
}

func parseProxy(proxyUri string) (*http.Transport, error) {
	var transport *http.Transport

	if proxyUri != "" {
		proxyUrl, err := url.Parse(proxyUri)
		if err != nil {
			return nil, err
		}

		password, _ := proxyUrl.User.Password()

		p, err := proxy.SOCKS5("tcp", proxyUrl.Host, &proxy.Auth{
			User:     proxyUrl.User.Username(),
			Password: password,
		}, proxy.Direct)

		if err != nil {
			return nil, err
		}

		transport = &http.Transport{
			Dial: p.Dial,
		}
	}

	return transport, nil
}
