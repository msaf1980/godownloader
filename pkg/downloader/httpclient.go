package downloader

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	//"github.com/PuerkitoBio/goquery"

	"github.com/franela/goreq"
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
				d.filesLock.Lock()
				err = d._genTaskFileName(task)
				d.filesLock.Unlock()
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
