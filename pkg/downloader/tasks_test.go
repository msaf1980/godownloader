package downloader

import (
	"container/list"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	lockfree_queue "github.com/msaf1980/go-lockfree-queue"
)

func TestDownloader_runTask(t *testing.T) {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("test"))
	mux.Handle("/", fs)

	ts := httptest.NewUnstartedServer(mux)
	ts.EnableHTTP2 = true
	ts.Start()
	defer ts.Close()

	type fields struct {
		saveMode     SaveMode
		simulate     bool
		timeout      time.Duration
		retry        int
		maxRedirects int
		queue        *lockfree_queue.Queue
		processLock  sync.Mutex
		processed    map[string]*task
		downloads    map[string]*task
		root         *list.List
		wg           sync.WaitGroup
		running      bool
		download     int32
		failed       bool
	}
	type args struct {
		task *task
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
