package factorlog

import (
	"testing"
)

func TestGlogFormatter(t *testing.T) {
	f := NewGlogFormatter()
	expect := "P0108 18:27:14.123456 01234 testing.go:391] hello there!\n"
	out := string(f.Format(fmtTestsContext))
	if expect != out {
		t.Fatalf("\nexpected: %#v\ngot:      %#v", expect, out)
	}
}

func BenchmarkGlogFormatter(b *testing.B) {
	f := NewGlogFormatter()

	b.ResetTimer()
	for x := 0; x < b.N; x++ {
		f.Format(fmtTestsContext)
	}
}
