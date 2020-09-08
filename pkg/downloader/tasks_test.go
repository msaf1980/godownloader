package downloader

import (
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/msaf1980/godownloader/pkg/urlutils"
)

func Test_task_TryLock(t *testing.T) {
	task := newLoadTask("", "/", 0, 0, 0, 0)
	if !task.TryLock() {
		t.Fatalf("task.TryLock() can't lock")

	}
	if task.TryLock() {
		t.Fatalf("task.TryLock() lock already locked")
	}
}

func TestDownloader_Map(t *testing.T) {
	var err error
	d := NewDownloader(FlatMode, 1, time.Second, 1)
	d.fMap, err = ioutil.TempFile("", "godownloader")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		d.closeMap()
		os.Remove(d.fMap.Name())
	}()

	tests := []*task{
		{url: "http://test.int/index.html", fileName: "index.html", contentType: "text/html"},
		{url: "http://test.int/link1.html", fileName: "link1.html", contentType: "text/html"},
		{url: "http://test.int/1.gif", fileName: "1.gif", contentType: "image/gif"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err = d._storeMap(tt)
			if err != nil {
				t.Fatalf("Downloader._storeMap() error = %v", err)
			}
		})
	}

	// Verify
	dv := NewDownloader(FlatMode, 1, time.Second, 1)

	dv.AddRootURL("http://test.int/index.html", 1, 0, 0)
	// Sync test task links
	tests[0].links = 1

	dv.fileMap = d.fMap.Name()
	err = dv.openMap()
	if err != nil {
		t.Fatal(err)
	}
	err = dv.closeMap()
	if err != nil {
		t.Error(err)
	}
	if dv.processed.Len() != len(tests) {
		t.Errorf("map length = %d, want %d", dv.processed.Len(), len(tests))
	}
	for _, task := range tests {
		if lTask := dv.taskByURL(task.url); lTask == nil {
			t.Errorf("map url %s not found", task.url)
		} else {
			if lTask.fileName != task.fileName {
				t.Errorf("map url %s fileName  = '%s', want '%s'", task.url, lTask.fileName, task.fileName)
			}
			if lTask.fileName != task.fileName {
				t.Errorf("map url %s contentType  = '%s', want '%s'", task.url, lTask.contentType, task.contentType)
			}
			links := task.Links()
			lLinks := lTask.Links()
			downLevel := task.DownLevel()
			lDownLevel := lTask.DownLevel()
			extLinks := task.ExtLinks()
			lExtLinks := lTask.ExtLinks()
			if links != lLinks {
				t.Errorf("map url %s links  = %d, want %d", task.url, lLinks, links)
			}
			if downLevel != lDownLevel {
				t.Errorf("map url %s downLevel  = %d, want %d", task.url, lDownLevel, downLevel)
			}
			if extLinks != lExtLinks {
				t.Errorf("map url %s extLinks  = %d, want %d", task.url, lExtLinks, extLinks)
			}
		}

	}
}

func Test_genTaskFileName_FlatMode(t *testing.T) {
	var err error
	saveMode := FlatMode
	d := NewDownloader(saveMode, 1, time.Second, 1)
	d.fMap, err = ioutil.TempFile("", "godownloader")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		d.closeMap()
		os.Remove(d.fMap.Name())
	}()

	tests := []struct {
		name        string
		url         string
		contentType string
		want        string
	}{
		{
			name: "check http://test.com/", url: "http://test.com/",
			contentType: "text/html", want: "index.html",
		},
		{
			name: "check http://test.com/test/", url: "http://test.com/test/",
			contentType: "text/html", want: "index-test.html",
		},
		{
			name: "check http://test.com/test.html", url: "http://test.com/test.html",
			contentType: "text/html", want: "test.html",
		},
		{
			name: "check http://test.com/2/test.html with 1 iteration", url: "http://test.com/2/test.html",
			contentType: "text/html", want: "test-1.html",
		},
		{
			name: "check http://test.com/3/test.html with 2 iteration", url: "http://test.com/3/test.html",
			contentType: "text/html", want: "test-2.html",
		},
		{
			name: "check http://test.com/i/1", url: "http://test.com/i/1",
			contentType: "image/gif", want: "1.gif",
		},
		{
			name: "check http://test.com/1.zip", url: "http://test.com/1.zip",
			contentType: "application/zip", want: "1.zip",
		},
		{ // Rewrite *.php with text/html
			name: "check http://test.com/index.php?p=12 with 1 iteration", url: "http://test.com/index.php?p=12",
			contentType: "text/html", want: "index-1.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := task{url: tt.url, contentType: tt.contentType}
			err := d._genTaskFileName(&task)
			if err != nil {
				t.Fatalf("taskGenerateFilename() return erorr '%s'", err)
			}
			if task.fileName != tt.want {
				t.Errorf("taskGenerateFilename() = %v, want %v", task.fileName, tt.want)
			}
		})
	}
}

