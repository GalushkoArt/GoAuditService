package logs

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	goLog "log"
	"os"
)

func Init(level string, logsPath string) {
	file, err := os.OpenFile(logsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		goLog.Panic(err)
	}
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		goLog.Panic(err)
	}
	logger := zerolog.New(file).Level(logLevel).With().Timestamp().Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	log.Logger = logger
}
