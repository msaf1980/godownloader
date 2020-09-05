package downloader

import (
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

func Test_genTaskFileName_FlatMode(t *testing.T) {
	saveMode := FlatMode
	d := NewDownloader(saveMode, 1, time.Second, 1)

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
	saveMode := FlatDirMode
	d := NewDownloader(saveMode, 1, time.Second, 1)

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
	saveMode := DirMode
	d := NewDownloader(saveMode, 1, time.Second, 1)

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
	saveMode := SiteDirMode
	d := NewDownloader(saveMode, 1, time.Second, 1)

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

	type args struct {
		t *task
	}
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
