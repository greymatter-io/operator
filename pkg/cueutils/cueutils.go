package cueutils

import (
	"bytes"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"github.com/go-logr/logr"
	"github.com/kylelemons/godebug/diff"
)

// Strings creates a cue.Value from fields defined in a map[string]string.
func Strings(kvs map[string]string) cue.Value {
	var ss []string
	for k, v := range kvs {
		if v != "" {
			ss = append(ss, fmt.Sprintf(`%s: "%s"`, k, v))
		}
	}
	return FromStrings(ss...)
}

// StringSlices creates a cue.Value from fields defined in a map[string][]string.
func StringSlices(kvs map[string][]string) cue.Value {
	var ss []string
	for k, v := range kvs {
		if len(v) > 0 {
			var quoted []string
			for _, vv := range v {
				quoted = append(quoted, fmt.Sprintf(`"%s"`, vv))
			}
			ss = append(ss, fmt.Sprintf(`%s: [%s]`, k, strings.Join(quoted, ",")))
		}
	}
	return FromStrings(ss...)
}

// StringSlices creates a cue.Value from fields defined in a map[string]interface{}.
// Supported Cue builtin types are bool, number, and struct.
func Interfaces(kvs map[string]interface{}) cue.Value {
	var ss []string
	for k, v := range kvs {
		ss = append(ss, fmt.Sprintf(`%s: %v`, k, v))
	}
	return FromStrings(ss...)
}

// FromStrings creates a cue.Value from string arguments.
func FromStrings(ss ...string) cue.Value {
	return cuecontext.New().CompileString(strings.Join(ss, "\n"))
}

// LogError logs errors that may or may not contain a list of cue/errors.Error.
// If the error provided is not a cue/errors.Error, a plain error is logged.
func LogError(logger logr.Logger, err error) {
	switch v := err.(type) {
	case errors.Error:
		for _, e := range errors.Errors(v) {
			logger.Error(e, e.Position().String())
		}
	default:
		logger.Error(v, "")
	}
}

// Diff prints a line-by-line diff between two Cue values.
func Diff(a, b cue.Value) {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	aLines := strings.Split(aStr, "\n")
	bLines := strings.Split(bStr, "\n")

	chunks := diff.DiffChunks(aLines, bLines)

	buf := new(bytes.Buffer)
	for _, c := range chunks {
		for _, d := range c.Added {
			fmt.Fprintf(buf, "\033[32m+%s\n", d)
		}
		for _, d := range c.Deleted {
			fmt.Fprintf(buf, "\033[31m-%s\n", d)
		}
		fmt.Fprintf(buf, "\033[1;34m\033[0m")
	}
	fmt.Println(strings.TrimRight(buf.String(), "\n"))
}
