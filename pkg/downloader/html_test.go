package downloader

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
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
	d.AddRootURL("http://127.0.0.1/", 2, 0, 0)
	_, _ = d.NewLoad(dir)
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
		links   map[string]bool
	}{
		{newLoadTask(
			baseAddr+"/index.html", "/", 2, 0, 0, 1), false, "", "test/index.html.tpl",
			map[string]bool{
				"http://127.0.0.1/":          true, // root URL, not downloaded at this step
				baseAddr + "/link1.html":     true,
				baseAddr + "/not_found.html": true,
				baseAddr + "/style.css":      true, baseAddr + "/1.gif": true, baseAddr + "/1.gz": true,
			},
		},
		{
			newLoadTask(baseAddr+"/1.gz", "/", 1, 0, 0, 1), false, "", "test/1.gz",
			map[string]bool{},
		}, // Read file
		{
			newLoadTask(baseAddr+"/test", "/", 2, 0, 0, 1), false, "", "test/index.html.tpl",
			map[string]bool{},
		}, // check redirect
		{
			newLoadTask(baseAddr+"/link1.html", "/", 2, 0, 0, 1), false, "", "test/link1.html.tpl",
			map[string]bool{
				"http://127.0.0.1/":      true, // root URL, not downloaded at this step
				baseAddr + "/index.html": true, baseAddr + "/link1.html": true, baseAddr + "/link2.html": true,
				baseAddr + "/not_found.html": true,
				baseAddr + "/style.css":      true, baseAddr + "/1.gif": true, baseAddr + "/1.gz": true,
			},
		},
		{
			newLoadTask(baseAddr+"/link2.html", "/", 1, 0, 0, 1), false, "", "test/link2.html.tpl",
			map[string]bool{
				"http://127.0.0.1/":      true, // root URL, not downloaded at this step
				baseAddr + "/index.html": true, baseAddr + "/link1.html": true, baseAddr + "/link2.html": true,
				baseAddr + "/not_found.html": true,
				baseAddr + "/style.css":      true, baseAddr + "/1.gif": true, baseAddr + "/1.gz": true,
			},
		},
		{
			newLoadTask(baseAddr+"/not_found.html", "/", 1, 0, 0, 1), true, "Not found", "",
			map[string]bool{},
		}, // not found
	}
	for _, tt := range tests {
		t.Run(tt.task.url, func(t *testing.T) {
			var err error
			if err = d.httpLoad(tt.task); (err != nil) != tt.wantErr {
				t.Fatalf("Downloader.httpLoad() error = '%v', wantErr '%v'", err, tt.wantErr)
			}
			if err != nil && len(tt.errStr) > 0 && err.Error() != tt.errStr {
				t.Fatalf("Downloader.httpLoad() error = '%v', wantErr '%s'", err, tt.errStr)
			}
			if len(tt.orig) > 0 {
				if strings.HasSuffix(tt.orig, ".tpl") {
					result, err := ioutil.ReadFile(d.outdir + "/" + tt.task.fileName)
					if err != nil {
						t.Fatalf("Downloader.httpLoad() load error '%v'", err)
					}
					resultHTML := string(result)
					data, err := ioutil.ReadFile(tt.orig)
					if err != nil {
						t.Errorf("Downloader.httpLoad() load template error '%v'", err)
					} else {
						want := strings.ReplaceAll(string(data), "{{ Host }}", baseAddr)
						if want != resultHTML {
							dmp := diffmatchpatch.New()
							diffs := dmp.DiffMain(want, resultHTML, false)
							t.Errorf("Downloader.httpLoad() result html file mismatched %s, diff\n%s", d.outdir+"/"+tt.task.fileName, diffs)
						}
					}
				} else {
					equal, err := cmp.CompareFile(tt.orig, d.outdir+"/"+tt.task.fileName)
					if err != nil {
						t.Errorf("Downloader.httpLoad() compare error '%v'", err)
					} else if !equal {
						t.Errorf("Downloader.httpLoad() result file mismatched %s", d.outdir+"/"+tt.task.fileName)
					}
				}
			}

			if len(tt.links) > 0 {
				for k := range d.processed.Iter() {
					url := k.Key.(string)
					_, ok := tt.links[url]
					if !ok {
						t.Errorf("Downloader.httpLoad() link %s extracted, but required", url)
					}
				}

				for url := range tt.links {
					_, ok := d.processed.Get(url)
					if !ok {
						t.Errorf("Downloader.httpLoad() link %s not extracted", url)
					}
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
