package cmd

import (
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/internal/myapp"
	"github.com/loveyourstack/lys/lyslog"
)

// Application contains the fields common to all commands
type Application struct {
	Config   *myapp.Config
	Logger   *slog.Logger
	Db       *pgxpool.Pool
	Validate *validator.Validate
}

// NewApplication returns an Application with default settings. Not all fields get initialized.
func NewApplication(conf *myapp.Config) (app *Application) {

	return &Application{
		Config:   conf,
		Logger:   slog.New(lyslog.NewSplitStreamHandler(os.Stdout, os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})),
		Validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}
