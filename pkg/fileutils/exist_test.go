package fileutils

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestExist(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "godownloader-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	tests := []struct {
		name      string
		path      string
		wantExist bool
		wantDir   bool
		wantErr   bool
	}{
		{"dir exist", tmpdir, true, true, false},
		{"not exist", tmpdir + "/not_found", false, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exist, isDir, err := Exist(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if exist != tt.wantExist {
				t.Errorf("Exist() exist = %v, want %v", exist, tt.wantExist)
			}
			if isDir != tt.wantDir {
				t.Errorf("Exist() isDir = %v, want %v", isDir, tt.wantDir)
			}
		})
	}
}
