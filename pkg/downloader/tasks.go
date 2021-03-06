package downloader

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/msaf1980/godownloader/pkg/fileutils"
	"github.com/msaf1980/godownloader/pkg/strutils"
	"github.com/msaf1980/godownloader/pkg/urlutils"

	"github.com/goware/urlx"
	"github.com/rs/zerolog/log"
)

// Protocol
type Protocol int8

const (
	Unsuppoted Protocol = iota
	HTTP
)

func StringToProtocol(s string) Protocol {
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
		return StringToProtocol(url[0:i])
	}
	return Unsuppoted
}

type task struct {
	url       string
	rootDir   string
	protocol  Protocol
	links     int32 // download links (from same site on same or upper dir)
	downLevel int32 // download links (from same sites underlying directories)
	extLinks  int32 // download links (from external sites)

	fileName    string // relative filename (blank if no try downloads else)
	contentType string
	success     bool
	size        int64 // size from header
	try         int   // retry count - stop on 0 or success

	lock      uint32     // atomic set 1 for hold task during download/parse (TryLock) and relase when done (Unlock)
	lockLevel sync.Mutex // set 1 for hold task during level
}

func (task *task) TryLock() bool {
	return atomic.CompareAndSwapUint32(&task.lock, 0, 1)
}

func (task *task) UnLock() {
	atomic.CompareAndSwapUint32(&task.lock, 1, 0)
}

func (task *task) UpdateLinks(links int32, downLevel int32, extLinks int32) bool {
	task.lockLevel.Lock()
	changed := false
	if atomic.LoadInt32(&task.links) < links {
		atomic.StoreInt32(&task.links, links)
		changed = true
	}
	if atomic.LoadInt32(&task.downLevel) < downLevel {
		atomic.StoreInt32(&task.downLevel, downLevel)
		changed = true
	}
	if atomic.LoadInt32(&task.extLinks) < extLinks {
		atomic.StoreInt32(&task.extLinks, extLinks)
		changed = true
	}
	task.lockLevel.Unlock()
	return changed
}

func (task *task) Links() int32 {
	return atomic.LoadInt32(&task.links)
}

func (task *task) DownLevel() int32 {
	return atomic.LoadInt32(&task.downLevel)
}

func (task *task) ExtLinks() int32 {
	return atomic.LoadInt32(&task.extLinks)
}

// newLoadTask create new load task
func newLoadTask(url, rootDir string, links int32, downLevel int32, extLinks int32, retry int) *task {
	return &task{url: url, rootDir: rootDir, links: links, downLevel: downLevel, extLinks: extLinks,
		protocol: URLProtocol(url),
		success:  false, try: retry,
	}
}

