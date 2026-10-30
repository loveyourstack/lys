package rootcli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	appCmd "github.com/loveyourstack/lys/internal/cmd"
	"github.com/loveyourstack/lys/internal/cmd/lyscli/cliapp"
	"github.com/loveyourstack/lys/internal/myapp"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/spf13/cobra"
)

var version = "0.0.1"
var rootCmd = &cobra.Command{
	Use:           "lyscli",
	Version:       version,
	Short:         "lyscli - CLI tool for Lys",
	Long:          `lyscli is a CLI tool for running Lys admin tasks`,
	SilenceErrors: true, // subcommand errors are returned upwards via RunE and handled in Execute() below
	SilenceUsage:  true,
	// no Run function: a subcommand is always needed
}

var cliApp *cliapp.App

func addSubCommands() {
	rootCmd.AddCommand(CreateTestDbCmd(cliApp))
}

func Run() error {

	// set up signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ensure that context cancelation propagates to subcommands
	rootCmd.SetContext(ctx)

	conf := myapp.Config{}
	err := conf.LoadFromFile("/usr/local/etc/lys_config.toml")
	if err != nil {
		return fmt.Errorf("conf.LoadFromFile failed: %w", err)
	}

	// create non-specific app
	app := appCmd.NewApplication(&conf)

	// create cli app
	cliApp = &cliapp.App{
		Application: app,
	}

	// connect to db and assign pool to cliApp
	cliApp.Db, err = lyspgdb.GetPool(ctx, conf.Db, conf.DbOwnerUser, conf.General.AppName+" Cli")
	if err != nil {
		return fmt.Errorf("lyspgdb.GetPool failed: %w", err)
	}
	defer cliApp.Db.Close()

	// subcommands
	addSubCommands()

	// execute the root command, which runs the appropriate subcommand based on user input
	return rootCmd.Execute()
}
