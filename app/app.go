package app

import (
	"net/http"

	"gitlab.com/toby3d/telegraph"
)

type App struct {
	api    *telegraph.Account
	client *http.Client
	author *Author
}

type Author struct {
	Name string
	URL  string
}

func CreateApp(api *telegraph.Account, client *http.Client, author *Author) *App {
	return &App{
		api:    api,
		client: client,
		author: author,
	}
}