// internal method, need lock filesLock before
func (d *Downloader) _setTaskFileName(task *task, p string) {
	if p != "" {
		task.fileName = p
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

func (d *Downloader) addTask(t *task) (*task, bool) {
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

func (d *Downloader) newMap() (err error) {
	if d.fMap != nil {
		return fmt.Errorf("map already open")
	}
	d.fMap, err = os.OpenFile(d.fileMap, os.O_CREATE|os.O_RDWR, 0o644)
	return
}

func (d *Downloader) openMap() (err error) {
	if d.fMap != nil {
		return fmt.Errorf("map already open")
	}
	d.fMap, err = os.OpenFile(d.fileMap, os.O_RDWR, 0o644)
	scanner := bufio.NewScanner(d.fMap)
	var t *task
	for scanner.Scan() {
		line := scanner.Text()
		if t == nil {
			t = &task{url: line}
		} else {
			s := strings.Split(line, " ")
			if len(s) != 2 {
				return fmt.Errorf("map fileName/contentType line incomplete: %s", line)
			}
			t.fileName = s[0]
			t.contentType = s[1]
			task, exist := d.addTask(t)
			if exist {
				task.fileName = t.fileName
				task.contentType = t.contentType
				task.UpdateLinks(t.Links(), t.DownLevel(), t.ExtLinks())
			}
			t = nil
		}
	}
	return
}

func (d *Downloader) closeMap() error {
	if d.fMap == nil {
		return nil
	}
	return d.fMap.Close()
}

// internal method, during filename generate
func (d *Downloader) _storeMap(task *task) error {
	_, err := d.fMap.Write([]byte(task.url + "\n" + task.fileName + " " + task.contentType + "\n"))
	if err != nil {
		d.Abort()
	}
	return err
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

func mkdir(dir string) error {
	exist, isDir, err := fileutils.Exist(dir)
	if err == nil {
		if exist {
			if isDir {
				return nil
			}
			return fmt.Errorf("%s is not dir", dir)
		}
		err = os.MkdirAll(dir, 0o755)
	}
	return err
}

// internal method, need lock filesLock before
func (d *Downloader) _genTaskFileName(task *task) error {
	u, err := urlx.Parse(task.url)
	if err != nil {
		return err
	}

	p := strings.TrimLeft(u.Path, "/")
	p = strutils.TranslitWithoutSpecSymbols(p, '_')

	switch d.saveMode {
	case FlatMode, FlatDirMode:
		var name string
		if len(p) == 0 {
			p = "index"
		} else {
			i := strings.LastIndex(p, "/")
			if i > 0 {
				if i == len(p)-1 {
					f := strings.LastIndex(p[0:i-1], "/")
					p = "index-" + p[f+1:i]
				} else {
					p = p[i+1:]
				}
			}
		}

		if d.saveMode == FlatDirMode {
			var dir string
			p, _ = appendFlatDir(p, task.contentType)
			err = mkdir(d.outdir + "/" + dir)
			if err != nil {
				return err
			}
		}

		p, name, ext := replaceExtension(p, task.contentType)
		if d.taskByFileName(p) != nil {
			p, err = d._inrTaskFileName(name, ext)
			if err != nil {
				return err
			}
		}
		d._setTaskFileName(task, p)
		return d._storeMap(task)
	case DirMode, SiteDirMode:
		var name string

		if len(p) > 0 && p[0] == '/' {
			if len(p) == 1 {
				p = "index"
			} else {
				p = p[1:]
			}
		}
		if len(p) == 0 {
			p = "index"
		} else {
			if p[len(p)-1] == '/' {
				p += "index"
			}
		}

		if d.saveMode == SiteDirMode {
			p = strutils.TranslitWithoutSpecSymbols(u.Host, '_') + "/" + p
		}

		p, name, ext := replaceExtension(p, task.contentType)
		err = mkdir(path.Dir(d.outdir + "/" + p))
		if err != nil {
			return err
		}
		if d.taskByFileName(p) != nil {
			p, err = d._inrTaskFileName(name, ext)
			if err != nil {
				return err
			}
		}
		d._setTaskFileName(task, p)
		return d._storeMap(task)
	}

	return fmt.Errorf("not realized")
}

func (d *Downloader) recheckTask(task *task) bool {
	// TODO: reparse files for load after change levels

	if !d.failed {
		d.failed = true
	}
	log.Error().Str("url", task.url).Str("file", task.fileName).Msg("recheck not realized at now")
	return false
}

func (d *Downloader) runTask(task *task) bool {
	// Check if file exist (continue download)
	if !task.success && len(task.fileName) > 0 {
		if s, err := os.Stat(d.outdir + "/" + task.fileName); err == nil {
			if s.IsDir() {
				log.Error().Str("url", task.url).Str("file", task.fileName).Msg("must be a file")
				return false
			}
			task.success = true
		} else if !os.IsNotExist(err) {
			log.Error().Str("url", task.url).Str("file", task.fileName).Msg(err.Error())
			return false
		}
	}

	if task.success {
		// already doanload, reload and check
		if task.protocol == HTTP && task.contentType == "text/html" {
			return d.recheckTask(task)
		}
		return true
	} else if task.try > 0 {
		var err error
		switch task.protocol {
		case HTTP:
			err = d.httpLoad(task)
			// if err == nil && task.contentType == "text/html" {
			// 	return d.recheckTask(task)
			// }
		default:
			task.try = 0
			log.Warn().Str("url", task.url).Str("file", task.fileName).Msg("protocol not supported")
			return false
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
			return false
		}
		log.Info().Str("url", task.url).Str("file", task.fileName).Int64("size", task.size).Msg("done")
		return true
	}
	return false
}

func level(url, baseHost, baseDir string, links, downLevel, extLinks int32) (int32, int32, int32) {
	if links == 0 {
		return 0, 0, 0
	}
	host, path := urlutils.SplitURL(url)
	if host == baseHost {
		if strings.HasPrefix(path, baseDir) {
			links--
		} else {
			// TODO: check for level and downLevel
			rootDir := urlutils.RootDir(urlutils.BaseURLDir(path), baseDir)
			n := int32(strutils.CountRune(baseDir[len(rootDir):], '/'))
			if n <= downLevel {
				links--
				downLevel -= n
			} else {
				links = 0
				downLevel = 0
				extLinks = 0
			}
		}
	} else {
		// TODO: secureAs
		// TODO: depth (go to N levels down on same site)
		links = extLinks
		downLevel = 0
		extLinks = 0
	}
	return links, downLevel, extLinks
}

// AddRootURL add root url to download queue
func (d *Downloader) AddRootURL(url string, level int32, downLevel int32, extLevel int32) bool {
	if level < 1 {
		return false
	}
	_, p := urlutils.SplitURL(url)
	dir := urlutils.BaseURLDir(p)
	task := newLoadTask(url, dir, level, downLevel, extLevel, d.retry)
	//d.processLock.Lock()
	_, exist := d.addTask(task)
	if exist {
		//d.processLock.Unlock()
		return false
	} else {
		//d.processLock.Unlock()
		//d.root.PushBack(task)
		d.queue.Put(task)
	}

	return true
}

func (d *Downloader) addURL(url string, pageContent bool, retry int,
	baseHost string, baseDir string,
	baseLinks int32, baseDownLevel int32, baseExtLinks int32) bool {

	stripURL := urlutils.StripAnchor(url)
	links, downLevel, extLinks := level(stripURL, baseHost, baseDir, baseLinks, baseDownLevel, baseExtLinks)
	queued := false
	exist := true
	if !pageContent {
		if links < 1 {
			return false
		}
	}

	t := d.taskByURL(stripURL)
	if t == nil {
		t = newLoadTask(stripURL, baseDir, links, downLevel, extLinks, d.retry)
		t, exist = d.addTask(t) // recheck, may be added by concurrent
		if !exist {
			queued = true
		}
	}
	if exist {
		queued = t.UpdateLinks(links, downLevel, extLinks)
	}
	if queued {
		d.queue.Put(t)
	}
	return true
}
