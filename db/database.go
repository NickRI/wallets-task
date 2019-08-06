package db

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

func Init() (*sql.DB, error) {
	db, err := sql.Open(viper.GetString("driver"), viper.GetString("dsn"))
	if err != nil {
		return db, xerrors.Errorf("Error due database initialization: %w", err)
	}

	db.SetMaxOpenConns(viper.GetInt("max-open-connections"))
	db.SetMaxIdleConns(viper.GetInt("max-idle-connections"))
	db.SetConnMaxLifetime(viper.GetDuration("conn-max-lifetime"))

	return db, db.Ping()
}