func Test_genTaskFileName_FlatDirMode(t *testing.T) {
	var err error
	saveMode := FlatDirMode
	d := NewDownloader(saveMode, 1, time.Second, 1)
	d.AddRootURL("http://127.0.0.1/", 2, 0, 0)
	tmpdir, err := ioutil.TempDir("", "godownloader-")
	if err != nil {
		t.Fatal(err)
	}

	dir := tmpdir + "/" + "out"
	_, err = d.NewLoad(dir, "godownloader.map")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		d.closeMap()
		defer os.RemoveAll(tmpdir)
	}()

	tests := []struct {
		name        string
		url         string
		contentType string
		want        string
	}{
		{
			name: "check http://test.com/", url: "http://test.com/",
			contentType: "text/html", want: "index.html",
		},
		{
			name: "check http://test.com/test/", url: "http://test.com/test/",
			contentType: "text/html", want: "index-test.html",
		},
		{
			name: "check http://test.com/test.html", url: "http://test.com/test.html",
			contentType: "text/html", want: "test.html",
		},
		{
			name: "check http://test.com/2/test.html with 1 iteration", url: "http://test.com/2/test.html",
			contentType: "text/html", want: "test-1.html",
		},
		{
			name: "check http://test.com/3/test.html with 2 iteration", url: "http://test.com/3/test.html",
			contentType: "text/html", want: "test-2.html",
		},
		{
			name: "check http://test.com/i/1", url: "http://test.com/i/1",
			contentType: "image/gif", want: "img/1.gif",
		},
		{
			name: "check http://test.com/1.zip", url: "http://test.com/1.zip",
			contentType: "application/zip", want: "download/1.zip",
		},
		{ // Rewrite *.php with text/html
			name: "check http://test.com/index.php?p=12 with 1 iteration", url: "http://test.com/index.php?p=12",
			contentType: "text/html", want: "index-1.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := task{url: tt.url, contentType: tt.contentType}
			err := d._genTaskFileName(&task)
			if err != nil {
				t.Fatalf("taskGenerateFilename() return error '%s'", err)
			}
			if task.fileName != tt.want {
				t.Errorf("taskGenerateFilename() = %v, want %v", task.fileName, tt.want)
			}
		})
	}
}

func Test_genTaskFileName_DirMode(t *testing.T) {
	var err error
	saveMode := DirMode
	d := NewDownloader(saveMode, 1, time.Second, 1)
	d.AddRootURL("http://127.0.0.1/", 2, 0, 0)
	tmpdir, err := ioutil.TempDir("", "godownloader-")
	if err != nil {
		t.Fatal(err)
	}

	dir := tmpdir + "/" + "out"
	_, err = d.NewLoad(dir, "godownloader.map")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		d.closeMap()
		defer os.RemoveAll(tmpdir)
	}()

	tests := []struct {
		name        string
		url         string
		contentType string
		want        string
	}{
		{
			name: "check http://test.com/", url: "http://test.com/",
			contentType: "text/html", want: "index.html",
		},
		{
			name: "check http://test.com/test/", url: "http://test.com/test/",
			contentType: "text/html", want: "test/index.html",
		},
		{
			name: "check http://test.com/test.html", url: "http://test.com/test.html",
			contentType: "text/html", want: "test.html",
		},
		{
			name: "check http://test.com/2/test.html", url: "http://test.com/2/test.html",
			contentType: "text/html", want: "2/test.html",
		},
		{
			name: "check http://test.com/3/test.html", url: "http://test.com/3/test.html",
			contentType: "text/html", want: "3/test.html",
		},
		{
			name: "check http://test.com/i/1", url: "http://test.com/i/1",
			contentType: "image/gif", want: "i/1.gif",
		},
		{
			name: "check http://test.com/1.zip", url: "http://test.com/1.zip",
			contentType: "application/zip", want: "1.zip",
		},
		{ // Rewrite *.php with text/html
			name: "check http://test.com/index.php?p=12 with 1 iteration", url: "http://test.com/index.php?p=12",
			contentType: "text/html", want: "index-1.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := task{url: tt.url, contentType: tt.contentType}
			err := d._genTaskFileName(&task)
			if err != nil {
				t.Fatalf("taskGenerateFilename() return error '%s'", err)
			}
			if task.fileName != tt.want {
				t.Errorf("taskGenerateFilename() = %v, want %v", task.fileName, tt.want)
			}
		})
	}
}

