package logger

import (
	"fmt"
	"sync"
	"time"
)

// Logger struct that holds logs in a slice and ensures thread safety.
type Logger struct {
	mu   sync.Mutex
	logs []string
}

// Global logger instance
var instance *Logger
var once sync.Once

// GetLogger returns the singleton instance of the logger.
func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			logs: make([]string, 0),
		}
	})
	return instance
}

// Log adds a new log entry with a timestamp.
func (l *Logger) Log(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("%s "+format, append([]interface{}{timestamp}, args...)...)
	l.logs = append(l.logs, logEntry)
	fmt.Println(logEntry)
}

// GetLogs returns all stored logs.
func (l *Logger) GetLogs() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string{}, l.logs...) // Return a copy to avoid modification
}
