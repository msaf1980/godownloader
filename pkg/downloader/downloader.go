package downloader

import (
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/msaf1980/godownloader/pkg/mimetypes"

	lockfree_queue "github.com/msaf1980/go-lockfree-queue"
	"github.com/rs/zerolog/log"
)

type SaveMode int8

const (
	// SiteDirMode preserve sites/dir/file structure
	SiteDirMode SaveMode = iota
	// DirMode preserve dir structure (dir/file)
	DirMode
	// FlatMode all files in main dir
	FlatMode
	// FlatDirMode all files in main dir (but css, js and img in separate dir)
	FlatDirMode
)

// appendFlatDir Append dir to filename in FlatDirMode
func appendFlatDir(path string, contentType string) string {
	var dir string
	switch contentType {
	case "", "text/html":
		return path
	case "text/css":
		dir = "css/"
	case "text/javascript", "application/javascript":
		dir = "js/"
	default:
		if strings.HasPrefix(contentType, "image/") {
			dir = "img/"
		} else if strings.HasPrefix(contentType, "audio/") {
			dir = "audio/"
		} else if strings.HasPrefix(contentType, "video/") {
			dir = "video/"
		} else {
			dir = "download/"
		}
	}
	return dir + path
}

func replaceExtension(path string, contentType string) (string, string, string) {
	ext := filepath.Ext(path)
	if ext == "" {
		ext = mimetypes.ContentTypeDefaultFileExt(contentType)
		return path + ext, path, ext
	}
	name := path[0 : len(path)-len(ext)]
	if contentType == "text/html" && ext != ".html" && ext != ".htm" {
		ext = ".html"
	}
	return name + ext, name, ext
}

// Downloader downloader instance
type Downloader struct {
	saveMode SaveMode

	timeout time.Duration
	retry   int // URL download retry count (not for not found or simulate)

	maxRedirects int

	queue *lockfree_queue.Queue // task queue

	processLock sync.Mutex
	processed   map[string]*task // processed tasks by url
	downloads   map[string]*task // processed tasks by filename
	root        *list.List

	wg       sync.WaitGroup
	running  bool
	download int32
	failed   bool

	outdir string
}

// NewDownloader return downloader instance
func NewDownloader(saveMode SaveMode, retry int, timeout time.Duration, maxRedirects int) *Downloader {
	if retry <= 0 {
		retry = 1
	}
	return &Downloader{saveMode: saveMode, retry: retry, timeout: timeout, maxRedirects: maxRedirects,
		processed: make(map[string]*task),
		downloads: make(map[string]*task),
		queue:     lockfree_queue.NewQueue(4096),
		root:      list.New(),
		running:   true,
	}
}

// NewLoad builder for new load
func (d *Downloader) NewLoad(dir string) (*Downloader, error) {
	if dir == "" {
		return nil, fmt.Errorf("output dir not set")
	}
	if d.root.Len() == 0 {
		return nil, fmt.Errorf("root url not set")
	}
	err := os.Mkdir(dir, 0755)
	if err != nil {
		return nil, err
	}
	d.outdir = dir
	return d, nil
}

func (d *Downloader) setTaskFileName(task *task, path string) {
	if path != "" {
		task.setFileName(path)
		d.downloads[task.fileName] = task
	}
}

func (d *Downloader) taskByFileName(fileName string) *task {
	t, ok := d.downloads[fileName]
	if ok {
		return t
	}
	return nil
}

func (d *Downloader) setTask(task *task) {
	d.processed[task.url] = task
}

func (d *Downloader) taskByURL(url string) *task {
	t, ok := d.processed[url]
	if ok {
		return t
	}
	return nil
}

// AddRootURL add root url to download queue
func (d *Downloader) AddRootURL(url string, level int, extLevel int, secureAs bool, protocols *map[Protocol]bool) *Downloader {
	if level == 0 {
		return d
	}
	task := newLoadTask(url, level, extLevel, secureAs, protocols, d.retry)
	if task.protocol == Unsuppoted {
		log.Error().Str("url", task.url).Msg("protocol not supported")
	} else {
		d.processLock.Lock()
		d.root.PushBack(task)
		d.queue.Put(task)
		d.setTask(task)
		d.processLock.Unlock()
	}
	return d
}

// Abort set stop flag (but need wait for end running goroutines)
func (d *Downloader) Abort() {
	if !d.failed {
		d.failed = true
	}
	d.running = false
}

// Wait wait for complete
func (d *Downloader) Wait() bool {
	d.wg.Wait()
	return d.failed
}

// Start start downloader
func (d *Downloader) Start(parallel int) {
	if len(d.outdir) == 0 {
		d.failed = true
		log.Error().Msg("outdir not set")
		return
	}
	for i := 0; i < parallel; i++ {
		d.startN("Thread#" + strconv.Itoa(i))
	}
}

func (d *Downloader) startN(thread string) {
	go func(thread string) {
		idle := 0
		d.wg.Add(1)

		defer func() {
			if idle == 0 {
				atomic.AddInt32(&d.download, -1)
			}
			d.wg.Done()
			log.Debug().Str("Thread", thread).Msg("Exit")
		}()

		log.Debug().Str("Thread", thread).Msg("Starting")

		atomic.AddInt32(&d.download, 1)
		for d.running {
			v, ok := d.queue.Get()
			if ok {
				if idle == 1 {
					idle = 0
					atomic.AddInt32(&d.download, 1)
				}
				t := v.(*task)
				d.processLock.Lock()
				// check file in processed
				task := d.taskByURL(t.url)
				if task == nil {
					task = t
					d.setTask(task)
					d.processLock.Unlock()
				} else {
					recheck := false
					if task.level < t.level {
						task.level = t.level
						recheck = true
					}
					if task.extLevel < t.extLevel {
						task.extLevel = t.extLevel
						recheck = true
					}
					d.processLock.Unlock()
					if task.success && !recheck {
						// already downloaded
						continue
					}
				}
				// run task
				d.runTask(task)
			} else {
				if idle == 0 {
					idle = 1
					atomic.AddInt32(&d.download, -1)
				} else {
					if atomic.LoadInt32(&d.download) == 0 {
						break
					}
				}
				time.Sleep(200 * time.Millisecond)
			}
		}
	}(thread)
}