func Test_genTaskFileName_SiteDirMode(t *testing.T) {
	var err error
	saveMode := SiteDirMode
	d := NewDownloader(saveMode, 1, time.Second, 1)
	d.AddRootURL("http://127.0.0.1/", 2, 0, 0)
	tmpdir, err := ioutil.TempDir("", "godownloader-")
	if err != nil {
		t.Fatal(err)
	}

	dir := tmpdir + "/" + "out"
	_, err = d.NewLoad(dir, "godownloader.map")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		d.closeMap()
		defer os.RemoveAll(tmpdir)
	}()

	tests := []struct {
		name        string
		url         string
		contentType string
		want        string
	}{
		{
			name: "check http://test.com/", url: "http://test.com/",
			contentType: "text/html", want: "test.com/index.html",
		},
		{
			name: "check http://test.com/test/", url: "http://test.com/test/",
			contentType: "text/html", want: "test.com/test/index.html",
		},
		{
			name: "check http://test2.com/test.html", url: "http://test2.com/test.html",
			contentType: "text/html", want: "test2.com/test.html",
		},
		{
			name: "check http://test.com/2/test.html", url: "http://test.com/2/test.html",
			contentType: "text/html", want: "test.com/2/test.html",
		},
		{
			name: "check http://test3.com/3/test.html", url: "http://test3.com/3/test.html",
			contentType: "text/html", want: "test3.com/3/test.html",
		},
		{
			name: "check http://test.com/i/1", url: "http://test.com/i/1",
			contentType: "image/gif", want: "test.com/i/1.gif",
		},
		{
			name: "check http://test.com/1.zip", url: "http://test.com/1.zip",
			contentType: "application/zip", want: "test.com/1.zip",
		},
		{ // Rewrite *.php with text/html
			name: "check http://test.com/index.php?p=12 with 1 iteration", url: "http://test.com/index.php?p=12",
			contentType: "text/html", want: "test.com/index-1.html",
		},
		{ // Rewrite *.php with text/html
			name: "check http://test3.com/index.php?p=12", url: "http://test3.com/index.php?p=12",
			contentType: "text/html", want: "test3.com/index.html",
		},
		{ // Алгоритмы
			name: "check http://Алгоритмы.com:8080/Алгоритм.html", url: `http://%D0%90%D0%BB.com:8080/%D0%90%D0%BB%D0%B3%D0%BE%D1%80%D0%B8%D1%82%D0%BC.html`,
			contentType: "text/html", want: "al.com_8080/Algoritm.html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := task{url: tt.url, contentType: tt.contentType}
			err := d._genTaskFileName(&task)
			if err != nil {
				t.Fatalf("taskGenerateFilename() return error '%s'", err)
			}
			if task.fileName != tt.want {
				t.Errorf("taskGenerateFilename() = %v, want %v", task.fileName, tt.want)
			}
		})
	}
}

