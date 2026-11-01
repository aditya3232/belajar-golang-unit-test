package config

import (
	"fmt"
	"net/url"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Postgres struct {
	DB *gorm.DB
}

func (cfg Config) ConnectionPostgres() (*Postgres, error) {
	dbConfig := cfg.Psql
	encodedPassword := url.QueryEscape(dbConfig.Password)

	targetURI := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbConfig.User,
		encodedPassword,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)

	db, err := gorm.Open(postgres.Open(targetURI), &gorm.Config{})
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "config.ConnectionPostgres").
			Str("host", cfg.Psql.Host).
			Msg("failed to connect to database")
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Error().
			Err(err).
			Str("source", "config.ConnectionPostgres").
			Msg("failed to get database instance")
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.Psql.DBMaxOpen)
	sqlDB.SetMaxIdleConns(cfg.Psql.DBMaxIdle)

	return &Postgres{DB: db}, nil
}
