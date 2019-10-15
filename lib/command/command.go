package command

import (
	"os/exec"
	"strings"
)

func Exec(shell string) (msg string, err error) {
	cmd := exec.Command("/bin/bash", "-c", shell)
	b, err := cmd.Output()
	if err != nil {
		return msg, err
	}

	msg = string(b)
	msg = strings.Replace(msg, " ", "", -1)
	msg = strings.Replace(msg, "\n", "", -1)
	msg = strings.Replace(msg, "\t", "", -1)
	msg = strings.Replace(msg, "kB", "", -1)
	return msg, nil
}
