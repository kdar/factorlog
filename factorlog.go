package factorlog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Severity int

const (
	NONE Severity = iota // NONE to be used for standard go log impl's
	TRACE
	DEBUG
	INFO
	WARN
	ERROR
	CRITICAL
	STACK
	FATAL
	PANIC
)

var SeverityStrings = [...]string{"", "TRAC", "DEBG", "INFO", "WARN", "EROR", "CRIT", "STAK", "FATL", "PANC"}

var LongSeverityStrings = [...]string{
	"NONE",
	"TRACE",
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"CRITICAL",
	"STACK",
	"FATAL",
	"PANIC",
}

func SeverityFromString(l string) Severity {
	switch l {
	case "TRACE":
		return TRACE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "CRITICAL":
		return CRITICAL
	case "STACK":
		return STACK
	case "FATAL":
		return FATAL
	case "PANIC":
		return PANIC
	}

	return NONE
}

type Logger interface {
	Trace(v ...interface{})
	Tracef(format string, v ...interface{})
	Traceln(v ...interface{})
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Debugln(v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Infoln(v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Warnln(v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Errorln(v ...interface{})
	Critical(v ...interface{})
	Criticalf(format string, v ...interface{})
	Criticalln(v ...interface{})
	Stack(v ...interface{})
	Stackf(format string, v ...interface{})
	Stackln(v ...interface{})
	Log(sev Severity, v ...interface{})

	// golang's log interface
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
}

type Level int32

func (l *Level) get() Level {
	return Level(atomic.LoadInt32((*int32)(l)))
}

func (l *Level) set(val Level) {
	atomic.StoreInt32((*int32)(l), int32(val))
}

type FactorLog struct {
	mu        sync.Mutex // ensures atomic writes; protects the following fields
	out       io.Writer  // destination for output
	formatter Formatter
	verbosity Level
}

// New creates a FactorLog with the given output and format.
func New(out io.Writer, format string) *FactorLog {
	return &FactorLog{out: out, formatter: NewStdFormatter(format)}
}

// just like Go's log.std
var std = New(os.Stderr, "%d %t %M")

// Sets the verbosity level of this log. Use IsV() or V() to
// utilize verbosity.
func (l *FactorLog) SetVerbosity(level Level) {
	l.verbosity.set(level)
}

// Output will write to the writer with the given severity, calldepth,
// and string. calldepth is only used if the format requires a call to
// runtime.Caller.
func (l *FactorLog) Output(sev Severity, calldepth int, s string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	data := LogRecord{
		Time:     time.Now(),
		Severity: sev,
		Message:  s,
	}

	if l.formatter.ShouldRuntimeCaller() {
		// release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		pc, file, line, ok := runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		} else {
			me := runtime.FuncForPC(pc)
			if me != nil {
				data.Function = me.Name()
				n := strings.LastIndex(data.Function, ".")
				if n != -1 {
					data.Package = data.Function[:n]
				}
			}
		}

		data.File = file
		data.Line = line

		l.mu.Lock()
	}

	_, err := l.out.Write(l.formatter.Format(data))
	return err
}

// SetOutput sets the output destination for thislogger.
func (l *FactorLog) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

// IsV tests whether the verbosity is of a certain level.
// Returns a bool.
// Example:
//    if log.IsV(2) {
//      log.Info("some info")
//    }
func (l *FactorLog) IsV(level Level) bool {
	if l.verbosity.get() >= level {
		return true
	}

	return false
}

// V tests whether the verbosity is of a certain level,
// and returns a Verbose object that allows you to
// chain calls. This is a convenience function and should
// be avoided if you care about raw performance (use IsV()
// instead).
// Example:
//   log.V(2).Info("some info")
func (l *FactorLog) V(level Level) Verbose {
	if l.verbosity.get() >= level {
		return Verbose{true, l}
	}

	return Verbose{false, l}
}

func (l *FactorLog) Trace(v ...interface{}) {
	l.Output(TRACE, 2, fmt.Sprint(v...))
}

func (l *FactorLog) Tracef(format string, v ...interface{}) {
	l.Output(TRACE, 2, fmt.Sprintf(format, v...))
}

func (l *FactorLog) Traceln(v ...interface{}) {
	l.Output(TRACE, 2, fmt.Sprintln(v...))
}

func (l *FactorLog) Debug(v ...interface{}) {
	l.Output(DEBUG, 2, fmt.Sprint(v...))
}

func (l *FactorLog) Debugf(format string, v ...interface{}) {
	l.Output(DEBUG, 2, fmt.Sprintf(format, v...))
}

func (l *FactorLog) Debugln(v ...interface{}) {
	l.Output(DEBUG, 2, fmt.Sprintln(v...))
}

func (l *FactorLog) Info(v ...interface{}) {
	l.Output(INFO, 2, fmt.Sprint(v...))
}

func (l *FactorLog) Infof(format string, v ...interface{}) {
	l.Output(INFO, 2, fmt.Sprintf(format, v...))
}

func (l *FactorLog) Infoln(v ...interface{}) {
	l.Output(INFO, 2, fmt.Sprintln(v...))
}

func (l *FactorLog) Warn(v ...interface{}) {
	l.Output(WARN, 2, fmt.Sprint(v...))
}

func (l *FactorLog) Warnf(format string, v ...interface{}) {
	l.Output(WARN, 2, fmt.Sprintf(format, v...))
}

func (l *FactorLog) Warnln(v ...interface{}) {
	l.Output(WARN, 2, fmt.Sprintln(v...))
}

func (l *FactorLog) Error(v ...interface{}) {
	l.Output(ERROR, 2, fmt.Sprint(v...))
}

func (l *FactorLog) Errorf(format string, v ...interface{}) {
	l.Output(ERROR, 2, fmt.Sprintf(format, v...))
}

func (l *FactorLog) Errorln(v ...interface{}) {
	l.Output(ERROR, 2, fmt.Sprintln(v...))
}

func (l *FactorLog) Critical(v ...interface{}) {
	l.Output(CRITICAL, 2, fmt.Sprint(v...))
}

func (l *FactorLog) Criticalf(format string, v ...interface{}) {
	l.Output(CRITICAL, 2, fmt.Sprintf(format, v...))
}

func (l *FactorLog) Criticalln(v ...interface{}) {
	l.Output(CRITICAL, 2, fmt.Sprintln(v...))
}

// Stack is equivalent to Print() followed by printing a stack
// trace to the configured writer.
func (l *FactorLog) Stack(v ...interface{}) {
	l.Output(STACK, 2, fmt.Sprint(v...))
	l.out.Write(stacks(true))
}

// Stackf is equivalent to Printf() followed by printing a stack
// trace to the configured writer.
func (l *FactorLog) Stackf(format string, v ...interface{}) {
	l.Output(STACK, 2, fmt.Sprintf(format, v...))
	l.out.Write(stacks(true))
}

// Stackln is equivalent to Println() followed by printing a stack
// trace to the configured writer.
func (l *FactorLog) Stackln(v ...interface{}) {
	l.Output(STACK, 2, fmt.Sprintln(v...))
	l.out.Write(stacks(true))
}

func (l *FactorLog) Log(sev Severity, v ...interface{}) {
	l.Output(sev, 2, fmt.Sprint(v...))
}

func (l *FactorLog) Print(v ...interface{}) {
	l.Output(DEBUG, 2, fmt.Sprint(v...))
}

func (l *FactorLog) Printf(format string, v ...interface{}) {
	l.Output(DEBUG, 2, fmt.Sprintf(format, v...))
}

func (l *FactorLog) Println(v ...interface{}) {
	l.Output(DEBUG, 2, fmt.Sprintln(v...))
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func (l *FactorLog) Fatal(v ...interface{}) {
	l.Output(FATAL, 2, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func (l *FactorLog) Fatalf(format string, v ...interface{}) {
	l.Output(FATAL, 2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func (l *FactorLog) Fatalln(v ...interface{}) {
	l.Output(FATAL, 2, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic is equivalent to Print() followed by a call to panic().
func (l *FactorLog) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.Output(PANIC, 2, s)
	panic(s)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func (l *FactorLog) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.Output(PANIC, 2, s)
	panic(s)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func (l *FactorLog) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.Output(PANIC, 2, s)
	panic(s)
}

// Verbose is a structure that enables syntatic sugar
// when testing for verbosity and calling a log function.
// See FactorLog.V().
type Verbose struct {
	True   bool
	logger *FactorLog
}

func (b Verbose) Trace(v ...interface{}) {
	if b.True {
		b.logger.Output(TRACE, 2, fmt.Sprint(v...))
	}
}

func (b Verbose) Tracef(format string, v ...interface{}) {
	if b.True {
		b.logger.Output(TRACE, 2, fmt.Sprintf(format, v...))
	}
}

func (b Verbose) Traceln(v ...interface{}) {
	if b.True {
		b.logger.Output(TRACE, 2, fmt.Sprintln(v...))
	}
}

func (b Verbose) Debug(v ...interface{}) {
	if b.True {
		b.logger.Output(DEBUG, 2, fmt.Sprint(v...))
	}
}

func (b Verbose) Debugf(format string, v ...interface{}) {
	if b.True {
		b.logger.Output(DEBUG, 2, fmt.Sprintf(format, v...))
	}
}

func (b Verbose) Debugln(v ...interface{}) {
	if b.True {
		b.logger.Output(DEBUG, 2, fmt.Sprintln(v...))
	}
}

func (b Verbose) Info(v ...interface{}) {
	if b.True {
		b.logger.Output(INFO, 2, fmt.Sprint(v...))
	}
}

func (b Verbose) Infof(format string, v ...interface{}) {
	if b.True {
		b.logger.Output(INFO, 2, fmt.Sprintf(format, v...))
	}
}

func (b Verbose) Infoln(v ...interface{}) {
	if b.True {
		b.logger.Output(INFO, 2, fmt.Sprintln(v...))
	}
}

func (b Verbose) Warn(v ...interface{}) {
	if b.True {
		b.logger.Output(WARN, 2, fmt.Sprint(v...))
	}
}

func (b Verbose) Warnf(format string, v ...interface{}) {
	if b.True {
		b.logger.Output(WARN, 2, fmt.Sprintf(format, v...))
	}
}

func (b Verbose) Warnln(v ...interface{}) {
	if b.True {
		b.logger.Output(WARN, 2, fmt.Sprintln(v...))
	}
}

func (b Verbose) Error(v ...interface{}) {
	if b.True {
		b.logger.Output(ERROR, 2, fmt.Sprint(v...))
	}
}

func (b Verbose) Errorf(format string, v ...interface{}) {
	if b.True {
		b.logger.Output(ERROR, 2, fmt.Sprintf(format, v...))
	}
}

func (b Verbose) Errorln(v ...interface{}) {
	if b.True {
		b.logger.Output(ERROR, 2, fmt.Sprintln(v...))
	}
}

func (b Verbose) Critical(v ...interface{}) {
	if b.True {
		b.logger.Output(CRITICAL, 2, fmt.Sprint(v...))
	}
}

func (b Verbose) Criticalf(format string, v ...interface{}) {
	if b.True {
		b.logger.Output(CRITICAL, 2, fmt.Sprintf(format, v...))
	}
}

func (b Verbose) Criticalln(v ...interface{}) {
	if b.True {
		b.logger.Output(CRITICAL, 2, fmt.Sprintln(v...))
	}
}

func (b Verbose) Stack(v ...interface{}) {
	if b.True {
		b.logger.Output(STACK, 2, fmt.Sprint(v...))
		b.logger.out.Write(stacks(true))
	}
}

func (b Verbose) Stackf(format string, v ...interface{}) {
	if b.True {
		b.logger.Output(STACK, 2, fmt.Sprintf(format, v...))
		b.logger.out.Write(stacks(true))
	}
}

func (b Verbose) Stackln(v ...interface{}) {
	if b.True {
		b.logger.Output(STACK, 2, fmt.Sprintln(v...))
		b.logger.out.Write(stacks(true))
	}
}

func (b Verbose) Log(sev Severity, v ...interface{}) {
	if b.True {
		b.logger.Output(sev, 2, fmt.Sprint(v...))
	}
}

func (b Verbose) Print(v ...interface{}) {
	if b.True {
		b.logger.Output(DEBUG, 2, fmt.Sprint(v...))
	}
}

func (b Verbose) Printf(format string, v ...interface{}) {
	if b.True {
		b.logger.Output(DEBUG, 2, fmt.Sprintf(format, v...))
	}
}

func (b Verbose) Println(v ...interface{}) {
	if b.True {
		b.logger.Output(DEBUG, 2, fmt.Sprintln(v...))
	}
}

func (b Verbose) Fatal(v ...interface{}) {
	if b.True {
		b.logger.Output(FATAL, 2, fmt.Sprint(v...))
		os.Exit(1)
	}
}

func (b Verbose) Fatalf(format string, v ...interface{}) {
	if b.True {
		b.logger.Output(FATAL, 2, fmt.Sprintf(format, v...))
		os.Exit(1)
	}
}

func (b Verbose) Fatalln(v ...interface{}) {
	if b.True {
		b.logger.Output(FATAL, 2, fmt.Sprintln(v...))
		os.Exit(1)
	}
}

func (b Verbose) Panic(v ...interface{}) {
	if b.True {
		s := fmt.Sprint(v...)
		b.logger.Output(PANIC, 2, s)
		panic(s)
	}
}

func (b Verbose) Panicf(format string, v ...interface{}) {
	if b.True {
		s := fmt.Sprintf(format, v...)
		b.logger.Output(PANIC, 2, s)
		panic(s)
	}
}

func (b Verbose) Panicln(v ...interface{}) {
	if b.True {
		s := fmt.Sprintln(v...)
		b.logger.Output(PANIC, 2, s)
		panic(s)
	}
}

// Global functions for the package. Uses a standard
// logger just like Go's log package.

// SetOutput sets the output destination for the standard logger.
func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = w
}

func SetVerbosity(level Level) {
	std.verbosity.set(level)
}

func IsV(level Level) bool {
	if std.verbosity.get() >= level {
		return true
	}

	return false
}

func V(level Level) Verbose {
	if std.verbosity.get() >= level {
		return Verbose{true, std}
	}

	return Verbose{false, std}
}

func Trace(v ...interface{}) {
	std.Output(TRACE, 2, fmt.Sprint(v...))
}

func Tracef(format string, v ...interface{}) {
	std.Output(TRACE, 2, fmt.Sprintf(format, v...))
}

func Traceln(v ...interface{}) {
	std.Output(TRACE, 2, fmt.Sprintln(v...))
}

func Debug(v ...interface{}) {
	std.Output(DEBUG, 2, fmt.Sprint(v...))
}

func Debugf(format string, v ...interface{}) {
	std.Output(DEBUG, 2, fmt.Sprintf(format, v...))
}

func Debugln(v ...interface{}) {
	std.Output(DEBUG, 2, fmt.Sprintln(v...))
}

func Info(v ...interface{}) {
	std.Output(INFO, 2, fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
	std.Output(INFO, 2, fmt.Sprintf(format, v...))
}

func Infoln(v ...interface{}) {
	std.Output(INFO, 2, fmt.Sprintln(v...))
}

func Warn(v ...interface{}) {
	std.Output(WARN, 2, fmt.Sprint(v...))
}

func Warnf(format string, v ...interface{}) {
	std.Output(WARN, 2, fmt.Sprintf(format, v...))
}

func Warnln(v ...interface{}) {
	std.Output(WARN, 2, fmt.Sprintln(v...))
}

func Error(v ...interface{}) {
	std.Output(ERROR, 2, fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
	std.Output(ERROR, 2, fmt.Sprintf(format, v...))
}

func Errorln(v ...interface{}) {
	std.Output(ERROR, 2, fmt.Sprintln(v...))
}

func Critical(v ...interface{}) {
	std.Output(CRITICAL, 2, fmt.Sprint(v...))
}

func Criticalf(format string, v ...interface{}) {
	std.Output(CRITICAL, 2, fmt.Sprintf(format, v...))
}

func Criticalln(v ...interface{}) {
	std.Output(CRITICAL, 2, fmt.Sprintln(v...))
}

func Stack(v ...interface{}) {
	std.Output(STACK, 2, fmt.Sprint(v...))
	std.out.Write(stacks(true))
}

func Stackf(format string, v ...interface{}) {
	std.Output(STACK, 2, fmt.Sprintf(format, v...))
	std.out.Write(stacks(true))
}

func Stackln(v ...interface{}) {
	std.Output(STACK, 2, fmt.Sprintln(v...))
	std.out.Write(stacks(true))
}

func Log(sev Severity, v ...interface{}) {
	std.Output(sev, 2, fmt.Sprint(v...))
}

func Print(v ...interface{}) {
	std.Output(DEBUG, 2, fmt.Sprint(v...))
}

func Printf(format string, v ...interface{}) {
	std.Output(DEBUG, 2, fmt.Sprintf(format, v...))
}

func Println(v ...interface{}) {
	std.Output(DEBUG, 2, fmt.Sprintln(v...))
}

func Fatal(v ...interface{}) {
	std.Output(FATAL, 2, fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	std.Output(FATAL, 2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	std.Output(FATAL, 2, fmt.Sprintln(v...))
	os.Exit(1)
}

func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	std.Output(PANIC, 2, s)
	panic(s)
}

func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	std.Output(PANIC, 2, s)
	panic(s)
}

func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	std.Output(PANIC, 2, s)
	panic(s)
}
