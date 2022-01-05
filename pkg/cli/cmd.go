package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type cmd struct {
	args  string
	stdin json.RawMessage
	// Notifies the caller to requeue the cmd if it fails.
	requeue bool
	// A custom logger; if not set, nothing is logged.
	log func(string, error)
	// If set, modifies the output before it is returned.
	modify func([]byte) ([]byte, error)
	// If set, is run with the stdout of a successful parent cmd piped in.
	then *cmd
}

func (c cmd) run(flags []string) (string, error) {
	args := strings.Split(c.args, " ")
	if len(flags) > 0 {
		args = append(flags, args...)
	}

	command := exec.Command("greymatter", args...)
	if len(c.stdin) > 0 {
		command.Stdin = bytes.NewReader(c.stdin)
	}

	out, err := command.CombinedOutput()
	outStr := string(out)

	// If err is a bad exit code, capture stderr as the error.
	if err != nil {
		err = fmt.Errorf(outStr)
	}

	if err == nil {
		// If cmd.modify is defined, call it on the output.
		// If modify fails, capture the error string for logging.
		if c.modify != nil {
			out, err = c.modify(out)
			if err != nil {
				outStr = err.Error()
			} else {
				outStr = string(out)
			}
		}

		// If cmd.then is defined, run it next.
		if err == nil && c.then != nil {
			c.then.stdin = out
			return c.then.run(flags)
		}
	}

	// If a log function is specified, call it
	if c.log != nil {
		c.log(outStr, err)
	}

	return outStr, err
}

func cliversion() (string, error) {
	output, err := (cmd{args: "--version"}).run(nil)
	if err != nil {
		return "", err
	}
	split := strings.Split(output, " ")
	if len(split) < 4 {
		return "", fmt.Errorf("unexpected output: %s", output)
	}
	return split[2], nil
}
