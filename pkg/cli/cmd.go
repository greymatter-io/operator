package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type cmd struct {
	args  string
	stdin json.RawMessage
	// Attempts to parse cmdout into loggable key-value pairs with maybe an error.
	// If not specified, running cmd.run returns a result with kvs == ["output", cmdout].
	// If the result has an error (from cmd or reader), result.kvs will always be nil.
	reader func(string) result
	// If > 0, runs cmd on interval until it succeeds
	persist time.Duration
	// If the cmd succeeds, run this cmd with the previous cmdout piped into stdin.
	// If specified, this skips the original cmd's reader.
	then cmdOpt
	// If the cmd (or read) fails, run this cmd as a backup.
	backup cmdOpt
}

type result struct {
	kvs []interface{}
	err error
}

type cmdOpt struct {
	*cmd
	runIf func(string) bool
}

func (c cmd) run(flags []string) result {
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
		// If there was no error
		if c.then.cmd != nil {
			// If c.then is specified, pipe cmdout into it and run it next.
			c.then.stdin = out
			logger.Info("chain", "cmd", c.args, "pipe", c.then.args)
			return c.then.run(flags)
		} else if c.reader != nil {
			// Otherwise, if reader is specified, attempt to read cmdout.
			r = c.reader(cmdout)
		} else {
			r.kvs = []interface{}{"output", cmdout}
		}
	}

	if r.err != nil {
		// Run c.backup if specified and if it passes backupIf (if specified).
		if c.backup.cmd != nil && (c.backup.runIf == nil || c.backup.runIf(cmdout)) {
			logger.Info("chain", "cmd", c.args, "backup", c.backup.args)
			return c.backup.run(flags)
		}
		// Otherwise, if c.persist > 0, run again after waiting.
		if c.persist > 0 {
			logger.Info("chain", "cmd", c.args, "retry-after", c.persist.String())
			time.Sleep(c.persist)
			return c.run(flags)
		}
	}

	// Log the final result.
	if r.err != nil {
		logger.Error(r.err, c.args)
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
			return result{kvs: []interface{}{"output", fields[1]}}
		},
	}).run(nil)
	if len(r.kvs) != 2 {
		return "", r.err
	}
	return r.kvs[1].(string), nil
}
