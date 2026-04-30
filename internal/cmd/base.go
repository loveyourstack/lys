package cmd

import (
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/internal/myapp"
)

// Application contains the fields common to all commands
type Application struct {
	Config   *myapp.Config
	InfoLog  *slog.Logger
	ErrorLog *slog.Logger
	Db       *pgxpool.Pool
	Validate *validator.Validate
}

// NewApplication returns an Application with default settings. Not all fields get initialized.
func NewApplication(conf *myapp.Config) (app *Application) {

	// declare and configure logs
	var infoLog, errorLog *slog.Logger
	infoLog = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	errorLog = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return &Application{
		Config:   conf,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
		Validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}
