package htmlutils

import (
	"bufio"
	"bytes"
	"io"
	"os"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/htmlindex"
)

func detectBytesCharset(b []byte) string {
	n := 1024
	if len(b) < n {
		n = len(b)
	}
	if _, name, ok := charset.DetermineEncoding(b[0:n], ""); ok {
		return name
	}
	return "utf-8"
}

func detectFileCharset(f *os.File) string {
	data := make([]byte, 1024)
	n, err := f.Read(data)
	f.Seek(0, 0)
	if err == nil {
		if _, name, ok := charset.DetermineEncoding(data[0:n], ""); ok {
			return name
		}
	}
	return "utf-8"
}

func detectContentCharset(body io.Reader) string {
	r := bufio.NewReader(body)
	if data, err := r.Peek(1024); err == nil {
		if _, name, ok := charset.DetermineEncoding(data, ""); ok {
			return name
		}
	}
	return "utf-8"
}

// DecodeHTMLBody returns an decoding reader of the html Body for the specified `charset`
// If `charset` is empty, DecodeHTMLBody tries to guess the encoding from the content
func DecodeHTMLBody(body io.Reader, charset string) (io.Reader, string, error) {
	if charset == "" {
		charset = detectContentCharset(body)
	}
	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, "", err
	}
	var name string
	if name, _ = htmlindex.Name(e); name != "utf-8" {
		return e.NewDecoder().Reader(body), name, nil
	}
	return body, name, nil
}

// DecodeHTMLFile returns an decoding reader of the os.File for the specified `charset`
// If `charset` is empty, DecodeHTMLFile tries to guess the encoding from the content
func DecodeHTMLFile(f *os.File, charset string) (io.Reader, string, error) {
	if charset == "" {
		charset = detectFileCharset(f)
	}
	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, "", err
	}
	var name string
	if name, _ = htmlindex.Name(e); name != "utf-8" {
		return e.NewDecoder().Reader(f), name, nil
	}
	return f, name, nil
}

func DecodeHTMLBytes(b []byte, charset string) (string, string, error) {
	if charset == "" {
		charset = detectBytesCharset(b)
	}
	e, err := htmlindex.Get(charset)
	if err != nil {
		return "", "", err
	}
	var name string
	if name, _ = htmlindex.Name(e); name != "utf-8" {
		conv, err := e.NewDecoder().Bytes(b)
		if err != nil {
			return "", "", err
		}
		return string(conv), name, nil
	}
	return string(b), name, nil
}

func DecodeHTMLReader(b []byte, charset string) (io.Reader, string, error) {
	if charset == "" {
		charset = detectBytesCharset(b)
	}
	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, "", err
	}
	var name string
	if name, _ = htmlindex.Name(e); name != "utf-8" {
		r := e.NewDecoder().Reader(bytes.NewReader(b))
		return r, name, nil
	}
	return bytes.NewReader(b), name, nil
}
