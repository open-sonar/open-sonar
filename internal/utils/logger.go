package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type LogEntry struct {
	Level     string `json:"level"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

// logs message at info level
func Info(msg string) {
	entry := LogEntry{
		Level:     "info",
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   msg,
	}
	printLog(entry)
}

// logs message at error level
func Error(msg string) {
	entry := LogEntry{
		Level:     "error",
		Timestamp: time.Now().Format(time.RFC3339),
		Message:   msg,
	}
	printLog(entry)
}

// encodes LogEntry to JSON
func printLog(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to compile log entry:", err)
		return
	}
	fmt.Println(string(data))
}
