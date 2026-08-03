package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markburgess/SSTorytime/internal/db/sqlc"
)

var (
	once   sync.Once
	pool   *pgxpool.Pool
	queries *sqlc.Queries
	initErr error
)

// Open returns a shared pool and sqlc querier. Migrations run on first load.
func Open(ctx context.Context) (*pgxpool.Pool, *sqlc.Queries, error) {
	once.Do(func() {
		dsn, err := ResolveDSN("", "")
		if err != nil {
			initErr = err
			return
		}
		pool, queries, initErr = openWithDSN(ctx, dsn)
	})
	return pool, queries, initErr
}

// OpenDSN is like Open but allows an explicit DSN (e.g. from --database-url).
// Only the first successful Open/OpenDSN wins (sync.Once).
func OpenDSN(ctx context.Context, dsn string) (*pgxpool.Pool, *sqlc.Queries, error) {
	once.Do(func() {
		if dsn == "" {
			var err error
			dsn, err = ResolveDSN("", "")
			if err != nil {
				initErr = err
				return
			}
		}
		pool, queries, initErr = openWithDSN(ctx, dsn)
	})
	return pool, queries, initErr
}

func openWithDSN(ctx context.Context, dsn string) (*pgxpool.Pool, *sqlc.Queries, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, nil, fmt.Errorf("migrate: %w", err)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("parse dsn: %w", err)
	}
	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("pool: %w", err)
	}
	if err := p.Ping(ctx); err != nil {
		p.Close()
		return nil, nil, fmt.Errorf("ping: %w", err)
	}
	return p, sqlc.New(p), nil
}

func runMigrations(dsn string) error {
	src, err := iofs.New(MigrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}
	// migrate pgx driver expects postgres:// or pgx5://
	migDSN := dsn
	if strings.HasPrefix(migDSN, "postgresql://") {
		migDSN = "pgx5://" + strings.TrimPrefix(migDSN, "postgresql://")
	} else if strings.HasPrefix(migDSN, "postgres://") {
		migDSN = "pgx5://" + strings.TrimPrefix(migDSN, "postgres://")
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, migDSN)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// ResolveDSN: flag > DATABASE_URL > POSTGRESQL_URI > ~/.SSTorytime > default.
func ResolveDSN(flagDSN, flagFile string) (string, error) {
	if flagDSN != "" {
		return flagDSN, nil
	}
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v, nil
	}
	if v := os.Getenv("POSTGRES_URL"); v != "" {
		return v, nil
	}
	if v := os.Getenv("POSTGRESQL_URI"); v != "" {
		return v, nil
	}
	if dsn, ok := dsnFromCredentialsFile(flagFile); ok {
		return dsn, nil
	}
	// Historical defaults (docker-compose / docs)
	return "postgres://sstoryline:sst_1234@127.0.0.1:5432/sstoryline?sslmode=disable", nil
}

func dsnFromCredentialsFile(override string) (string, bool) {
	path := override
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		path = filepath.Join(home, ".SSTorytime")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	user, pass, dbname := "sstoryline", "sst_1234", "sstoryline"
	host := "127.0.0.1"
	port := "5432"
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.ToLower(key))
		val = strings.TrimSpace(val)
		switch key {
		case "user":
			user = val
		case "passwd", "password":
			pass = val
		case "db", "dbname":
			dbname = val
		case "host":
			host = val
		case "port":
			port = val
		}
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbname), true
}

// Queries returns the shared querier after Open.
func Queries() *sqlc.Queries { return queries }

// Pool returns the shared pool after Open.
func Pool() *pgxpool.Pool { return pool }
