package lyscmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/loveyourstack/lys/lyspgdb"
)

// Config contains all configuration settings
type Config struct {
	Db          lyspgdb.Database `toml:"database"`
	DbSuperUser lyspgdb.User
	DbOwnerUser lyspgdb.User
}

func (c *Config) LoadFromFile(configFilePath string) (err error) {

	// ensure supplied path exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return fmt.Errorf("configFilePath does not exist: %s", configFilePath)
	} else if err != nil {
		return fmt.Errorf("os.Stat failed: %w", err)
	}

	// read conf from toml file
	if _, err := toml.DecodeFile(configFilePath, c); err != nil {
		return fmt.Errorf("toml.DecodeFile failed: %w", err)
	}

	return nil
}

func MustGetConfig(t testing.TB) Config {

	_, b, _, _ := runtime.Caller(0)
	projectRootDir := filepath.Join(filepath.Dir(b), "../..")
	configFilePath := projectRootDir + "/lys_config.toml"

	conf := Config{}
	err := conf.LoadFromFile(configFilePath)
	if err != nil {
		t.Fatalf("%s not found: %v", configFilePath, err)
	}

	return conf
}
