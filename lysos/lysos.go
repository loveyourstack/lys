package lysos

import (
	"fmt"
	"os"
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
