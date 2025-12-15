package bootstrap

import (
	"database/sql"
	"errors"
	"os"

	"github.com/api-monolith-template/internal/config"
	"github.com/api-monolith-template/internal/util"
	"github.com/pressly/goose/v3"
)

func StartMigrate(actionType string, name string, version *int64) {
	migrationDir := "migration/db"

	var err error

	db, err := sql.Open("postgres", config.Env.Database.DSN)
	util.ContinueOrFatal(err)
	err = goose.SetDialect("postgres")
	util.ContinueOrFatal(err)

	switch actionType {
	case "create":
		err = goose.Create(db, migrationDir, name, "sql")
	case "up":
		err = goose.Up(db, migrationDir, goose.WithAllowMissing())
	case "up-by-one":
		err = goose.UpByOne(db, migrationDir, goose.WithAllowMissing())
	case "up-to":
		err = goose.UpTo(db, migrationDir, *version, goose.WithAllowMissing())
	case "down":
		err = goose.Down(db, migrationDir, goose.WithAllowMissing())
	case "down-to":
		err = goose.DownTo(db, migrationDir, *version, goose.WithAllowMissing())
	case "status":
		err = goose.Status(db, migrationDir)
	case "reset":
		err = goose.Reset(db, migrationDir, goose.WithAllowMissing())
		if err != nil {
			break
		}
		err = goose.Up(db, migrationDir, goose.WithAllowMissing())
	case "db-reset":
		content, ioErr := os.ReadFile("scripts/reset-db.sql")
		if ioErr != nil {
			err = ioErr
			break
		}
		_, err = db.Exec(string(content))
	default:
		err = errors.New("invalid command")
	}

	util.ContinueOrFatal(err)
}
