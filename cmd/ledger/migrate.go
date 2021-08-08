package main

import (
	"context"
	"database/sql"

	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ddouglas/ledger"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

const migrationDir = ".config/mysql/migrations/"

func actionCreateMigration(c *cli.Context) error {

	core := buildCore()

	name := c.String("name")
	files, err := filepath.Glob(fmt.Sprintf("%s*.sql", migrationDir))
	if err != nil {
		core.logger.WithError(err).Fatal("failed to read migrations directory")
	}

	sort.Strings(files)

	for _, file := range files {
		filename := strings.TrimPrefix(file, migrationDir)
		filename = strings.TrimSuffix(filename, ".sql")
		entry := core.logger.WithField("name", filename)
		parts := strings.Split(filename, "_")
		if len(parts) < 2 {
			entry.Fatal("unable to process filename, not enough parts")
		}
		filename = strings.Join(parts[1:], "_")

		if filename == name {
			entry.WithField("migration", file).Fatal("migration with this name already exists")
		}

	}

	now := time.Now()
	filename := fmt.Sprintf("%s%s_%s.sql", migrationDir, now.Format("20060102150405"), name)
	entry := core.logger.WithField("name", name).WithField("filename", filename)
	f, err := os.Create(filename)
	if err != nil {
		entry.WithError(err).Fatal("failed to create migration")
	}
	defer f.Close()

	entry.Info("migration created successfully")
	return nil
}

func runMigrations(core *core) {
	if logger == nil {
		panic("[runMigrations] logger is not configured, please run buildCore first")
	}

	if core == nil {
		logger.Fatal("[runMigrations] core cannot be nil")
	}

	if dbx == nil {
		logger.Fatal("[runMigrations] DB has not been initialized")
	}

	if !cfg.MySQL.Migrate {
		return
	}

	files, err := filepath.Glob(fmt.Sprintf("%s*.sql", migrationDir))
	if err != nil {
		core.logger.WithError(err).Fatal("[runMigrations] failed to read migrations")
	}

	sort.Strings(files)

	var ctx = context.Background()

	err = core.repos.migrations.InitializeMigrationsTable(ctx, cfg.MySQL.DB, "migrations")
	if err != nil {
		core.logger.WithError(err).Fatal("[runMigrations] failed to initialize migrations table")
	}

	core.logger.Info("[runMigrations] migration table initialized, checking migrations")

	for _, file := range files {

		name := strings.TrimPrefix(file, migrationDir)
		entry := core.logger.WithField("migration", name)

		_, err := core.repos.migrations.Migration(ctx, name)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			entry.WithError(err).Fatal("[runMigrations] failed to check if migration has been executed")
		}

		if err == nil {
			continue
		}

		entry.Info("new migration found, executing migration")

		f, err := os.Open(file)
		if err != nil {
			entry.WithError(err).Fatal("[runMigrations] failed to open migration file")
		}

		data, err := io.ReadAll(f)
		if err != nil {
			entry.WithError(err).Fatal("[runMigrations] failed to read migration file")
		}

		if len(data) == 0 {
			entry.WithError(err).Fatal("[runMigrations] empty migration file detected, halting execution")
		}

		query := string(data)

		tx, err := dbx.BeginTxx(ctx, nil)
		if err != nil {
			entry.WithError(err).Fatal("[runMigrations] failed initialize transaction")
		}

		m := &ledger.Migration{Name: name, Executed: false}

		m, err = core.repos.migrations.CreateMigrationTx(ctx, tx, m)
		if err != nil {
			entry.WithError(err).Fatal("[runMigrations] failed to create migration")
		}

		_, err = dbx.ExecContext(ctx, query)
		if err != nil {
			entry.WithError(err).Fatal("[runMigrations] failed to execute migration")
		}

		m.Executed = true

		_, err = core.repos.migrations.UpdateMigrationTx(ctx, tx, m)
		if err != nil {
			entry.WithError(err).Fatal("[runMigrations] failed to update migration")
		}

		err = tx.Commit()
		if err != nil {
			entry.WithError(err).Fatal("[runMigrations] failed to commit transactions")
		}

		entry.Info("migration executed successfully")
		time.Sleep(time.Second)

	}

}
