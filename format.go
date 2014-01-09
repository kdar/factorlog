package factorlog

import (
	"path/filepath"
	"time"
)

type fmtVerb int

const (
	vSTRING fmtVerb = 1 << iota
	vDATE_D
	vDATE_d
	vTIME_T
	vTIME_t
	vSEVERITY
	vSHORT_SEVERITY
	vFILE
	vSHORT_FILE
	vEXTRA_SHORT_FILE
	vLINE
	vMESSAGE
	vPACKAGE_FUNCTION
	vFUNCTION
)

const (
	// If formatter.flags is set to any of these, we need runtime.Caller
	vRUNTIME_CALLER = int(vPACKAGE_FUNCTION | vFUNCTION | vFILE | vSHORT_FILE | vEXTRA_SHORT_FILE | vLINE)
)

// Interface to format anything
type Formatter interface {
	// Formats LogRecord and returns the []byte that will
	// be written by the log. This is not inherently thread
	// safe but FactorLog uses a mutex before calling this.
	Format(data LogRecord) []byte

	// Returns true if we should call runtime.Caller because
	// we have a format that requires it. We do this because
	// it is expensive.
	ShouldRuntimeCaller() bool
}

type StdFormatter struct {
	// the original format
	frmt string
	// all the strings that are not verbs
	// e.g. "hey %D there" will give us
	// []string{"hey ", " there"}
	strings []string
	// a slice depicting each part of the format
	// we build the final []byte from this
	parts []fmtVerb
	// temporary buffer to help in formatting.
	// initialized by newFormatter
	tmp []byte
	// flags represents all the verbs we used
	// this is useful in speeding things up like
	// not calling runtime.Caller if we don't have
	// a format string that requires it
	flags int
}

// parse a format string and return a formatter
func NewStdFormatter(frmt string) *StdFormatter {
	f := &StdFormatter{
		frmt: frmt,
		tmp:  make([]byte, 64),
	}

	isverb := false
	var raw []byte
	for _, c := range frmt {
		if !isverb && c == '%' {
			if len(raw) > 0 {
				f.strings = append(f.strings, string(raw))
				f.parts = append(f.parts, vSTRING)
				raw = []byte{}
			}

			isverb = true
			continue
		}

		if isverb {
			switch c {
			case 'T':
				f.flags |= int(vTIME_T)
				f.parts = append(f.parts, vTIME_T)
			case 't':
				f.flags |= int(vTIME_t)
				f.parts = append(f.parts, vTIME_t)
			case 'D':
				f.flags |= int(vDATE_D)
				f.parts = append(f.parts, vDATE_D)
			case 'd':
				f.flags |= int(vDATE_d)
				f.parts = append(f.parts, vDATE_d)
			case 'L':
				f.flags |= int(vSEVERITY)
				f.parts = append(f.parts, vSEVERITY)
			case 'l':
				f.flags |= int(vSHORT_SEVERITY)
				f.parts = append(f.parts, vSHORT_SEVERITY)
			case 'F':
				f.flags |= int(vFILE)
				f.parts = append(f.parts, vFILE)
			case 'f':
				f.flags |= int(vSHORT_FILE)
				f.parts = append(f.parts, vSHORT_FILE)
			case 'x':
				f.flags |= int(vEXTRA_SHORT_FILE)
				f.parts = append(f.parts, vEXTRA_SHORT_FILE)
			case 's':
				f.flags |= int(vLINE)
				f.parts = append(f.parts, vLINE)
			case 'M':
				f.flags |= int(vMESSAGE)
				f.parts = append(f.parts, vMESSAGE)
			case 'P':
				f.flags |= int(vPACKAGE_FUNCTION)
				f.parts = append(f.parts, vPACKAGE_FUNCTION)
			case 'p':
				f.flags |= int(vFUNCTION)
				f.parts = append(f.parts, vFUNCTION)
			case '%':
				raw = append(raw, '%')
			default:
				raw = append(raw, '%')
				raw = append(raw, byte(c))
			}
		} else {
			raw = append(raw, byte(c))
		}

		isverb = false
	}

	if len(raw) > 0 {
		f.strings = append(f.strings, string(raw))
		f.parts = append(f.parts, vSTRING)
	}

	return f
}

