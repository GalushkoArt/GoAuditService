package utils

import (
	"github.com/rs/zerolog/log"
	"strings"
)

func PanicOnError(err error, msgs ...string) {
	msg := strings.Join(msgs, " ")
	if err != nil {
		log.Panic().Err(err).Msg(msg)
	}
}
