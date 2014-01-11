package main

import (
	"github.com/kdar/factorlog"
	"os"
)

func main() {
	log := factorlog.New(os.Stdout, factorlog.NewGlogFormatter())
	log.Print("Hello there!\n")
}
