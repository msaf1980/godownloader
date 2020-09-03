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

// SplitURL split absolute url to scheme + host and path part
func SplitURL(url string) (string, string) {
	p := strings.Index(url, "://")
	if p == -1 {
		return "", url
	} else {
		p += 3
	}

	i := strings.IndexRune(url[p:], '/')
	if i == -1 {
		return url[0:], "/"
	}
	i += p
	return url[0:i], url[i:]
}

// SplitURL3 split absolute url to scheme, host and path part
func SplitURL3(url string) (string, string, string) {
	var scheme string
	p := strings.Index(url, "://")
	if p == -1 {
		return scheme, "", url
	} else {
		scheme = url[0:p]
		p += 3
	}

	i := strings.IndexRune(url[p:], '/')
	if i == -1 {
		return scheme, url[p:], "/"
	}
	i += p
	return scheme, url[p:i], url[i:]
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
