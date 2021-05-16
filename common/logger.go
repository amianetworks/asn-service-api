package commonapi

import (
	"log"
)

type ASNLogger struct {
	Error   *log.Logger // Error log will be printed to the console and written to the log file
	Warning *log.Logger // Warning log will be written into the log file
	Info    *log.Logger // Info log will be written into the log file
	Debug   *log.Logger // Debug log will be printed to the console
}
