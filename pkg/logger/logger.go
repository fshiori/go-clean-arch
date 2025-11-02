package logger

import (
	"log"
	"os"
)

var (
	// Info logger for informational messages
	Info *log.Logger

	// Warning logger for warning messages
	Warning *log.Logger

	// Error logger for error messages
	Error *log.Logger
)

// Init initializes the loggers
func Init() {
	Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	Info.Println("Logger initialized")
}

// LogInfo logs an informational message
func LogInfo(format string, v ...interface{}) {
	Info.Printf(format, v...)
}

// LogWarning logs a warning message
func LogWarning(format string, v ...interface{}) {
	Warning.Printf(format, v...)
}

// LogError logs an error message
func LogError(format string, v ...interface{}) {
	Error.Printf(format, v...)
}
