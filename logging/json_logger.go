package logging

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"os"
)

type Logger struct {
	Logger zerolog.Logger
}

func Init() *Logger {
	logLevel, err := zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	return &Logger{
		zerolog.New(os.Stdout).Level(logLevel),
	}
}

func (s *Logger) Log(event *LoggableEvent) {
	marshalledEvent, err := json.Marshal(event)
	if err != nil {
		s.Logger.Log().Err(err).Msg("could not marshal event")
		return
	}

	fmt.Println(string(marshalledEvent))
}
