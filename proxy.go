package main

import (
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

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
