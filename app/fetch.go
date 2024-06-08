package app

import (
	"io"
	"net/http"
)

func (a *App) Fetch(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"`)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
