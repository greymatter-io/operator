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
	// A custom logger; if not used, the full result.cmdout will be logged.
	log func(error)
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

	// If a log function is specified, call it
	if c.log != nil {
		c.log(err)
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
