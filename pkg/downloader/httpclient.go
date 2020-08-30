package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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
			var f *os.File
			fileName := d.outdir + "/" + task.fileName
			tmpfile := fileName + ".part"
			f, err = os.OpenFile(tmpfile, os.O_RDWR|os.O_CREATE, 0644)
			if err == nil {
				task.size = resp.ContentLength
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
				if err == nil {
					task.success = true
				}
			}
		}
		resp.Body.Close()
	}
	return err
}

func (d *Downloader) htmlExtractLinks(task *task) {

}
