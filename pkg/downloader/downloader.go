package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/msaf1980/godownloader/pkg/mimetypes"

	"github.com/cornelk/hashmap"
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

var (
	saveModeMap = map[string]SaveMode{"site_dir": SiteDirMode, "dir": DirMode, "flat": FlatMode, "flat_dir": FlatDirMode}
	saveModeStr = []string{"site_dir", "dir", "flat", "flat_dir"}
)

func (s *SaveMode) Set(value string) error {
	mode, ok := saveModeMap[strings.ToLower(value)]
	if ok {
		*s = mode
		return nil
	}
	return fmt.Errorf("unknown save mode: '%s'", value)
}

func (s *SaveMode) String() string {
	return saveModeStr[*s]
}

// appendFlatDir Append dir to filename in FlatDirMode
func appendFlatDir(path string, contentType string) (string, string) {
	var dir string
	switch contentType {
	case "", "text/html":
		return path, dir
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
	return dir + path, dir
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

	//processLock sync.Mutex       // set when check for existing/insert during add new task
	processed *hashmap.HashMap // lock-free map[url]*task - processed tasks by url
	filesLock sync.Mutex       // set when generate/insert new filename for task
	files     *hashmap.HashMap // lock-free map[filename]*task - processed tasks by filename

	fileMap string // map
	fMap    *os.File

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
		processed: &hashmap.HashMap{},
		files:     &hashmap.HashMap{},
		queue:     lockfree_queue.NewQueue(4096),
		//root:      list.New(),
		running: true,
	}
}

// NewLoad builder for new load
func (d *Downloader) NewLoad(dir string, fileMap string) (*Downloader, error) {
	if dir == "" {
		return nil, fmt.Errorf("output dir not set")
	}
	if d.processed.Len() == 0 {
		return nil, fmt.Errorf("root url not set")
	}
	err := os.Mkdir(dir, 0755)
	if err != nil {
		return nil, err
	}
	d.outdir = dir
	d.fileMap = dir + "/" + fileMap
	err = d.newMap()
	if err != nil {
		return nil, err
	}
	return d, nil
}

// ExistingLoad builder for new load
func (d *Downloader) ExistingLoad(dir string, fileMap string) (*Downloader, error) {
	if dir == "" {
		return nil, fmt.Errorf("output dir not set")
	}
	if d.processed.Len() == 0 {
		return nil, fmt.Errorf("root url not set")
	}
	d.outdir = dir
	d.fileMap = dir + "/" + fileMap
	err := d.openMap()
	if err != nil {
		return nil, err
	}
	return d, nil
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
	err := d.closeMap()
	if err != nil {
		d.failed = true
		log.Error().Str("where", "map").Msg(err.Error())
	}
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
				// check file in processed
				task, exist := d.addTask(t)
				if exist {
					recheck := task.UpdateLinks(t.Links(), t.DownLevel(), t.ExtLinks())
					if task.success && !recheck {
						// already downloaded
						continue
					}
				}
				// run task
				if task.TryLock() {
					d.runTask(task)
				} else {
					// Task already running, requeue
					d.queue.Put((task))
				}
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
