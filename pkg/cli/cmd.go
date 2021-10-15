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
	// If not specified, running cmd.run returns result.kvs with ["output", cmdout].
	// If the result has an error (from cmd or reader), result.kvs will always be nil.
	reader func(string) result
	// If the cmd (or read) fails, run this cmd as a backup. This makes cmd chainable.
	backup *cmd
}

type result struct {
	kvs []interface{}
	err error
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
	// If err is a bad exit code, out will be from stderr, which should be logged.
	if err != nil {
		r.err = fmt.Errorf(cmdout)
	} else {
		// If there was no error, and a reader is specified, attempt to read cmdout.
		if c.reader != nil {
			r = c.reader(cmdout)
		} else {
			r.kvs = []interface{}{"output", cmdout}
		}
	}
	// If there is an error (either from command or c.reader), and a backup is set, run it.
	if r.err != nil && c.backup != nil {
		return c.backup.run(flags)
	}
	// Log the final result (the original cmd or the last cmd in a cmd.backup chain).
	if r.err != nil {
		logger.Error(r.err, c.args)
	} else {
		logger.Info(c.args, r.kvs...)
	}
	return r
}

// maybe temp while CLI is being worked on
func (c cmd) persist(seconds int, flags []string) {
	r := c.run(flags)
	if r.err != nil {
		time.Sleep(time.Second * time.Duration(seconds))
		c.persist(seconds, flags)
	}
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
			return result{kvs: []interface{}{fields[1]}}
		},
	}).run(nil)
	if len(r.kvs) != 1 {
		return "", r.err
	}
	return r.kvs[0].(string), nil
}
