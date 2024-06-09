package app

import (
	"bytes"
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/cixtor/readability"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gitlab.com/toby3d/telegraph"
	"golang.org/x/net/html"
)

func (a *App) Publish(article *readability.Article, source url.URL) ([]*telegraph.Page, error) {
	var buf bytes.Buffer
	html.Render(&buf, article.Node)

	id, err := gonanoid.Generate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 5)
	if err != nil {
		return nil, err
	}

	max := 20000
	text := buf.String()
	var chunks []string

	if len(buf.String()) > max {
		totalLength := len(text)
		totalChunks := int(math.Ceil(float64(totalLength) / float64(max)))

		lastCut := 0

		for i := range totalChunks {
			cut := max * (i + 1)
			if i+1 == totalChunks {
				cut = totalLength
			}
			slice := text[cut-5000 : cut]
			index := strings.LastIndex(slice, "</p>")
			nextCut := cut - 5000 + index + 4
			chunks = append(chunks, text[lastCut:nextCut])
			lastCut = nextCut
		}
	} else {
		chunks = append(chunks, text)
	}

	var pages []*telegraph.Page

	for i, chunk := range chunks {
		content, err := telegraph.ContentFormat(chunk)
		if err != nil {
			return nil, err
		}

		page, err := a.api.CreatePage(telegraph.Page{
			Title:       id,
			Description: article.Excerpt,
			ImageURL:    article.Image,
			Content:     content,
			AuthorName:  a.author.Name,
			AuthorURL:   a.author.URL,
		}, false)

		if err != nil {
			return nil, err
		}

		pages = append(pages, page)
		fmt.Println(i+1, page.URL)
	}

	// add pagination, and meta links
	var resultPages []*telegraph.Page
	for i, page := range pages {
		nextPage := true
		if i+1 == len(pages) {
			nextPage = false
		}

		var chunk string

		if i > 0 {
			chunk += fmt.Sprintf(`<p><a href="%s">« Previous page</a></p>`, pages[i-1].URL)
		}

		chunk += chunks[i]

		if i+1 == len(pages) {
			chunk += fmt.Sprintf(`<p><i>Originally published at <a href="%s">%s</a></i></p>`, source.String(), source.Host)
		}
		if nextPage || i > 0 {
			if nextPage {
				chunk += fmt.Sprintf(`<p><a href="%s">Next page »</a></p>`, pages[i+1].URL)
			}
		}

		content, err := telegraph.ContentFormat(chunk)
		if err != nil {
			return nil, err
		}

		newPage, err := a.api.EditPage(telegraph.Page{
			Title:       article.Title,
			Description: article.Excerpt,
			ImageURL:    article.Image,
			Content:     content,
			Path:        page.Path,
			AuthorName:  a.author.Name,
			AuthorURL:   a.author.URL,
		}, false)

		if pages[i] != nil {
			resultPages = append(resultPages, newPage)
		}

		if err != nil {
			return nil, err
		}
	}

	return resultPages, nil
}
