package factorlog

import (
	"bytes"
	"fmt"
	"path/filepath"
)

type GlogFormatter struct {
	tmp []byte
}

func NewGlogFormatter() *GlogFormatter {
	return &GlogFormatter{make([]byte, 64)}
}

// This will always be true.
func (f *GlogFormatter) ShouldRuntimeCaller() bool {
	return true
}

// Log lines have this form:
//         Lmmdd hh:mm:ss.uuuuuu threadid file:line] msg...
// where the fields are defined as follows:
//         L                A single character, representing the log level (eg 'I' for INFO)
//         mm               The month (zero padded; ie May is '05')
//         dd               The day (zero padded)
//         hh:mm:ss.uuuuuu  Time in hours, minutes and fractional seconds
//         threadid         The space-padded thread ID as returned by GetTID()
//         file             The file name
//         line             The line number
//         msg              The user-supplied message
func (f *GlogFormatter) Format(context LogContext) []byte {
	res := &bytes.Buffer{}

	file := context.File
	slash := len(file) - 1
	for ; slash >= 0; slash-- {
		if file[slash] == filepath.Separator {
			break
		}
	}
	if slash >= 0 {
		file = file[slash+1:]
	}

	_, month, day := context.Time.Date()
	hour, minute, second := context.Time.Clock()
	f.tmp[0] = UcShortestSeverityStrings[SeverityToIndex(context.Severity)][0]
	TwoDigits(&f.tmp, 1, int(month))
	TwoDigits(&f.tmp, 3, day)
	f.tmp[5] = ' '
	TwoDigits(&f.tmp, 6, hour)
	f.tmp[8] = ':'
	TwoDigits(&f.tmp, 9, minute)
	f.tmp[11] = ':'
	TwoDigits(&f.tmp, 12, second)
	f.tmp[14] = '.'
	NDigits(&f.tmp, 6, 15, context.Time.Nanosecond()/1000)
	f.tmp[21] = ' '
	NDigits(&f.tmp, 5, 22, context.Pid)
	f.tmp[27] = ' '
	res.Write(f.tmp[:28])
	res.WriteString(file)
	f.tmp[0] = ':'
	n := Itoa(&f.tmp, 1, context.Line)
	f.tmp[n+1] = ']'
	f.tmp[n+2] = ' '
	res.Write(f.tmp[:n+3])
	message := ""
	if context.Format != nil {
		message = fmt.Sprintf(*context.Format, context.Args...)
	} else {
		message = fmt.Sprint(context.Args...)
	}

	res.WriteString(message)

	l := len(message)
	if l > 0 && message[l-1] != '\n' {
		res.WriteRune('\n')
	}

	return res.Bytes()
}
