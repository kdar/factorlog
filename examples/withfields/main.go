package main

import (
	"fmt"
	"os"

	log "github.com/kdar/factorlog"
	"github.com/mgutz/ansi"
)

func simple() {
	l := log.New(os.Stdout, log.NewStdFormatter("%{Date} %{Time} %{File}:%{Line} %{Message}%{Fields}"))
	l.WithField("total", 100).Debug("with fields")
}

func colors() {
	sevFrmt := "%%{Color `red` `ERROR`}%%{Color `yellow` `WARN`}%%{Color `green` `INFO`}%%{Color `cyan` `DEBUG`}%%{Color `blue` `TRACE`}%%{Color `magenta` `CRITICAL`}%s%%{Color `reset`}"
	frmt := fmt.Sprintf(`%s %%{Message}%%{Fields " %s%%k%s=%%v"}`, fmt.Sprintf(sevFrmt, "%{SEV}"), ansi.ColorCode("blue"), ansi.ColorCode("reset"))
	l := log.New(os.Stdout, log.NewStdFormatter(frmt))
	logger := l.WithFields(log.Fields{
		"var":   "string",
		"count": 5.6,
	})

	logger.Debug("some message")
	logger.Warn("some message")
	logger.Error("some message")
	logger.Trace("some message")
	logger.Critical("some message")
}

func main() {
	simple()
	colors()
}
