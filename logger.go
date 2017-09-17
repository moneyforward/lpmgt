package lpmgt

import (
	"github.com/motemen/go-colorine"
	"os"
)

var logger = &colorine.Logger{
	Prefixes: colorine.Prefixes{
		"warning": colorine.Warn,
		"error": colorine.Error,
		"":        colorine.Info,
		"created": colorine.Info,
		"deleted": colorine.Info,
	},
}

// Log outputs `message` with `prefix` by go-colorine
func Log(prefix, message string) {
	logger.Log(prefix, message)
}

// DieIf outputs log and exit(1) if `err` occurs.
func DieIf(err error) {
	if err != nil {
		Log("error", err.Error())
		os.Exit(1)
	}
}
