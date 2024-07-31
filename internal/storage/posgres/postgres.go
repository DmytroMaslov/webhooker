package posgres

import (
	"database/sql"
	"fmt"
	"webhooker/config"

	_ "github.com/lib/pq"
)

const (
	driverName = "postgres"
	sslMode    = "disable"
)

type PgClient struct {
	db *sql.DB
}

func NewPgClient(cfg *config.PgCredentials) (*PgClient, error) {
	connectionStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host, cfg.DbName, sslMode)
	db, err := sql.Open(driverName, connectionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db, %w", err)
	}

	return &PgClient{
		db: db,
	}, nil

}

func (c *PgClient) Close() error {
	return c.db.Close()
}
