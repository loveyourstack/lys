package lyscli

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/loveyourstack/lys/internal/cmd"
	"github.com/loveyourstack/lys/internal/lyscmd"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/spf13/cobra"
)

var version = "0.0.1"
var rootCmd = &cobra.Command{
	Use:     "lyscli",
	Version: version,
	Short:   "lyscli - CLI tool for Lys",
	Long:    `lyscli is a CLI tool for running Lys admin tasks`,
	// no Run function: a subcommand is always needed
}

type cliApplication struct {
	*cmd.Application
}

var (
	cliApp *cliApplication
)

func init() {
	cobra.OnInitialize(initApp)
}

func initApp() {

	// load config from file
	conf := lyscmd.Config{}
	err := conf.LoadFromFile("/usr/local/etc/lys_config.toml")
	if err != nil {
		log.Fatalf("initialization: lys_config.toml not found: %s" + err.Error())
	}

	ctx := context.Background()

	// create non-specific app
	app := &cmd.Application{
		Config:   &conf,
		InfoLog:  slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})),
		ErrorLog: slog.New(slog.NewTextHandler(os.Stderr, nil)),
		Validate: validator.New(validator.WithRequiredStructEnabled()),
	}

	// create cli app
	cliApp = &cliApplication{app}

	// connect to db and assign conn to cliApp
	cliApp.Db, err = lyspgdb.GetPool(ctx, conf.Db, conf.DbOwnerUser)
	if err != nil {
		log.Fatalf("initialization: failed to create db connection pool: %s", err.Error())
	}
	// not deferring cliApp.Db.Close() here: it is called before subcommand is reached. Defer close in subcommand instead
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf(err.Error())
	}
}
