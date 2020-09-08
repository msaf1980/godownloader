package strutils

import (
	"testing"
)

func TestReplaceSpecSymbols(t *testing.T) {
	r := '_'

	tests := []struct {
		str  string
		want string
	}{
		{`t#e$s%t^:test`, "t_e_s%t__test"},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			if got := ReplaceSpecSymbols(tt.str, r); got != tt.want {
				t.Errorf("ReplaceSpecSymbols() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranslitWithoutSpecSymbols(t *testing.T) {
	r := '_'

	tests := []struct {
		str  string
		want string
	}{
		{`t#e$s%t^:Алгоритм`, "t_e_s%t__Algoritm"},
	}
	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			if got := TranslitWithoutSpecSymbols(tt.str, r); got != tt.want {
				t.Errorf("ReplaceSpecSymbols() = %v, want %v", got, tt.want)
			}
		})
	}
}
