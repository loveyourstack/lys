package lysexec

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Run is a wrapper for exec.Cmd.Run() which includes the results of Stdout and Stderr
func Run(name string, args ...string) (out string, err error) {

	cmd := exec.Command(name, args...)

	var outB, errB bytes.Buffer
	cmd.Stdout = &outB
	cmd.Stderr = &errB
	if err = cmd.Run(); err != nil {
		if errB.String() != "" {
			return "", fmt.Errorf("cmd.Run failed: stdout: %s, stderr: %s", outB.String(), errB.String())
		}
		return "", fmt.Errorf("cmd.Run failed (no stderr): %w", err)
	}

	return outB.String(), nil
}
