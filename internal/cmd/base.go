package cmd

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/internal/lyscmd"
)

// Application contains the fields common to all commands
type Application struct {
	Config   *lyscmd.Config
	InfoLog  *slog.Logger
	ErrorLog *slog.Logger
	Db       *pgxpool.Pool
	Validate *validator.Validate
}
