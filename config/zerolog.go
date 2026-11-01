package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func NewLogger(appEnv string, logLevel string, serviceName string) {
	level := zerolog.InfoLevel

	parsedLevel, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err == nil {
		level = parsedLevel
	}

	var logger zerolog.Logger

	if appEnv == "development" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
			NoColor:    false,
		}

		consoleWriter.FormatLevel = func(i any) string {
			lvl := strings.ToUpper(fmt.Sprintf("%s", i))
			switch lvl {
			case "DEBUG":
				return "\033[36m| DEBUG |\033[0m"
			case "INFO":
				return "\033[32m| INFO  |\033[0m"
			case "WARN":
				return "\033[33m| WARN  |\033[0m"
			case "ERROR":
				return "\033[31m| ERROR |\033[0m"
			case "FATAL":
				return "\033[35m| FATAL |\033[0m"
			default:
				return fmt.Sprintf("| %-5s|", lvl)
			}
		}

		logger = zerolog.New(consoleWriter).
			Level(level).
			With().
			Timestamp().
			Str("service", serviceName).
			Logger()
	} else {
		zerolog.TimeFieldFormat = time.RFC3339
		logger = zerolog.New(os.Stdout).
			Level(level).
			With().
			Timestamp().
			Str("service", serviceName).
			Logger()
	}

	log.Logger = logger
}
