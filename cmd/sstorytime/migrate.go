package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/markburgess/SSTorytime/internal/db"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database schema migrations",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := newMigrate()
		if err != nil {
			return err
		}
		defer m.Close()
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		v, dirty, err := m.Version()
		if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
			return fmt.Errorf("%w: %w", ErrMigrateVersion, err)
		}
		fmt.Printf("ok version=%d dirty=%v\n", v, dirty)
		return nil
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down [n]",
	Short: "Roll back n migrations (default 1)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n := 1
		if len(args) == 1 {
			var err error
			n, err = strconv.Atoi(args[0])
			if err != nil || n < 1 {
				return ErrPositiveN
			}
		}
		m, err := newMigrate()
		if err != nil {
			return err
		}
		defer m.Close()
		if err := m.Steps(-n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		return nil
	},
}

var migrateVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print current migration version",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := newMigrate()
		if err != nil {
			return err
		}
		defer m.Close()
		v, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("version=none")
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
		return nil
	},
}

func init() {
	registerMultiCall("migrate", "sstorytime-migrate")
	migrateCmd.AddCommand(migrateUpCmd, migrateDownCmd, migrateVersionCmd)
}

func newMigrate() (*migrate.Migrate, error) {
	dsn, err := db.ResolveDSN(databaseURL, "")
	if err != nil {
		return nil, err
	}
	src, err := iofs.New(db.MigrationsFS, "migrations")
	if err != nil {
		return nil, err
	}
	// pgx5 URL scheme for migrate driver
	migDSN := dsn
	if len(migDSN) > 11 && migDSN[:11] == "postgres://" {
		migDSN = "pgx5://" + migDSN[11:]
	} else if len(migDSN) > 13 && migDSN[:13] == "postgresql://" {
		migDSN = "pgx5://" + migDSN[13:]
	}
	return migrate.NewWithSourceInstance("iofs", src, migDSN)
}
