package factorlog

import (
	"bytes"
	"github.com/mgutz/ansi"
	"path/filepath"
	"regexp"
)

type fmtVerb int

const (
	vSTRING fmtVerb = 1 << iota
	vSEVERITY
	vSeverity
	vseverity
	vSEV
	vSev
	vsev
	vS
	vs
	vDate
	vTime
	vUnix
	vUnixNano
	vFullFile
	vFile
	vShortFile
	vLine
	vFullFunction
	vPkgFunction
	vFunction
	vColor
	vMessage
	vSafeMessage
)

const (
	// If formatter.flags is set to any of these, we need runtime.Caller
	vRUNTIME_CALLER = int(vFullFile |
		vFile |
		vShortFile |
		vLine |
		vFullFunction |
		vPkgFunction |
		vFunction)
)

var (
	formatRe = regexp.MustCompile(`%{([A-Za-z]+)(?:\s(.*?[^\\]))?}`)
	verbMap  = map[string]fmtVerb{
		"SEVERITY":     vSEVERITY,
		"Severity":     vSeverity,
		"severity":     vseverity,
		"SEV":          vSEV,
		"Sev":          vSev,
		"sev":          vsev,
		"S":            vS,
		"s":            vs,
		"Date":         vDate,
		"Time":         vTime,
		"Unix":         vUnix,
		"UnixNano":     vUnixNano,
		"FullFile":     vFullFile,
		"File":         vFile,
		"ShortFile":    vShortFile,
		"Line":         vLine,
		"FullFunction": vFullFunction,
		"PkgFunction":  vPkgFunction,
		"Function":     vFunction,
		"Color":        vColor,
		"Message":      vMessage,
		"SafeMessage":  vSafeMessage,
	}
)

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
	// a slice of layouts of verbs
	layouts []string
	// temporary buffer to help in formatting.
	// initialized by newFormatter
	tmp []byte
	// temporary buffer used for safe messages.
	stmp []byte
	// flags represents all the verbs we used.
	// this is useful in speeding things up like
	// not calling runtime.Caller if we don't have
	// a format string that requires it
	flags int
}

