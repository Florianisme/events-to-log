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
	return &Logger{
		zerolog.New(os.Stdout),
	}
}

func (s *Logger) Log(event *LoggableEvent) {
	marshalledEvent, err := json.Marshal(event)
	if err != nil {
		s.Logger.Log().Err(err)
		return
	}

	fmt.Println(string(marshalledEvent))
}
