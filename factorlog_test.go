package factorlog

import (
	"bytes"
	"log"
	"testing"
)

var (
	// Test to make sure these types satisfy the Logger interface.
	_ Logger = &FactorLog{}
	_ Logger = Verbose{}
	// too bad this doesn't work
	//_ Logger = factorlog
)

var logTests = []struct {
	frmt string
	in   string
	out  []byte
}{
	{
		// we can't use every verb here, because the test will fail
		"%p-%P [%L:%l:%f:%x] %%%M%%",
		"hello there!",
		[]byte("github.com/kdar/factorlog-github.com/kdar/factorlog.TestLog [ERROR:EROR:factorlog_test.go:factorlog_test] %hello there!%\n"),
	},
}

func TestLog(t *testing.T) {
	buf := &bytes.Buffer{}
	for _, tt := range logTests {
		f := New(buf, tt.frmt)
		f.Error(tt.in)
		if !bytes.Equal(tt.out, buf.Bytes()) {
			t.Fatalf("\nexpected: %#v\ngot:      %#v", string(tt.out), buf.String())
		}
	}
}

func TestVerbosity(t *testing.T) {
	buf := &bytes.Buffer{}
	f := New(buf, "%M")
	f.SetVerbosity(2)
	f.V(3).Info("should not appear")
	if buf.Len() > 0 {
		t.Fatal("Verbosity set to 3, Info() called with verbosity of 3. Yet, we still got a log.")
	}

	f.SetVerbosity(4)
	f.V(3).Info("should appear")
	if buf.Len() == 0 {
		t.Fatal("Verbosity set to 4, Info() called with verbosity of 3. We should have got a log.")
	}
}

func BenchmarkGoLogBuffer(b *testing.B) {
	buf := &bytes.Buffer{}
	l := log.New(buf, "", log.Ldate|log.Ltime|log.Lshortfile)
	b.ResetTimer()
	for x := 0; x < b.N; x++ {
		l.Print("hey")
	}
}

func BenchmarkFactorLogBuffer(b *testing.B) {
	buf := &bytes.Buffer{}
	l := New(buf, "%d %t %f:%s: %M")
	b.ResetTimer()
	for x := 0; x < b.N; x++ {
		l.Info("hey")
	}
}
