package lysexec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Run is a wrapper for exec.Cmd.Run() which includes the results of Stdout and Stderr in case of error
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

// RunWithStdout is a wrapper for exec.Cmd.Run() which includes the results of Stdout and Stderr in case of error. It continues to write Stdout as normal.
func RunWithStdout(name string, args ...string) (out string, err error) {

	cmd := exec.Command(name, args...)

	var outB, errB bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &outB)
	cmd.Stderr = &errB
	if err = cmd.Run(); err != nil {
		if errB.String() != "" {
			return "", fmt.Errorf("cmd.Run failed: stdout: %s, stderr: %s", outB.String(), errB.String())
		}
		return "", fmt.Errorf("cmd.Run failed (no stderr): %w", err)
	}

	return outB.String(), nil
}
