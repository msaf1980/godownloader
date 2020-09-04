package downloader

import (
	"testing"
)

func Test_task_TryLock(t *testing.T) {
	task := newLoadTask("", 0, 0, false, nil, 0)
	if !task.TryLock() {
		t.Fatalf("task.TryLock() can't lock")

	}
	if task.TryLock() {
		t.Fatalf("task.TryLock() lock already locked")
	}
}
