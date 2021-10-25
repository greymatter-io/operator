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

func FromStrings(ss ...string) cue.Value {
	return cuecontext.New().CompileString(strings.Join(ss, "\n"))
}

func Strings(kvs map[string]string) cue.Value {
	var s string
	for k, v := range kvs {
		if v != "" {
			s += fmt.Sprintf(`%s: "%s"`, k, v) + "\n"
		}
	}
	return FromStrings(s)
}

func StringSlices(kvs map[string][]string) cue.Value {
	var s string
	for k, v := range kvs {
		if len(v) > 0 {
			var quoted []string
			for _, vv := range v {
				quoted = append(quoted, fmt.Sprintf(`"%s"`, vv))
			}
			s += fmt.Sprintf(`%s: [%s]`, k, strings.Join(quoted, ",")) + "\n"
		}
	}
	return FromStrings(s)
}

func Interfaces(kvs map[string]interface{}) cue.Value {
	var s string
	for k, v := range kvs {
		s += fmt.Sprintf(`%s: %v`, k, v) + "\n"
	}
	return FromStrings(s)
}

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
