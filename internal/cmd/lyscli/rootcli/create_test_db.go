package rootcli

import (
	"fmt"

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
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			// (re-)create test db
			if err = lyspgdb.CreateLocalDb(cmd.Context(), ddl.SQLAssets, cliApp.Config.Db, cliApp.Config.DbSuperUser, cliApp.Config.DbOwnerUser, true, false,
				nil, cliApp.Logger); err != nil {
				return fmt.Errorf("lyspgdb.CreateLocalDb failed: %w", err)
			}

			return nil
		},
	}
}
