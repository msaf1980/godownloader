package downloader

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/msaf1980/godownloader/pkg/urlutils"

	"github.com/goware/urlx"
	"github.com/rs/zerolog/log"
)

// Proto
type Protocol int8

const (
	Unsuppoted Protocol = iota
	HTTP
)

func StringToProto(s string) Protocol {
	switch s {
	case "http", "https":
		return HTTP
	default:
		return Unsuppoted
	}
}

func URLProtocol(url string) Protocol {
	i := strings.Index(url, "://")
	if i > 0 {
		return StringToProto(url[0:i])
	}
	return Unsuppoted
}

type task struct {
	url       string
	protocol  Protocol
	level     int32 // download level (from same site)
	extLevel  int32 // download level (from external sites)
	secureAs  bool  // cast secure and unsecure links to one site (no external)
	protocols *map[Protocol]bool

	fileName    string // relative filename (blank if no try downloads else)
	contentType string
	success     bool
	size        int64 // size from header
	try         int   // retry count - stop on 0 or success

	lock      uint32     // set 1 for hold task during download/parse
	lockLevel sync.Mutex // set 1 for hold task during level
}

func (task *task) TryLock() bool {
	if atomic.CompareAndSwapUint32(&task.lock, 0, 1) {
		return true
	}
	return false
}

func (task *task) UnLock() {
	atomic.CompareAndSwapUint32(&task.lock, 1, 0)
}

func (task *task) UpdateLevel(level int32, extLevel int32) bool {
	task.lockLevel.Lock()
	changed := false
	if atomic.LoadInt32(&task.level) < level {
		atomic.StoreInt32(&task.level, level)
		changed = true
	}
	if atomic.LoadInt32(&task.extLevel) < extLevel {
		atomic.StoreInt32(&task.extLevel, extLevel)
		changed = true
	}
	task.lockLevel.Unlock()
	return changed
}

func (task *task) Level() int32 {
	return atomic.LoadInt32(&task.level)
}

func (task *task) ExtLevel() int32 {
	return atomic.LoadInt32(&task.extLevel)
}

// newLoadTask create new load task
func newLoadTask(url string, level int32, extLevel int32, secureAs bool, protocols *map[Protocol]bool, retry int) *task {
	return &task{url: url, level: level, extLevel: extLevel,
		secureAs: secureAs, protocol: URLProtocol(url), protocols: protocols,
		success: false, try: retry,
	}
}

func (t *task) setFileName(fileName string) {
	t.fileName = fileName
}

// internal method, need lock filesLock before
func (d *Downloader) _setTaskFileName(task *task, path string) {
	if path != "" {
		task.setFileName(path)
		d.files.Set(task.fileName, task)
	}
}

func (d *Downloader) taskByFileName(fileName string) *task {
	t, ok := d.files.Get(fileName)
	if ok {
		return t.(*task)
	}
	return nil
}

// internal method, need lock filesLock before
// func (d *Downloader) _setTask(task *task) {
// 	d.processed.Set(task.url, task)
// }

func (d *Downloader) AddTask(t *task) (*task, bool) {
	p, exist := d.processed.GetOrInsert(t.url, t)
	return p.(*task), exist
}

func (d *Downloader) taskByURL(url string) *task {
	t, ok := d.processed.Get(url)
	if ok {
		return t.(*task)
	}
	return nil
}

// internal method, need lock filesLock before
func (d *Downloader) _inrTaskFileName(name string, ext string) (string, error) {
	i := int64(1)
	for {
		newPath := name + "-" + strconv.FormatInt(i, 10) + ext
		if d.taskByFileName(newPath) == nil {
			return newPath, nil
		}
		i++
		if i == math.MaxInt64 {
			return newPath, fmt.Errorf("generate overflow")
		}
	}
}

