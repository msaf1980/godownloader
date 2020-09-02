package downloader

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/udhos/equalfile"
)

func testHandler() http.Handler {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("test"))
	mux.Handle("/", fs)

	mux.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/index.html", 301)
	})

	return mux
}

func TestDownloader_httpLoad(t *testing.T) {
	ts := httptest.NewServer(testHandler())
	defer ts.Close()

	tmpdir, err := ioutil.TempDir("", "godownloader-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	dir := tmpdir + "/" + "out"
	saveMode := FlatMode
	d := NewDownloader(saveMode, 1, time.Second, 2)
	d.AddRootURL("http://127.0.0.1", 1, 0, true, nil)
	d.NewLoad(dir)
	if err != nil {
		t.Fatal(err)
	}

	baseAddr := "http://" + ts.Listener.Addr().String()

	cmp := equalfile.New(nil, equalfile.Options{})

	tests := []struct {
		task    *task
		wantErr bool
		errStr  string
		orig    string
	}{
		{newLoadTask(baseAddr+"/index.html", 1, 0, false, nil, 1), false, "", "test/index.html"},
		{newLoadTask(baseAddr+"/1.gz", 1, 0, false, nil, 1), false, "", "test/1.gz"},          // Read file
		{newLoadTask(baseAddr+"/test", 1, 0, false, nil, 1), false, "", "test/index.html"},    // check redirect
		{newLoadTask(baseAddr+"/not_found.html", 1, 0, false, nil, 1), true, "Not found", ""}, // not found
	}
	for _, tt := range tests {
		t.Run(tt.task.url, func(t *testing.T) {
			var err error
			if err = d.httpLoad(tt.task); (err != nil) != tt.wantErr {
				t.Errorf("Downloader.httpLoad() error = '%v', wantErr '%v'", err, tt.wantErr)
			}
			if err != nil && len(tt.errStr) > 0 && err.Error() != tt.errStr {
				t.Errorf("Downloader.httpLoad() error = '%v', wantErr '%s'", err, tt.errStr)
			}
			if err == nil && len(tt.orig) > 0 {
				equal, err := cmp.CompareFile(tt.orig, d.outdir+"/"+tt.task.fileName)
				if err != nil {
					t.Errorf("Downloader.httpLoad() compare error '%v'", err)
				} else if !equal {
					t.Errorf("Downloader.httpLoad() result file mismatched %s", d.outdir+"/"+tt.task.fileName)
				}
			}
		})
	}
}

// func TestDownloader_htmlExtractLinks(t *testing.T) {
// 	tests := []struct {
// 		path    string
// 		wantErr bool
// 	}{
// 		{"index.html", false},
// 	}

// 	d := NewDownloader(FlatMode, 1, time.Second, 2)
// 	d.outdir = "test"
// 	for _, tt := range tests {
// 		t.Run(tt.path, func(t *testing.T) {
// 			task := task{fileName: tt.path}
// 			if err := d.htmlExtractLinks(&task); (err != nil) != tt.wantErr {
// 				t.Errorf("Downloader.htmlExtractLinks() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
