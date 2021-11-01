package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type cmd struct {
	args    string
	stdin   json.RawMessage
	requeue bool

	// Attempts to parse cmdout into loggable key-value pairs with maybe an error.
	// If not specified, cmd.run returns a result with kvs == ["output", cmdout].
	// If the result has an error (from cmd or reader), kvs will always be nil.
	reader func(string) result

	// If the cmd succeeds, run this cmd with the cmdout piped into stdin.
	// If reader is specified, it is called on cmdout prior to piping into stdin.
	and cmdOpt
	// If the cmd (or read) fails, run this cmd instead.
	or cmdOpt
}

type result struct {
	cmdout string
	kvs    []interface{}
	err    error
}

type cmdOpt struct {
	*cmd
	when func(string) bool
}

// describes the previous cmd in a cmd chain
type src struct {
	args, action string
}

func (c cmd) run(flags []string, from ...src) result {
	args := strings.Split(c.args, " ")
	if len(flags) > 0 {
		for _, flag := range flags {
			args = append(strings.Split(flag, " "), args...)
		}
	}

	command := exec.Command("greymatter", args...)
	if len(c.stdin) > 0 {
		command.Stdin = bytes.NewReader(c.stdin)
	}

	out, err := command.CombinedOutput()
	cmdout := string(out)
	var r result

	// If err is a bad exit code, capture stderr as the error.
	if err != nil {
		r.err = fmt.Errorf(cmdout)
	} else {
		// Otherwise, if c.reader is defined, call it on cmdout to parse it.
		r.cmdout = cmdout
		if c.reader != nil {
			r = c.reader(r.cmdout)
		}
	}

	if r.err != nil {
		// If there is an err (as cmdout or from c.reader), run c.or if specified.
		if c.or.cmd != nil && (c.or.when == nil || c.or.when(r.cmdout)) {
			from = append(from, src{strings.Split(c.args, " ")[0], "||"})
			return c.or.run(flags, from...)
		}
	} else {
		// If c.and is specified, pipe cmdout into it and run it next.
		if c.and.cmd != nil && (c.and.when == nil || c.and.when(r.cmdout)) {
			c.and.stdin = json.RawMessage(r.cmdout)
			from = append(from, src{strings.Split(c.args, " ")[0], ">"})
			return c.and.run(flags, from...)
		} else {
			r.kvs = []interface{}{"output", cmdout}
		}
	}

	if len(from) > 0 {
		for i := len(from) - 1; i >= 0; i-- {
			c.args = fmt.Sprintf("%s %s %s", from[i].args, from[i].action, c.args)
		}
	}

	// Log the final result.
	if r.err != nil {
		logger.Error(r.err, c.args, r.kvs...)
	} else {
		logger.Info(c.args, r.kvs...)
	}

	return r
}

func cliversion() (string, error) {
	r := (cmd{
		args: "version",
		// args: "--version",
		reader: func(out string) result {
			// split := strings.Split(out, " ")
			// if len(split) < 3 {
			// 	return result{err: fmt.Errorf("unexpected output")}
			// }
			// return result{kvs: []interface{}{split[2]}}

			// CLI <4 outputs 6 lines, with the 2nd being the version.
			lines := strings.Split(out, "\n")
			if len(lines) != 6 {
				return result{err: fmt.Errorf("unexpected output")}
			}
			fields := strings.Fields(lines[1])
			if len(fields) != 2 {
				return result{err: fmt.Errorf("unexpected output")}
			}
			return result{cmdout: fields[1]}
		},
	}).run(nil)
	if len(r.kvs) != 2 {
		return "", r.err
	}
	return r.cmdout, nil
}
