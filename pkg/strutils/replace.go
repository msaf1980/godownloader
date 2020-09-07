package strutils

import (
	"strings"
	"unicode"

	"github.com/mxmCherry/translit/ruicao"
	"golang.org/x/text/transform"
)

func ReplaceSpecSymbols(s string, r rune) string {
	var b strings.Builder
	for _, c := range s {
		if c == '.' || c == '_' || c == '\\' || c == '/' || c == '%' || unicode.IsDigit(c) || unicode.IsLetter(c) {
			b.WriteRune(c)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func TranslitWithoutSpecSymbols(v string, r rune) string {
	if len(v) == 0 {
		return v
	}
	ru := ruicao.ToLatin()
	s, _, _ := transform.String(ru.Transformer(), v)
	return ReplaceSpecSymbols(s, r)
}
