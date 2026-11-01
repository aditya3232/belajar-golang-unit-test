package config

import (
	"github.com/spf13/viper"
)

type App struct {
	AppPort      string `json:"app_port"`
	AppEnv       string `json:"app_env"`
	AppName      string `json:"app_name"`
	WebPrefork   bool   `json:"web_prefork"`
	LogLevel     string `json:"log_level"`
	JwtSecretKey string `json:"jwt_secret_key"`
	JwtIssuer    string `json:"jwt_issuer"`
}

type PsqlDB struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	DBName    string `json:"db_name"`
	DBMaxOpen int    `json:"db_max_open"`
	DBMaxIdle int    `json:"db_max_idle"`
}

type Redis struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
}

type Config struct {
	App   App    `json:"app"`
	Psql  PsqlDB `json:"psql"`
	Redis Redis  `json:"redis"`
}

func NewConfig() *Config {
	maxOpen := viper.GetInt("DATABASE_MAX_OPEN_CONNECTION")
	if maxOpen == 0 {
		maxOpen = 25
	}

	maxIdle := viper.GetInt("DATABASE_MAX_IDLE_CONNECTION")
	if maxIdle == 0 {
		maxIdle = 10
	}

	return &Config{
		App: App{
			AppPort:      viper.GetString("APP_PORT"),
			AppEnv:       viper.GetString("APP_ENV"),
			AppName:      viper.GetString("APP_NAME"),
			WebPrefork:   viper.GetBool("WEB_PREFORK"),
			LogLevel:     viper.GetString("LOG_LEVEL"),
			JwtSecretKey: viper.GetString("JWT_SECRET_KEY"),
			JwtIssuer:    viper.GetString("JWT_ISSUER"),
		},
		Psql: PsqlDB{
			Host:      viper.GetString("DATABASE_HOST"),
			Port:      viper.GetString("DATABASE_PORT"),
			User:      viper.GetString("DATABASE_USER"),
			Password:  viper.GetString("DATABASE_PASSWORD"),
			DBName:    viper.GetString("DATABASE_NAME"),
			DBMaxOpen: maxOpen,
			DBMaxIdle: maxIdle,
		},
		Redis: Redis{
			Host:     viper.GetString("REDIS_HOST"),
			Port:      viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
		},
	}
}
