package logs

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	goLog "log"
	"os"
)

func Init(level string, logsPath string) {
	logger := log.Logger
	if len(logsPath) != 0 {
		file, err := os.OpenFile(logsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			goLog.Panic(err)
		}
		logger = zerolog.New(file)
	}
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		goLog.Panic(err)
	}
	logger = logger.Level(logLevel).With().Timestamp().Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	log.Logger = logger
}