func Test_level(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		baseTask      *task
		wantLinks     int32
		wantDownLevel int32
		wantExtLinks  int32
	}{
		{
			"no download (same site dir)", "http://test.int/test1/test2/link1.html",
			newLoadTask("http://test.int/test1/test2/index.html", "/test1/test2/",
				1, 1, 1, 1,
			),
			0, 1, 1,
		},
		{
			"download (same site dir)", "http://test.int/test1/test2/link2.html",
			newLoadTask("http://test.int/test1/test2/index.html", "/test1/test2/",
				2, 2, 1, 1,
			),
			1, 2, 1,
		},
		{
			"download (up to site dir)", "http://test.int/test1/test2/test3/link3.html",
			newLoadTask("http://test.int/test1/test2/index.html", "/test1/test2/",
				4, 1, 1, 1,
			),
			3, 1, 1,
		},
		{
			"no download (down from site dir) #1", "http://test.int/test1/link1.html",
			newLoadTask("http://test.int/test1/test2/index.html", "/test1/test2/",
				2, 0, 1, 1,
			),
			0, 0, 0,
		},
		{
			"download (down from site dir) #1", "http://test.int/test1/link1.html",
			newLoadTask("http://test.int/test1/test2/index.html", "/test1/test2/",
				3, 1, 1, 1,
			),
			2, 0, 1,
		},
		{
			"download (down-up from site dir) #2", "http://test.int/test1/test3/link1.html",
			newLoadTask("http://test.int/test1/test2/index.html", "/test1/test2/",
				3, 1, 0, 1,
			),
			2, 0, 0,
		},
		{
			"no download (from other site)", "http://test.int/test1/link1.html",
			newLoadTask("http://no.int/test1/test2/index.html", "/test1/test2/",
				2, 3, 0, 1,
			),
			0, 0, 0,
		},
		{
			"download (from other site)", "http://test.int/test1/link1.html",
			newLoadTask("http://no.int/test1/test2/index.html", "/test1/test2/",
				2, 1, 3, 1,
			),
			3, 0, 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseHost, _ := urlutils.SplitURL(tt.baseTask.url)
			links, downLevel, extLinks := level(tt.url, baseHost, tt.baseTask.rootDir,
				tt.baseTask.Links(), tt.baseTask.DownLevel(), tt.baseTask.ExtLinks(),
			)
			if links != tt.wantLinks {
				t.Errorf("Links() links = %v, want %v", links, tt.wantLinks)
			}
			if downLevel != tt.wantDownLevel {
				t.Errorf("Links() downLinks = %v, want %v", downLevel, tt.wantDownLevel)
			}
			if extLinks != tt.wantExtLinks {
				t.Errorf("Links() extLinks = %v, want %v", extLinks, tt.wantExtLinks)
			}
		})
	}
}

func TestDownloader_addTask(t *testing.T) {
	d := NewDownloader(FlatMode, 1, time.Second, 1)

	tests := []struct {
		name  string
		task  *task
		exist bool
	}{
		{
			name:  "add first task",
			task:  newLoadTask("http://test.int/index.html", "/", 1, 0, 0, 1),
			exist: false,
		},
		{
			name:  "readd first task (no changes)",
			task:  newLoadTask("http://test.int/index.html", "/", 1, 0, 0, 1),
			exist: true,
		},
		{
			name:  "readd first task (with changes)",
			task:  newLoadTask("http://test.int/index.html", "/", 1, 0, 0, 1),
			exist: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, exist := d.addTask(tt.task)
			if exist != tt.exist {
				t.Fatalf("Downloader.addTask() exist got = %v, want %v", exist, tt.exist)
			} else if exist && task == tt.task {
				t.Fatalf("Downloader.addTask() with %s exist, but got checked task", tt.task.url)
			}
			if exist && task.url != tt.task.url {
				t.Errorf("Downloader.addTask() url got = %s, want %s", task.url, tt.task.url)
			}
		})
	}
}

func TestDownloader_runTaskNew(t *testing.T) {
	ts := httptest.NewServer(testHandler())
	defer ts.Close()

	tmpdir, err := ioutil.TempDir("", "godownloader-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	baseAddr := "http://" + ts.Listener.Addr().String()
	dir := tmpdir + "/" + "out"

	d := NewDownloader(SiteDirMode, 1, time.Second, 2)

	d.AddRootURL(baseAddr+"/index.html", 2, 0, 0)
	rootTpl := "test/index.html.tpl"
	urlQueue := map[string]bool{
		baseAddr + "/style.css":      true,
		baseAddr + "/1.gif":          true,
		baseAddr + "/1.gz":           true,
		baseAddr + "/link1.html":     true,
		baseAddr + "/not_found.html": true,
	}

	_, err = d.NewLoad(dir, "godownloader.map")
	if err != nil {
		t.Fatal(err)
	}

	p, ok := d.queue.Get()
	if !ok {
		t.Fatal("Downloader.queue emphy")
	}
	ta := p.(*task)
	if !d.runTask(ta) {
		t.Fatal("Downloader.runTask() = false, want true")
	}

	verifyFile(t, d.outdir, ta.fileName, rootTpl, baseAddr)
	n := 0
	for {
		p, ok = d.queue.Get()
		if !ok {
			break
		}
		ta = p.(*task)
		_, ok = urlQueue[ta.url]
		if !ok {
			t.Errorf("Downloader.runTask() queue unknown url '%s'", ta.url)
		}
		n++
	}
	if n != len(urlQueue) {
		t.Fatalf("Downloader.runTask() produce queue len = %d, want %d", n, len(urlQueue))
	}
}
