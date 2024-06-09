package app

import (
	"bytes"
	"io"
	"log"
	"net/url"
	"slices"
	"strings"

	"github.com/cixtor/readability"
	"github.com/linkosmos/sleekhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func (a *App) Parse(body io.ReadCloser, link url.URL) (*readability.Article, error) {
	read := readability.New()
	article, err := read.Parse(body, link.String())
	if err != nil {
		return nil, err
	}

	// process every node in article
	var before func(*html.Node)
	before = func(n *html.Node) {

		// trim spaces
		if n.Type == html.TextNode {
			// n.Data = strings.TrimSpace(n.Data)

			// if strings.HasPrefix(n.Data, "<") {
			// 	log.Println(n.Data)
			// }
			// if len(n.Data) > 1 {
			// 	if res := strings.TrimSpace(n.Data); res == "" {
			// 		n.Data = strings.TrimSpace(n.Data)
			// 	}

			// 	if slices.Contains([]string{"h2", "h3", "h4", "div"}, n.Parent.Data) {
			// 		n.Data = strings.TrimSpace(n.Data)
			// 	}
			// }
		}

		if n.Type == html.ElementNode {

			if n.Data == "figure" {

				log.Println(n.Data)
				n.Data = "div"
			}

			// rm avatars on medium
			if n.Data == "img" && strings.Contains(link.Host, "medium.com") {
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

			// add spaces before links
			if (n.Data == "a" || n.Data == "b" || n.Data == "i") && n.PrevSibling != nil && n.PrevSibling.Type == html.TextNode {
				if n.PrevSibling.Data != "" {
					n.PrevSibling.Data += " "
				}
			}

			if (n.Data == "a" || n.Data == "b" || n.Data == "i") && n.NextSibling != nil && n.NextSibling.Type == html.TextNode {
				if n.NextSibling.Data != "" && n.NextSibling.Data != "." && n.NextSibling.Data != "," && n.NextSibling.Data != ";" && n.NextSibling.Data != ":" {
					n.NextSibling.Data += " "
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
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			prev := c.PrevSibling
			if prev != nil && prev.Parent != nil {

				if prev.Type == html.ElementNode {
					// remove svg
					if prev.Data == "svg" {
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

					// remove source and set img source to it
					if prev.Data == "source" {
						if c.Data == "img" {
							index := slices.IndexFunc(c.Attr, func(e html.Attribute) bool {
								return e.Key == "src"
							})

							originSrc := c.Attr[index].Val

							if strings.HasPrefix(originSrc, "/") {
								c.Attr[index].Val = link.Scheme + "://" + link.Host + originSrc
							}

							if srcsetIndex := slices.IndexFunc(prev.Attr, func(e html.Attribute) bool {
								return e.Key == "srcset"
							}); srcsetIndex != -1 {
								srcset := prev.Attr[srcsetIndex].Val
								srcsetItems := strings.Split(strings.TrimSpace(srcset), " ")
								src := srcsetItems[len(srcsetItems)-2]
								if strings.HasPrefix(src, "/") {
									src = link.Scheme + "://" + link.Host + src
								}

								if src != "" {
									if index == -1 {
										c.Attr = append(c.Attr, html.Attribute{
											Key: "src",
											Val: src,
										})
									} else if strings.HasPrefix(originSrc, "data") {
										c.Attr[index] = html.Attribute{
											Key: "src",
											Val: src,
										}
									}
								}
							}
						}

						prev.Parent.RemoveChild(prev)
					}
				}
			}
			before(c)
		}
	}
	before(article.Node)

	if link.Host == "vc.ru" {

		tags := sleekhtml.NewTags()

		tags.IgnoredHTMLTags = append(tags.IgnoredHTMLTags, atom.Img)
		tags.IgnoredHTMLTags = append(tags.IgnoredHTMLTags, atom.Figure)
		output, err := sleekhtml.Sanitize(strings.NewReader(article.Content), tags)
		if err != nil {
			return nil, err
		}

		main, err := html.Parse(bytes.NewReader(output))
		if err != nil {
			return nil, err
		}

		article.Node = main
	}

	var b bytes.Buffer

	html.Render(&b, article.Node)

	log.Println("article", b.String())

	// var after func(*html.Node)
	// after = func(n *html.Node) {
	// 	if n.Type == html.ElementNode {

	// 		// add spaces before links
	// 		if (n.Data == "a" || n.Data == "b" || n.Data == "i") && n.PrevSibling != nil && n.PrevSibling.Type == html.TextNode {
	// 			if n.PrevSibling.Data != "" {
	// 				n.PrevSibling.Data += " "
	// 			}
	// 		}

	// 		if (n.Data == "a" || n.Data == "b" || n.Data == "i") && n.NextSibling != nil && n.NextSibling.Type == html.TextNode {
	// 			if n.NextSibling.Data != "" && n.NextSibling.Data != "." && n.NextSibling.Data != "," && n.NextSibling.Data != ";" && n.NextSibling.Data != ":" {
	// 				n.NextSibling.Data += " "
	// 			}
	// 		}

	// 		// fix typography formating
	// 		if n.Data == "h2" {
	// 			n.Data = "h3"
	// 		}

	// 		if n.Data == "h5" {
	// 			n.Data = "h4"
	// 		}

	// 		if n.Data == "h6" {
	// 			n.Data = "b"
	// 		}

	// 	}

	// 	for c := n.FirstChild; c != nil; c = c.NextSibling {
	// 		after(c)
	// 	}
	// }
	// after(article.Node)

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
