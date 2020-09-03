package urlutils

import (
	"testing"
)

func TestStripAnchor(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"http://test.int", "http://test.int"},
		{"http://test.int/index.html#1", "http://test.int/index.html"},
		{"#1", ""},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := StripAnchor(tt.url); got != tt.want {
				t.Errorf("StripAnchor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitURL(t *testing.T) {
	tests := []struct {
		url      string
		wantHost string
		wantPath string
	}{
		{"http://test.int", "http://test.int", "/"},
		{"https://test.int/", "https://test.int", "/"},
		{"test.int/", "", "test.int/"},
		{"index.html", "", "index.html"},
		{"/index.html", "", "/index.html"},
		{"/1/index.html", "", "/1/index.html"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			host, path := SplitURL(tt.url)
			if host != tt.wantHost {
				t.Errorf("SplitURL() host got = %v, want %v", host, tt.wantHost)
			}
			if path != tt.wantPath {
				t.Errorf("SplitURL() path = %v, want %v", path, tt.wantPath)
			}
		})
	}
}

func TestSplitURL3(t *testing.T) {
	tests := []struct {
		url        string
		wantScheme string
		wantHost   string
		wantPath   string
	}{
		{"http://test.int", "http", "test.int", "/"},
		{"https://test.int/", "https", "test.int", "/"},
		{"test.int/", "", "", "test.int/"},
		{"index.html", "", "", "index.html"},
		{"/index.html", "", "", "/index.html"},
		{"/1/index.html", "", "", "/1/index.html"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			scheme, host, path := SplitURL3(tt.url)
			if scheme != tt.wantScheme {
				t.Errorf("SplitURL() scheme got = %v, want %v", scheme, tt.wantScheme)
			}
			if host != tt.wantHost {
				t.Errorf("SplitURL() host got = %v, want %v", host, tt.wantHost)
			}
			if path != tt.wantPath {
				t.Errorf("SplitURL() path = %v, want %v", path, tt.wantPath)
			}
		})
	}
}

func TestAbsURL(t *testing.T) {
	baseHost := "http://127.0.0.1"
	tests := []struct {
		url        string
		wantAbsURL string
	}{
		{"http://test.int", "http://test.int"},
		{"http://test.int/", "http://test.int/"},
		{"test.int/", baseHost + "/test.int/"},
		{"index.html", baseHost + "/index.html"},
		{"/index.html", baseHost + "/index.html"},
		{"/1/index.html", baseHost + "/1/index.html"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			absURL := AbsURL(tt.url, baseHost)
			if absURL != tt.wantAbsURL {
				t.Errorf("AbsURL() got = %v, want %v", absURL, tt.wantAbsURL)
			}
		})
	}
}
