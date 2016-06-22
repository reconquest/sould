package main

import (
	"runtime"

	"github.com/kovetskiy/lorg"
)

// @TODO: create a log files
const (
	logFormatting = `${level:%s\::right:false} ${time} ${prefix}%s`
)

// NewLogger creates new instance of pluggable logger without any prefixes.
func NewLogger() lorg.Logger {
	return NewPrefixedLogger("")
}

// NewPrefixedLogger creates new instance of pluggable logger using specified
// prefix.
func NewPrefixedLogger(prefix string) lorg.Logger {
	if prefix != "" {
		prefix += " "
	}

	format := lorg.NewFormat(logFormatting)
	format.SetPlaceholder("prefix", func(_ lorg.Level, _ string) string {
		return prefix
	})

	logger := lorg.NewLog()
	logger.SetFormat(format)

	return logger
}

func stack() []byte {
	buffer := make([]byte, 1024)
	for {
		stack := runtime.Stack(buffer, true)
		if stack < len(buffer) {
			return buffer[:stack]
		}
		buffer = make([]byte, 2*len(buffer))
	}
}
