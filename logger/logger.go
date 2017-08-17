package logger

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
		"updated": colorine.Info,
		"thrown":  colorine.Info,
		"retired": colorine.Info,
	},
}

// Log outputs `message` with `prefix` by go-colorine
func Log(prefix, message string) {
	logger.Log(prefix, message)
}

// ErrorIf outputs log if `err` occurs.
func ErrorIf(err error) bool {
	if err != nil {
		Log("error", err.Error())
		return true
	}

	return false
}

// DieIf outputs log and exit(1) if `err` occurs.
func DieIf(err error) {
	if err != nil {
		Log("error", err.Error())
		os.Exit(1)
	}
}

// PanicIf raise panic if `err` occurs.
func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}
