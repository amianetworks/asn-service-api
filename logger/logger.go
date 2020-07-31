package logger

import (
	"fmt"
	"log"
)

type Logger struct {
	Tag string
}

func (l *Logger) Print(output string) {
	output = fmt.Sprintf("[%s]: ", l.Tag) + output
	log.Println(output)
}
