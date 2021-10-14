package clients

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type cmd struct {
	args string
	read func(string) (string, error)
}

type result struct {
	out string
	err error
}

func (c cmd) run(flags []string) (string, error) {
	args := strings.Split(c.args, " ")
	if len(flags) > 0 {
		args = append(flags, args...)
	}
	command := exec.Command("greymatter", args...)
	var out bytes.Buffer
	command.Stdout = &out
	if err := command.Run(); err != nil {
		return "", err
	}
	if c.read == nil {
		return out.String(), nil
	}
	return c.read(out.String())
}

func cliVersion() (string, error) {
	return cmd{
		args: "--version",
		read: func(out string) (string, error) {
			split := strings.Split(out, " ")
			if len(split) < 3 {
				return "", fmt.Errorf("unexpected output: %s", out)
			}
			return split[2], nil
		},
	}.run(nil)
}
