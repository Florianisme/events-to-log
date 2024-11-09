package logging

import (
	"encoding/json"
	"fmt"
)

func Log(event *LoggableEvent) {
	marshalledEvent, err := json.Marshal(event)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(marshalledEvent))
}
