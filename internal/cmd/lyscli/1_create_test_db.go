package main

import (
	"context"
	"os"

	"github.com/loveyourstack/lys/internal/sql/ddl"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/spf13/cobra"
)

// expects db users to have been created first (create_users.sql)

var createTestDbCmd = &cobra.Command{
	Use:   "createTestDb",
	Short: "Creates test database from embedded SQL files. Drops existing if present.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		defer cliApp.Db.Close()

		// (re-)create test db
		ctx := context.Background()
		if err := lyspgdb.CreateLocalDb(ctx, ddl.SQLAssets, cliApp.Config.Db, cliApp.Config.DbSuperUser, cliApp.Config.DbOwnerUser, true, false,
			nil, cliApp.InfoLog); err != nil {
			cliApp.ErrorLog.Error("lyspgdb.CreateLocalDb failed: " + err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(createTestDbCmd)
}
