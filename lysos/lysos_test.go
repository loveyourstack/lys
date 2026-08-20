package lysos

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/loveyourstack/lys/lyserr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDir(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		err := ValidateDir("", "Destination")
		require.Error(t, err)

		var userErr lyserr.User
		require.True(t, errors.As(err, &userErr))
		assert.Equal(t, "Destination: dirPath param is empty", userErr.Message)
	})

	t.Run("missing path", func(t *testing.T) {
		missingPath := filepath.Join(t.TempDir(), "missing-dir")

		err := ValidateDir(missingPath, "Destination")
		require.Error(t, err)

		var userErr lyserr.User
		require.True(t, errors.As(err, &userErr))
		assert.Equal(t, "Destination: path does not exist", userErr.Message)
	})

	t.Run("path is not a directory", func(t *testing.T) {
		file, err := os.CreateTemp(t.TempDir(), "not-a-dir-*.txt")
		require.NoError(t, err)
		require.NoError(t, file.Close())

		err = ValidateDir(file.Name(), "Destination")
		require.Error(t, err)

		var userErr lyserr.User
		require.True(t, errors.As(err, &userErr))
		assert.Equal(t, "Destination: path is not a directory", userErr.Message)
	})

	t.Run("valid directory", func(t *testing.T) {
		dirPath := t.TempDir()

		err := ValidateDir(dirPath, "Destination")
		require.NoError(t, err)
	})
}
