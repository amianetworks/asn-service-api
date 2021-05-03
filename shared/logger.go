package shared

import (
	"log"
)

type ASNLogger struct {
	Error   *log.Logger
	Warning *log.Logger
	Info    *log.Logger
	Debug   *log.Logger

}
