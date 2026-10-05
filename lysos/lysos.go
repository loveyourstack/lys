package lysos

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/loveyourstack/lys/lyserr"
)

// ValidateDir checks if the provided directory path exists and is a directory.
// errPrefix is used in error messages to indicate the context of the validation (e.g. "Destination").
func ValidateDir(dirPath, errPrefix string) error {

	if strings.TrimSpace(dirPath) == "" {
		return lyserr.User{Message: fmt.Sprintf("%s: dirPath param is empty", errPrefix)}
	}
	if errPrefix == "" {
		errPrefix = "ValidateDir"
	}

	fInfo, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return lyserr.User{Message: fmt.Sprintf("%s: path does not exist: %s", errPrefix, dirPath)}
		}
		return fmt.Errorf("%s: os.Stat failed: %w", errPrefix, err)
	}
	if !fInfo.IsDir() {
		return lyserr.User{Message: fmt.Sprintf("%s: path is not a directory: %s", errPrefix, dirPath)}
	}

	return nil
}

// WriteToClipboard writes s to the clipboard. Only tested on WSL2 so far.
func WriteToClipboard(s string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("pbcopy")
	case "linux":
		// Check if running in WSL
		if isWSL() {
			cmd = exec.Command("clip.exe")
		} else {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		}
	case "windows":
		cmd = exec.Command("clip.exe")
	default:
		return fmt.Errorf("WriteToClipboard not supported on %s", runtime.GOOS)
	}

	cmd.Stdin = strings.NewReader(s)
	return cmd.Run()
}

func isWSL() bool {
	content, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(content)), "microsoft") ||
		strings.Contains(strings.ToLower(string(content)), "wsl")
}