// internal method, need lock filesLock before
func (d *Downloader) _genTaskFileName(task *task) error {
	u, err := urlx.Parse(task.url)
	if err != nil {
		return err
	}

	path := strings.TrimLeft(u.Path, "/")
	path = strings.ReplaceAll(path, `[~!@#$%^&*()+=\\|\[\]?<>]`, "_")

	switch d.saveMode {
	case FlatMode, FlatDirMode:
		var name string
		if len(path) == 0 {
			path = "index"
		} else {
			i := strings.LastIndex(path, "/")
			if i > 0 {
				if i == len(path)-1 {
					f := strings.LastIndex(path[0:i-1], "/")
					path = "index-" + path[f+1:i]
				} else {
					path = path[i+1:]
				}
			}
		}

		if d.saveMode == FlatDirMode {
			path = appendFlatDir(path, task.contentType)
		}

		path, name, ext := replaceExtension(path, task.contentType)
		if d.taskByFileName(path) != nil {
			path, err = d._inrTaskFileName(name, ext)
			if err != nil {
				return err
			}
		}
		d._setTaskFileName(task, path)
		return nil
	case DirMode, SiteDirMode:
		var name string
		if d.saveMode == SiteDirMode {
			path = strings.ReplaceAll(u.Host, `[~!@#$%^&*()+=\\|\[\]?<>]`, "_") + "/" + path
		}
		if len(path) == 0 {
			path = "index"
		} else if path[len(path)-1] == '/' {
			path += "index"
		}

		path, name, ext := replaceExtension(path, task.contentType)
		if d.taskByFileName(path) != nil {
			path, err = d._inrTaskFileName(name, ext)
			if err != nil {
				return err
			}
		}
		d._setTaskFileName(task, path)
		return nil
	}

	return fmt.Errorf("not realized")
}

func (d *Downloader) recheckTask(task *task) {
	// TODO: reparse files for load after change levels

	if !d.failed {
		d.failed = true
	}
	log.Error().Str("url", task.url).Str("file", task.fileName).Msg("recheck not realized at now")
}

func (d *Downloader) runTask(task *task) {
	if task.success {
		// already doanload, reload and check
		if task.protocol == HTTP && task.contentType == "text/html" {
			d.recheckTask(task)
		}
	} else if task.try > 0 {
		var err error
		switch task.protocol {
		case HTTP:
			err = d.httpLoad(task)
			if err == nil && task.contentType == "text/html" {
				d.recheckTask(task)
			}
		default:
			task.try = 0
			log.Warn().Str("url", task.url).Str("file", task.fileName).Msg("protocol not supported")
			return
		}
		if err != nil {
			if task.try > 1 {
				task.try--
				// requeue task
				d.queue.Put(task)
			}
			if !d.failed {
				d.failed = true
			}
			log.Error().Str("url", task.url).Str("file", task.fileName).Msg(err.Error())
		} else {
			log.Info().Str("url", task.url).Str("file", task.fileName).Int64("size", task.size).Msg("done")
		}
	}
}

func level(url string, baseTask *task, baseHost string) (int32, int32) {
	level := baseTask.Level()
	extLevel := baseTask.ExtLevel()
	host, _ := urlutils.SplitURL(url)
	if host == baseHost {
		level--
	} else {
		// TODO: secureAs
		// TODO: depth (go to N levels down on same site)
		level = extLevel
		extLevel--
	}
	return level, extLevel
}

// AddRootURL add root url to download queue
func (d *Downloader) AddRootURL(url string, level int32, extLevel int32, secureAs bool, protocols *map[Protocol]bool) *Downloader {
	if level == 0 {
		return d
	}
	task := newLoadTask(url, level, extLevel, secureAs, protocols, d.retry)
	if task.protocol == Unsuppoted {
		log.Error().Str("url", task.url).Msg("protocol not supported")
	} else {
		//d.processLock.Lock()
		_, exist := d.AddTask(task)
		if exist {
			//d.processLock.Unlock()
			log.Warn().Str("url", task.url).Msg("already added")
		} else {
			//d.processLock.Unlock()
			//d.root.PushBack(task)
			d.queue.Put(task)
		}
	}
	return d
}

func (d *Downloader) addURL(url string, pageContent bool, retry int, baseTask *task, baseHost string) bool {
	stripURL := urlutils.StripAnchor(url)
	level, extLevel := level(stripURL, baseTask, baseHost)
	queued := false
	exist := true
	if level == 0 && !pageContent {
		return false
	}

	t := d.taskByURL(stripURL)
	if t == nil {
		t = newLoadTask(stripURL, level, extLevel, baseTask.secureAs, baseTask.protocols, d.retry)
		t, exist = d.AddTask(t) // recheck, may be added by concurrent
		if !exist {
			queued = true
		}
	}
	if exist {
		queued = t.UpdateLevel(level, extLevel)
	}
	if queued {
		d.queue.Put(t)
	}
	return true
}
