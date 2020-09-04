package downloader

import (
	"testing"
	"time"
)

func Test_task_TryLock(t *testing.T) {
	task := newLoadTask("", 0, 0, 0)
	if !task.TryLock() {
		t.Fatalf("task.TryLock() can't lock")

	}
	if task.TryLock() {
		t.Fatalf("task.TryLock() lock already locked")
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
			task:  newLoadTask("http://test.int/index.html", 1, 0, 1),
			exist: false,
		},
		{
			name:  "readd first task (no changes)",
			task:  newLoadTask("http://test.int/index.html", 1, 0, 1),
			exist: true,
		},
		{
			name:  "readd first task (with changes)",
			task:  newLoadTask("http://test.int/index.html", 2, 1, 1),
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
