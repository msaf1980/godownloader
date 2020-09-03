package urlutils

import (
	"strings"
)

// StripAnchor strip anchor (#...) from url
func StripAnchor(url string) string {
	i := strings.LastIndex(url, "#")
	if i == -1 {
		return url
	}
	return url[0:i]
}

// SplitURL split absolute url to host and path part
func SplitURL(url string) (string, string) {
	p := strings.Index(url, "://")
	if p == -1 {
		return "", url
	} else {
		p += 3
	}
	i := strings.IndexRune(url[p:], '/')
	if i == -1 {
		return url, "/"
	}
	i += p
	return url[0:i], url[i:]
}

// AbsURL check if url is absolute and if not, append base host
func AbsURL(url string, baseHost string) string {
	p := strings.Index(url, "://")
	if p == -1 {
		if len(url) > 0 && url[0] == '/' {
			return baseHost + url
		}
		return baseHost + "/" + url
	}
	return url
}
