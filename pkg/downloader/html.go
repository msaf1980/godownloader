package downloader

import (
	"bytes"
	"io/ioutil"

	"github.com/calbucci/go-htmlparser"
	"github.com/msaf1980/godownloader/pkg/htmlutils"
	"github.com/msaf1980/godownloader/pkg/urlutils"
)

func (d *Downloader) htmlParse(data []byte, task *task, firstParse bool) error {
	changed := false

	// r, _, err := htmlutils.DecodeHTMLReader(data, "")
	// if err != nil {
	// 	return err
	// }

	// doc, err := goquery.NewDocumentFromReader(r)

	// // Change charset to utf-8
	// doc.Find("meta[charset]").Each(func(i int, s *goquery.Selection) {
	// 	charset, ok := s.Attr("charset")
	// 	if ok {
	// 		if charset != "utf-8" {
	// 			s.SetAttr("charset", "utf-8")
	// 			changed = true
	// 		}
	// 	}
	// })
	// doc.Find("meta[http-equiv]").Each(func(i int, s *goquery.Selection) {
	// 	httpEq, ok := s.Attr("http-equiv")
	// 	if ok {
	// 		if httpEq == "Content-Type" {
	// 			content, ok := s.Attr("content")
	// 			if ok && content != "text/html; charset=utf-8" {
	// 				s.SetAttr("content", "text/html; charset=utf-8")
	// 				changed = true
	// 			}
	// 		}
	// 	}
	// })

	// // Extract links
	// doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
	// 	href, ok := s.Attr("href")
	// 	if ok {
	// 		//abs, ok := s.Attr("tppabs")
	// 		fmt.Printf("a href='%s'\n", href)
	// 	}
	// })

	// doc.Find("link[href]").Each(func(i int, s *goquery.Selection) {
	// 	typ, _ := s.Attr("type")
	// 	rel, _ := s.Attr("rel")
	// 	href, ok := s.Attr("href")
	// 	//abs, ok := s.Attr("tppabs")
	// 	if ok {
	// 		fmt.Printf("link href='%s' type='%s' rel='%s'\n", href, typ, rel)
	// 	}
	// })

	// if changed || firstParse {
	// 	html, err := goquery.OuterHtml(doc.Selection)
	// 	if err == nil {
	// 		err = ioutil.WriteFile(d.outdir+"/"+task.fileName, []byte(html), 0644)
	// 	}
	// }

	baseHost, basePath := urlutils.SplitURL(task.url)
	baseDir := urlutils.BaseURLDir(basePath)
	baseLevel := task.Level()
	baseDownLevel := task.DownLevel()
	baseExtLevel := task.DownLevel()

	var newHTML bytes.Buffer
	r, _, err := htmlutils.DecodeHTMLBytes(data, "")
	if err != nil {
		return err
	}
	parser := htmlparser.NewParser(r)
	parser.PreserveCRLFTab = false
	tags := 0
	indent := 2
	ended := false
	parser.Parse(
		func(text string, parent *htmlparser.HtmlElement) {
			newHTML.WriteString(text)
		},
		func(e *htmlparser.HtmlElement, isEmpty bool) {
			switch e.TagName {
			case "meta":
				charset, ok := e.GetAttributeValue("charset")
				if ok {
					if charset != "utf-8" {
						e.SetAttribute("charset", "utf-8")
						changed = true
					}
				} else {
					httpEquiv, _ := e.GetAttributeValue("http-equiv")
					if httpEquiv == "Content-Type" {
						content, ok := e.GetAttributeValue("content")
						if ok && content != "text/html; charset=utf-8" {
							e.SetAttribute("content", "text/html; charset=utf-8")
							changed = true
						}
					}
				}
			case "link":
				href, ok := e.GetAttributeValue("href")
				if ok {
					if len(href) > 0 && href[0] != '#' {
						//typ, _ := e.GetAttributeValue("type")
						rel, _ := e.GetAttributeValue("rel")
						needLoad := false
						// if typ == "application/rss+xml" || rel == "alternate" || rel == "search" || rel == "canonical" || rel="amphtml" {
						// 	needLoad = false
						// }
						if rel == "stylesheet" || rel == "preload" || rel == "image_src" ||
							rel == "shortcut icon" || rel == "apple-touch-icon" || rel == "icon" {
							needLoad = true
						}
						var absURL string
						if firstParse {
							absURL = urlutils.AbsURL(href, baseHost)
							e.SetAttribute("tppabs", absURL)
						} else {
							absURL, ok = e.GetAttributeValue("tppabs")
							if !ok {
								absURL = urlutils.AbsURL(href, baseHost)
							}
						}
						if !needLoad || !d.addURL(absURL, true, d.retry, baseHost, baseDir, baseLevel, baseDownLevel, baseExtLevel) {
							e.SetAttribute("href", absURL)
						}
					}
					//fmt.Printf("link href='%s' type='%s' rel='%s'\n", absURL, typ, rel)
				}
			case "a":
				href, ok := e.GetAttributeValue("href")
				if ok {
					if len(href) > 0 && href[0] != '#' {
						var absURL string
						if firstParse {
							absURL = urlutils.AbsURL(href, baseHost)
							e.SetAttribute("tppabs", absURL)
						} else {
							absURL, ok = e.GetAttributeValue("tppabs")
							if !ok {
								absURL = urlutils.AbsURL(href, baseHost)
							}
						}
						if !d.addURL(absURL, false, d.retry, baseHost, baseDir, baseLevel, baseDownLevel, baseExtLevel) {
							e.SetAttribute("href", absURL)
						}
					}
					//fmt.Printf("a href='%s'\n", absURL)
				}
			case "iframe", "img", "script":
				pageContent := true
				if e.TagName == "iframe" {
					pageContent = false
				}
				src, ok := e.GetAttributeValue("src")
				if ok {
					if len(src) > 0 && src[0] != '#' {
						var absURL string
						if firstParse {
							absURL = urlutils.AbsURL(src, baseHost)
							e.SetAttribute("tppabs", absURL)
						} else {
							absURL, ok = e.GetAttributeValue("tppabs")
							if !ok {
								absURL = urlutils.AbsURL(src, baseHost)
							}
						}
						if !d.addURL(absURL, pageContent, d.retry, baseHost, baseDir, baseLevel, baseDownLevel, baseExtLevel) {
							e.SetAttribute("src", absURL)
						}
					}
					//fmt.Printf("%s src='%s'\n", e.TagName, absURL)
				}
			}
			if tags > 0 {
				if e.TagName != "br" {
					newHTML.WriteRune('\n')
					newHTML.WriteString(nSpaces(tags * indent))
				}
			}
			newHTML.WriteString(e.GetOpenTag(false, false))
			if e.ElementInfo.TagFormatting == htmlparser.HTFSingle {
				ended = true
			} else {
				ended = false
				tags++
			}
		},
		func(tag string) {
			tags--
			if ended {
				newHTML.WriteRune('\n')
				newHTML.WriteString(nSpaces(tags * indent))
			}
			newHTML.WriteString("</" + tag + ">")
			ended = true
		},
	)
	newHTML.WriteRune('\n')

	if changed || firstParse {
		err = ioutil.WriteFile(d.outdir+"/"+task.fileName, newHTML.Bytes(), 0644)
	}

	return err
}