// Available verbs:
// %{SEVERITY} - TRACE, DEBUG, INFO, WARN, ERROR, CRITICAL, STACK, FATAL, PANIC
// %{Severity} - Trace, Debug, Info, Warn, Error, Critical, Stack, Fatal, Panic
// %{severity} - trace, debug, info, warn, error, critical, stack, fatal, panic
// %{SEV}      - TRAC, DEBG, INFO, WARN, EROR, CRIT, STAK, FATL, PANC
// %{Sev}      - Trac, Debg, Info, Warn, Eror, Crit, Stak, Fatl, Panc
// %{sev}      - trac, debg, info, warn, eror, crit, stak, fatl, panc
// %{S}        - T, D, I, W, E, C, S, F, P
// %{s}        - t, d, i, w, e, c, s, f, p
// %{Date}
// %{Time}
// %{Unix}
// %{UnixNano}
// %{FullFile}
// %{File}
// %{ShortFile}
// %{Line}
// %{FullFunction}
// %{PkgFunction}
// %{Function}
// %{Color}
// %{Message}
// %{SafeMessage}
func NewStdFormatter(frmt string) *StdFormatter {
	f := &StdFormatter{
		frmt: frmt,
		tmp:  make([]byte, 64),
		stmp: make([]byte, 0, 64),
	}

	matches := formatRe.FindAllStringSubmatchIndex(frmt, -1)
	prev := 0
	for _, m := range matches {
		start, end := m[0], m[1]
		verb := frmt[m[2]:m[3]]

		layout := ""
		if m[4] != -1 {
			layout = frmt[m[4]:m[5]]
		}

		if start > prev {
			f.appendString(frmt[prev:start])
		}

		if v, ok := verbMap[verb]; ok {
			// Colors are special and can be processed now
			if v == vColor {
				if layout == "reset" {
					f.appendString(ansi.Reset)
				} else {
					code := ansi.ColorCode(layout)
					f.appendString(code)
				}
			} else {
				f.flags |= int(v)
				f.parts = append(f.parts, v)
			}
		}

		prev = end
	}

	if frmt[prev:] != "" {
		f.appendString(frmt[prev:])
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

func (f *StdFormatter) Format(context LogContext) []byte {
	buf := &bytes.Buffer{}
	stringi := 0
	for _, p := range f.parts {
		switch p {
		case vSTRING:
			buf.WriteString(f.strings[stringi])
			stringi++
		case vSEVERITY:
			buf.WriteString(UcSeverityStrings[context.Severity])
		case vSeverity:
			buf.WriteString(CapSeverityStrings[context.Severity])
		case vseverity:
			buf.WriteString(LcSeverityStrings[context.Severity])
		case vSEV:
			buf.WriteString(UcShortSeverityStrings[context.Severity])
		case vSev:
			buf.WriteString(CapShortSeverityStrings[context.Severity])
		case vsev:
			buf.WriteString(LcShortSeverityStrings[context.Severity])
		case vS:
			buf.WriteString(UcShortestSeverityStrings[context.Severity])
		case vs:
			buf.WriteString(LcShortestSeverityStrings[context.Severity])
		case vDate:
			year, month, day := context.Time.Date()
			nDigits(&f.tmp, 4, 0, year)
			f.tmp[4] = '-'
			twoDigits(&f.tmp, 5, int(month))
			f.tmp[7] = '-'
			twoDigits(&f.tmp, 8, day)
			buf.Write(f.tmp[:10])
		case vTime:
			hour, min, sec := context.Time.Clock()
			twoDigits(&f.tmp, 0, hour)
			f.tmp[2] = ':'
			twoDigits(&f.tmp, 3, min)
			f.tmp[5] = ':'
			twoDigits(&f.tmp, 6, sec)
			buf.Write(f.tmp[:8])
		case vUnix:
			n := i64toa(&f.tmp, 0, context.Time.Unix())
			buf.Write(f.tmp[:n])
		case vUnixNano:
			n := i64toa(&f.tmp, 0, context.Time.UnixNano())
			buf.Write(f.tmp[:n])
		case vFullFile:
			buf.WriteString(context.File)
		case vFile, vShortFile:
			file := context.File
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
			}

			if p == vShortFile {
				file = file[:len(file)-3]
			}

			buf.WriteString(file)
		case vLine:
			n := itoa(&f.tmp, 0, context.Line)
			buf.Write(f.tmp[:n])
		case vFullFunction:
			buf.WriteString(context.Function)
		case vPkgFunction:
			fun := context.Function
			slash := len(fun) - 1
			for ; slash >= 0; slash-- {
				if fun[slash] == filepath.Separator {
					break
				}
			}
			if slash >= 0 {
				fun = fun[slash+1:]
			}

			buf.WriteString(fun)
		case vFunction:
			fun := context.Function

			slash := len(fun) - 1
			lastDot := -1
			for ; slash >= 0; slash-- {
				if fun[slash] == filepath.Separator {
					break
				} else if fun[slash] == '.' {
					lastDot = slash
				}
			}

			fun = fun[lastDot+1:]
			buf.WriteString(fun)
		case vMessage:
			buf.WriteString(context.Message)
		case vSafeMessage:
			f.stmp = f.stmp[:0]
			l := len(context.Message)
			ca := cap(f.stmp)
			if l > ca {
				f.stmp = make([]byte, 0, l)
			} else if ca > 8000 { // don't let memory usage get too big
				f.stmp = f.stmp[0:0:l]
			}

			for _, c := range context.Message {
				if int(c) < 32 {
					f.tmp[0] = '\\'
					f.tmp[1] = 'x'
					twoDigits(&f.tmp, 2, int(c))
					f.stmp = append(f.stmp, f.tmp[:4]...)
				} else {
					f.stmp = append(f.stmp, byte(c))
				}
			}
			buf.Write(f.stmp)
		}
	}

	b := buf.Bytes()
	if buf.Len() > 0 && b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}

	return b
}
