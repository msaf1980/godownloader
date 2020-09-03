package downloader

import (
	"fmt"
	"math"
	"strconv"
	"strings"

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
	level     int  // download level (from same site)
	extLevel  int  // download level (from external sites)
	secureAs  bool // cast secure and unsecure links to one site (no external)
	protocols *map[Protocol]bool

	fileName    string // relative filename (blank if no try downloads else)
	contentType string
	success     bool
	size        int64 // size from header
	try         int   // retry count - stop on 0 or success
}

// newLoadTask create new load task
func newLoadTask(url string, level int, extLevel int, secureAs bool, protocols *map[Protocol]bool, retry int) *task {
	return &task{url: url, level: level, extLevel: extLevel,
		secureAs: secureAs, protocol: URLProtocol(url), protocols: protocols,
		success: false, try: retry,
	}
}

func (t *task) setFileName(fileName string) {
	t.fileName = fileName
}

func (d *Downloader) inrTaskFileName(name string, ext string) (string, error) {
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

func (d *Downloader) genTaskFileName(task *task) error {
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
			path, err = d.inrTaskFileName(name, ext)
			if err != nil {
				return err
			}
		}
		d.setTaskFileName(task, path)
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
			path, err = d.inrTaskFileName(name, ext)
			if err != nil {
				return err
			}
		}
		d.setTaskFileName(task, path)
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

func level(url string, baseTask *task, baseHost string) (int, int) {
	level := baseTask.level
	extLevel := baseTask.extLevel
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

func (d *Downloader) addURL(url string, contentType string, rel string, retry int, baseTask *task, baseHost string) bool {
	if contentType == "application/rss+xml" || rel == "alternate" || rel == "search" || rel == "canonical" {
		return false
	}
	stripURL := urlutils.StripAnchor(url)
	queued := false
	d.processLock.Lock()
	task, ok := d.processed[stripURL]
	if ok {
		level, extLevel := level(stripURL, baseTask, baseHost)
		if task.level < level {
			task.level = level
			queued = true
		}
		if task.extLevel < extLevel {
			task.extLevel = extLevel
			queued = true
		}
	} else {
		level, extLevel := level(stripURL, baseTask, baseHost)
		if level == 0 {
			d.processLock.Unlock()
			return false
		}
		task = newLoadTask(stripURL, level, extLevel, baseTask.secureAs, baseTask.protocols, d.retry)
		d.setTask(task)
		queued = true
	}
	d.processLock.Unlock()
	if queued {
		d.queue.Put(task)
	}
	return true
}
