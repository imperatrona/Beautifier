package app

import (
	"io"
	"slices"
	"strings"

	"github.com/cixtor/readability"
	"golang.org/x/net/html"
)

func (a *App) Parse(body io.ReadCloser, url string) (*readability.Article, error) {
	read := readability.New()
	article, err := read.Parse(body, url)
	if err != nil {
		return nil, err
	}

	// process every node in article
	var f func(*html.Node)
	f = func(n *html.Node) {

		// trim spaces
		if n.Type == html.TextNode {
			if len(n.Data) > 1 {
				if res := strings.TrimSpace(n.Data); res == "" {
					n.Data = strings.TrimSpace(n.Data)
				}

				if slices.Contains([]string{"h2", "h3", "h4", "div"}, n.Parent.Data) {
					n.Data = strings.TrimSpace(n.Data)
				}
			}
		}

		if n.Type == html.ElementNode {

			// rm avatars on medium
			if n.Data == "img" && strings.Contains(url, "medium.com") {
				if index := slices.IndexFunc(n.Attr, func(e html.Attribute) bool {
					return e.Key == "data-testid" && e.Val == "authorPhoto"
				}); index != -1 {
					srcKey := slices.IndexFunc(n.Attr, func(e html.Attribute) bool {
						return e.Key == "src"
					})
					if srcKey > -1 {
						n.Attr[srcKey].Val = ""
					}
				}
			}

			// fix typography formating
			if n.Data == "h2" {
				n.Data = "h3"
			}

			if n.Data == "h5" {
				n.Data = "h4"
			}

			if n.Data == "h6" {
				n.Data = "b"
			}

			// remove code elements formating (dot.dev / medium.com)
			if n.Data == "pre" {
				code := parseCode(n)

				parent := n.Parent
				parent.RemoveChild(n)
				parent.AppendChild(&html.Node{
					Type: html.TextNode,
					Data: code,
				})
			}

			// replace div[data-image-src] to img[src] (vc.ru)
			if n.Data == "div" {
				if index := slices.IndexFunc(n.Attr, func(e html.Attribute) bool {
					return e.Key == "data-image-src"
				}); index > -1 {
					src := n.Attr[index].Val

					n.AppendChild(&html.Node{
						Type:     html.ElementNode,
						DataAtom: 0,
						Data:     "img",
						Attr: []html.Attribute{
							{
								Key: "src",
								Val: src,
							},
						},
					})
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			prev := c.PrevSibling
			if prev != nil && prev.Parent != nil {
				// remove empty text
				if prev.Type == html.TextNode && len(strings.TrimSpace(prev.Data)) == 0 {
					prev.Parent.RemoveChild(prev)
				}

				if prev.Type == html.ElementNode {

					// remove svg
					if prev.Data == "svg" {
						prev.Parent.RemoveChild(prev)
					}

					// remove source
					if prev.Data == "source" {
						if c.Data == "img" {

							if index := slices.IndexFunc(c.Attr, func(e html.Attribute) bool {
								return e.Key == "src"
							}); index == -1 {
								if srcsetIndex := slices.IndexFunc(prev.Attr, func(e html.Attribute) bool {
									return e.Key == "srcset"
								}); srcsetIndex != -1 {
									srcset := prev.Attr[srcsetIndex].Val

									srcsetItems := strings.Split(srcset, " ")

									src := srcsetItems[len(srcsetItems)-2]

									if src != "" {
										c.Attr = append(c.Attr, html.Attribute{
											Key: "src",
											Val: src,
										})
									}
								}
							}
						}

						prev.Parent.RemoveChild(prev)
					}

					// remove video
					if prev.Data == "video" {
						prev.Parent.RemoveChild(prev)
					}

					// remove audio
					if prev.Data == "audio" {
						prev.Parent.RemoveChild(prev)
					}

				}
			}
			f(c)
		}
	}
	f(article.Node)

	return &article, nil
}

func parseCode(node *html.Node) string {
	code := ""
	var cd func(*html.Node)
	cd = func(e *html.Node) {

		if e.Type == html.TextNode {
			code += e.Data
		}

		for c := e.FirstChild; c != nil; c = c.NextSibling {
			cd(c)
		}
	}

	cd(node)
	return code
}
