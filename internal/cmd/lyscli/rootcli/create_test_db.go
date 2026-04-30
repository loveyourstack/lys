package rootcli

import (
	"os"

	"github.com/loveyourstack/lys/internal/cmd/lyscli/cliapp"
	"github.com/loveyourstack/lys/internal/sql/ddl"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/spf13/cobra"
)

// expects db users to have been created first (create_users.sql)

func CreateTestDbCmd(cliApp *cliapp.App) *cobra.Command {
	return &cobra.Command{
		Use:   "createTestDb",
		Short: "Creates test database from embedded SQL files. Drops existing if present.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {

			defer cliApp.Db.Close()

			// (re-)create test db
			if err := lyspgdb.CreateLocalDb(cmd.Context(), ddl.SQLAssets, cliApp.Config.Db, cliApp.Config.DbSuperUser, cliApp.Config.DbOwnerUser, true, false,
				nil, cliApp.InfoLog); err != nil {
				cliApp.ErrorLog.Error("lyspgdb.CreateLocalDb failed: " + err.Error())
				os.Exit(1)
			}
		},
	}
}
