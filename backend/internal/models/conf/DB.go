package conf

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type DB struct {
	Host      string `env:"DB_HOST" envDefault:"localhost"`
	User      string `env:"DB_USER" envDefault:"postgres"`
	Password  string `env:"DB_PASSWORD" envDefault:""`
	TableName string `env:"DB_NAME" envDefault:"omniav"`
	Port      string `env:"DB_PORT" envDefault:"5432"`
	SSLMode   string `env:"DB_SSLMODE" envDefault:"disable"`
}

func (db *DB) Parse() error {
	if err := godotenv.Load(); err != nil {
		slog.Warn("no .env file found, falling back to environment variables")
	}

	return env.Parse(db)
}

func (db *DB) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		db.Host, db.User, db.Password, db.TableName, db.Port, db.SSLMode,
	)
}
