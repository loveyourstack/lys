package lysexec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Run is a wrapper for exec.CommandContext.Run which includes the results of Stdout and Stderr in every case
func Run(ctx context.Context, writeStdOut bool, name string, args ...string) (stdOut, stdErr string, err error) {

	cmd := exec.CommandContext(ctx, name, args...)

	var outB, errB bytes.Buffer
	if writeStdOut {
		cmd.Stdout = io.MultiWriter(os.Stdout, &outB)
	} else {
		cmd.Stdout = &outB
	}
	cmd.Stderr = &errB
	if err = cmd.Run(); err != nil {
		return outB.String(), errB.String(), fmt.Errorf("cmd.Run failed: %s", errB.String())
	}

	return outB.String(), errB.String(), nil
}
