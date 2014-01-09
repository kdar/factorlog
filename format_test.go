package factorlog

import (
	"github.com/ngmoco/timber"
	"testing"
	"time"
)

var fmtTestsData = LogRecord{
	Time:     time.Unix(0, 1389223634000123456),
	Severity: PANIC,
	File:     "/path/to/testing.go",
	Line:     391,
	Message:  "hello there!",
	Function: "pkg.func",
	Package:  "pkg",
}

var fmtTests = []struct {
	frmt string
	out  string
}{
	{
		"%p-%P [%D]%%[%d]%%[%T][%t] [%L:%l:%F:%f:%x:%s] %%%M%%",
		"pkg-pkg.func [2014-01-08]%[2014/01/08]%[18:27:14.123456][18:27:14] [PANIC:PANC:/path/to/testing.go:testing.go:testing:391] %hello there!%\n",
	},
	{
		"",
		"",
	},
	{
		"just text here",
		"just text here\n",
	},
	{
		"%L",
		"PANIC\n",
	},
	{
		"%@",
		"%@\n",
	},
	{
		"%notsupported",
		"%notsupported\n",
	},
}

// allDigits converts an integer d to its ascii presentation,
// no matter how big the number is
// i is the deinstation index in buf
func allDigits(buf *[]byte, i, d int) int {
	j := len(*buf)
	// reverse order
	for {
		j--
		(*buf)[j] = digits[d%10]
		d /= 10
		if d == 0 {
			break
		}
	}
	return copy((*buf)[i:], (*buf)[j:])
}

func TestFormat(t *testing.T) {
	//f := newFormatter(`[%D]%%[%d]%%[%T][%t] [%L:%l:%F:%f:%x:%s] %%%M%%`)
	//fmt.Println(string(f.Format(fmtTestsData)))
	for _, tt := range fmtTests {
		f := NewStdFormatter(tt.frmt)
		out := string(f.Format(fmtTestsData))
		if tt.out != out {
			t.Fatalf("\nexpected: %#v\ngot:      %#v", tt.out, out)
		}
	}
}

func BenchmarkFormatter(b *testing.B) {
	f := NewStdFormatter("[%D %t] [%L:%f:%s] %M")
	d := LogRecord{
		Time:     time.Now(),
		Severity: PANIC,
		File:     "/path/to/testing.go",
		Line:     391,
		Message:  "hello there!",
		Function: "pkg.func",
		Package:  "pkg",
	}
	for x := 0; x < b.N; x++ {
		f.Format(d)
	}
}

func BenchmarkTimberFormatter(b *testing.B) {
	f := timber.NewPatFormatter("[%D %t] [%L:%s] %M")
	lr := &timber.LogRecord{
		Level:       timber.INFO,
		Timestamp:   time.Now(),
		SourceFile:  "/path/to/testing.go",
		SourceLine:  391,
		Message:     "hello there!",
		FuncPath:    "pkg.func",
		PackagePath: "pkg",
	}

	for x := 0; x < b.N; x++ {
		f.Format(lr)
	}
}

func BenchmarkLangAllDigits(b *testing.B) {
	var buf []byte
	tmp := make([]byte, 64)
	for x := 0; x < b.N; x++ {
		allDigits(&tmp, 0, 3456)
		buf = append(buf, tmp...)
		buf = []byte{}
	}
}

func BenchmarkLangItoa(b *testing.B) {
	var buf []byte
	tmp := make([]byte, 64)
	for x := 0; x < b.N; x++ {
		itoa(&tmp, 0, 3456)
		buf = append(buf, tmp...)
		buf = []byte{}
	}
}

func BenchmarkLangNdigits(b *testing.B) {
	var buf []byte
	tmp := make([]byte, 64)
	for x := 0; x < b.N; x++ {
		nDigits(&tmp, 4, 0, 3456)
		buf = append(buf, tmp...)
		buf = []byte{}
	}
}
