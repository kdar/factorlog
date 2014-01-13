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
		"%{FullFunction} [%{SEVERITY}:%{SEV}:%{File}:%{ShortFile}] %%{Message}%",
		"hello there!",
		[]byte("github.com/kdar/factorlog.TestLog [ERROR:EROR:factorlog_test.go:factorlog_test] %hello there!%\n"),
	},
	{
		"%{Message} %{File}",
		"hello there!",
		[]byte("hello there! factorlog_test.go\n"),
	},
}

func TestLog(t *testing.T) {
	buf := &bytes.Buffer{}
	for _, tt := range logTests {
		buf.Reset()
		f := New(buf, NewStdFormatter(tt.frmt))
		f.Errorln(tt.in)
		if !bytes.Equal(tt.out, buf.Bytes()) {
			t.Fatalf("\nexpected: %#v\ngot:      %#v", string(tt.out), buf.String())
		}
	}
}

func TestVerbosity(t *testing.T) {
	buf := &bytes.Buffer{}
	f := New(buf, NewStdFormatter("%{Message}"))

	f.SetVerbosity(2)
	f.V(3).Info("should not appear")
	if buf.Len() > 0 {
		t.Fatal("Verbosity set to 3, Info() called with verbosity of 3. Yet, we still got a log.")
	}

	buf.Reset()
	f.SetVerbosity(4)
	f.V(3).Info("should appear")
	if buf.Len() == 0 {
		t.Fatal("Verbosity set to 4, Info() called with verbosity of 3. We should have got a log.")
	}
}

type sevTestType int

const (
	sevTest_Set sevTestType = iota
	sevTest_MinMax
)

type sevTestFunc func(l *FactorLog, v ...interface{})

var severitiesTests = []struct {
	typ     sevTestType
	min     Severity // also used for l.SetSeverities()
	max     Severity
	funName string
	fun     sevTestFunc
	output  bool
}{
	{sevTest_Set, INFO, 0, "Info", (*FactorLog).Info, true},
	{sevTest_Set, PANIC, 0, "Info", (*FactorLog).Info, false},
	{sevTest_MinMax, WARN, CRITICAL, "Info", (*FactorLog).Info, false},
	{sevTest_MinMax, WARN, CRITICAL, "Warn", (*FactorLog).Warn, true},
	{sevTest_MinMax, WARN, CRITICAL, "Error", (*FactorLog).Error, true},
	{sevTest_MinMax, WARN, CRITICAL, "Critical", (*FactorLog).Critical, true},
	{sevTest_MinMax, WARN, CRITICAL, "Stack", (*FactorLog).Stack, false},
}

func TestSeverities(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(buf, NewStdFormatter("%{Message}"))

	for _, tt := range severitiesTests {
		buf.Reset()
		if tt.typ == sevTest_Set {
			l.SetSeverities(tt.min)
			tt.fun(l, "hello")
			if tt.output && buf.Len() == 0 {
				t.Fatalf("Severity set to %s. Called %s(). We didn't get a log we expected.", UcSeverityStrings[SeverityToIndex(tt.min)], tt.funName)
			} else if !tt.output && buf.Len() > 0 {
				t.Fatalf("Severity set to %s. Called %s(). We got a log we didn't expect.", UcSeverityStrings[SeverityToIndex(tt.min)], tt.funName)
			}
		} else if tt.typ == sevTest_MinMax {
			l.SetMinMaxSeverity(tt.min, tt.max)
			tt.fun(l, "hello")
			if tt.output && buf.Len() == 0 {
				t.Fatalf("Severity set to %s-%s. Called %s(). We didn't get a log we expected.", UcSeverityStrings[SeverityToIndex(tt.min)], UcSeverityStrings[SeverityToIndex(tt.max)], tt.funName)
			} else if !tt.output && buf.Len() > 0 {
				t.Fatalf("Severity set to %s-%s. Called %s(). We got a log we didn't expect.", UcSeverityStrings[SeverityToIndex(tt.min)], UcSeverityStrings[SeverityToIndex(tt.max)], tt.funName)
			}
		}
	}
}

// Ensure `std`'s format is correct.
func TestStdFormat(t *testing.T) {
	output := std.formatter.Format(fmtTestsContext)
	expect := "2014-01-08 18:27:14 hello there!\n"
	if string(output) != expect {
		t.Fatalf("\nexpected: %#v\ngot:      %#v", expect, string(output))
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
	l := New(buf, NewStdFormatter("%{Date} %{Time} %{File}:%{Line}: %{Message}"))
	b.ResetTimer()
	for x := 0; x < b.N; x++ {
		l.Info("hey")
	}
}
