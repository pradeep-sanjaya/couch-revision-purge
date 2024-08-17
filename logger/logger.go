// Package logger provides a custom logger with a configurable timestamp format.
// It wraps the standard log.Logger and adds additional features like custom log formatting,
// including timestamps, file names, and line numbers in each log entry.
package logger

import (
    "fmt"
    "log"
    "os"
    "runtime"
    "time"
)

// Logger wraps the standard log.Logger and provides custom log formatting.
// The custom formatting includes a timestamp, the file name, and the line number
// from where the log entry was generated.
type Logger struct {
    *log.Logger
}


// NewLogger creates a new Logger instance that writes to the specified file.
// The Logger prefixes log messages with a custom timestamp format (yyyy-mm-dd hh:mm:ss),
// the file name, and the line number from where the log entry was generated.
//
// Parameters:
// - logFile: The path to the log file where logs will be written.
//
// Returns:
// - A pointer to a Logger instance.
// - An error if the log file cannot be opened or created.
//
// Example usage:
//
//     logger, err := logger.NewLogger("app.log")
//     if err != nil {
//         log.Fatalf("Failed to create logger: %v", err)
//     }
//     logger.Println("This is a log message.")
//
func NewLogger(logFile string) (*Logger, error) {
    file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return nil, err
    }

    logger := log.New(file, "", 0) // Disable default flags
    return &Logger{logger}, nil
}

// Write implements the io.Writer interface for Logger and adds a custom log entry format.
// Each log entry is prefixed with:
// - The log level ("INFO:")
// - A timestamp in the format "yyyy-mm-dd hh:mm:ss"
// - The file name and line number from where the log entry was generated
//
// Parameters:
// - p: The log message as a byte slice.
//
// Returns:
// - The number of bytes written, and any error encountered during the write.
//
// Example usage:
//
//     logger, _ := logger.NewLogger("app.log")
//     logger.Write([]byte("This is a log message."))
//
func (l *Logger) Write(p []byte) (n int, err error) {
    timestamp := time.Now().Format("2006-01-02 15:04:05")
    _, file, line, _ := runtime.Caller(2)
    fileLine := fmt.Sprintf("%s:%d", file, line)
    message := fmt.Sprintf("INFO: %s %s: %s", timestamp, fileLine, string(p))
    err = l.Logger.Output(2, message)
    return len(p), err
}