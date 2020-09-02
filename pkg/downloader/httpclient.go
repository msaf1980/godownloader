package downloader

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	//"github.com/PuerkitoBio/goquery"

	"github.com/calbucci/go-htmlparser"
	"github.com/franela/goreq"
	"github.com/msaf1980/godownloader/pkg/htmlutils"
)

func (d *Downloader) httpLoad(task *task) error {
	resp, err := goreq.Request{Uri: task.url, MaxRedirects: d.maxRedirects, Timeout: d.timeout}.Do()
	if err == nil {
		if resp.StatusCode == http.StatusNotFound {
			err = fmt.Errorf("Not found")
			task.try = 0
		} else if resp.StatusCode == http.StatusOK {
			if len(task.fileName) == 0 {
				c := resp.Header.Get("Content-Type")
				i := strings.Index(c, ";")
				if i > 0 {
					task.contentType = c[0:i]
				}
				d.processLock.Lock()
				err = d.genTaskFileName(task)
				d.processLock.Unlock()
				//err = fmt.Errorf("download not realized at now")
			}
		} else {
			err = fmt.Errorf("Failed with http status %d", resp.StatusCode)
		}
		// TODO: restart download
		if err == nil {
			task.size = resp.ContentLength
			if task.contentType == "text/html" {
				err = d.htmlLoad(resp.Body, task)
			} else {
				var f *os.File
				fileName := d.outdir + "/" + task.fileName
				tmpfile := fileName + ".part"
				f, err = os.OpenFile(tmpfile, os.O_RDWR|os.O_CREATE, 0644)
				if err == nil {
					_, err = io.Copy(f, resp.Body)
					if err == nil {
						if task.size == 0 {
							stat, _ := f.Stat()
							task.size = stat.Size()
						}
						err = f.Close()
					} else {
						f.Close()
					}
					if err == nil {
						err = os.Rename(tmpfile, fileName)
					}
				}
			}
			if err == nil {
				task.success = true
			}
		}
		resp.Body.Close()
	}
	return err
}

// Load the HTML document
func (d *Downloader) htmlLoad(body io.Reader, task *task) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	return d.htmlParse(data, task, true)
}

func nSpaces(n int) string {
	s := make([]rune, n)
	for i := range s {
		s[i] = ' '
	}
	return string(s)
}

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
					typ, _ := e.GetAttributeValue("type")
					rel, _ := e.GetAttributeValue("rel")
					fmt.Printf("link href='%s' type='%s' rel='%s'\n", href, typ, rel)
				}
			case "a":
				href, ok := e.GetAttributeValue("href")
				if ok {
					fmt.Printf("a href='%s'\n", href)
				}
			case "iframe", "img", "script":
				src, ok := e.GetAttributeValue("src")
				if ok {
					fmt.Printf("%s src='%s'\n", e.TagName, src)
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