func (f *StdFormatter) ShouldRuntimeCaller() bool {
	return f.flags&(vRUNTIME_CALLER) != 0
}

func (f *StdFormatter) appendString(s string) {
	if len(s) > 0 {
		f.strings = append(f.strings, s)
		f.parts = append(f.parts, vSTRING)
	}
}

// structure used to hold the data used for formatting
type LogRecord struct {
	Time     time.Time
	Severity Severity
	File     string
	Line     int
	Function string
	Package  string
	Message  string
}

// format the data and return a []byte.
// this is exactly what is written to a log
func (f *StdFormatter) Format(data LogRecord) []byte {
	var buf []byte
	stringi := 0
	for _, p := range f.parts {
		switch p {
		case vSTRING:
			buf = append(buf, f.strings[stringi]...)
			stringi++
		case vTIME_T:
			hour, min, sec := data.Time.Clock()
			twoDigits(&f.tmp, 0, hour)
			f.tmp[2] = ':'
			twoDigits(&f.tmp, 3, min)
			f.tmp[5] = ':'
			twoDigits(&f.tmp, 6, sec)
			f.tmp[8] = '.'
			nDigits(&f.tmp, 6, 9, data.Time.Nanosecond())
			buf = append(buf, f.tmp[:15]...)
		case vTIME_t:
			hour, min, sec := data.Time.Clock()
			twoDigits(&f.tmp, 0, hour)
			f.tmp[2] = ':'
			twoDigits(&f.tmp, 3, min)
			f.tmp[5] = ':'
			twoDigits(&f.tmp, 6, sec)
			buf = append(buf, f.tmp[:8]...)
		case vDATE_D:
			year, month, day := data.Time.Date()
			nDigits(&f.tmp, 4, 0, year)
			f.tmp[4] = '-'
			twoDigits(&f.tmp, 5, int(month))
			f.tmp[7] = '-'
			twoDigits(&f.tmp, 8, day)
			buf = append(buf, f.tmp[:10]...)
		case vDATE_d:
			year, month, day := data.Time.Date()
			nDigits(&f.tmp, 4, 0, year)
			f.tmp[4] = '/'
			twoDigits(&f.tmp, 5, int(month))
			f.tmp[7] = '/'
			twoDigits(&f.tmp, 8, day)
			buf = append(buf, f.tmp[:10]...)
		case vSEVERITY:
			buf = append(buf, LongSeverityStrings[data.Severity]...)
		case vSHORT_SEVERITY:
			buf = append(buf, SeverityStrings[data.Severity]...)
		case vFILE:
			buf = append(buf, data.File...)
		case vSHORT_FILE, vEXTRA_SHORT_FILE:
			file := data.File
			if len(file) == 0 {
				file = "???"
			} else {
				slash := len(file) - 1
				for ; slash >= 0; slash-- {
					if file[slash] == filepath.Separator {
						break
					}
				}
				if slash >= 0 {
					file = file[slash+1:]
				}

				if p == vEXTRA_SHORT_FILE {
					file = file[:len(file)-3]
				}
			}

			buf = append(buf, file...)
		case vLINE:
			n := itoa(&f.tmp, 0, data.Line)
			buf = append(buf, f.tmp[:n]...)
		case vMESSAGE:
			buf = append(buf, data.Message...)
		case vPACKAGE_FUNCTION:
			buf = append(buf, data.Function...)
		case vFUNCTION:
			buf = append(buf, data.Package...)
		}
	}

	if len(buf) > 0 && buf[len(buf)-1] != '\n' {
		buf = append(buf, '\n')
	}

	return buf
}
