package network

import (
    "log"
    "sync"
    "testing"
    "fmt"
)

// mockLogger is a mock implementation of a logger used for testing.
// It captures log messages in memory for later inspection.
type mockLogger struct {
    messages []string
    mu       sync.Mutex
}

// Printf formats according to a format specifier and appends the resulting string
// to the mockLogger's messages slice.
func (ml *mockLogger) Printf(format string, v ...interface{}) {
    ml.mu.Lock()
    defer ml.mu.Unlock()
    ml.messages = append(ml.messages, fmt.Sprintf(format, v...))
}

// Println appends the provided arguments as a single string to the mockLogger's
// messages slice, similar to fmt.Println.
func (ml *mockLogger) Println(v ...interface{}) {
    ml.mu.Lock()
    defer ml.mu.Unlock()
    msg := fmt.Sprintln(v...)
    ml.messages = append(ml.messages, msg)
}

// Write implements the io.Writer interface for mockLogger, allowing it to be used
// with log.New. It appends the provided byte slice to the messages slice.
func (ml *mockLogger) Write(p []byte) (n int, err error) {
    ml.mu.Lock()
    defer ml.mu.Unlock()
    ml.messages = append(ml.messages, string(p))
    return len(p), nil
}

// TestScanNetwork verifies that ScanNetwork correctly identifies running CouchDB instances
// in a given CIDR range. It uses a mock logger and a mocked IsCouchDBRunning function.
func TestScanNetwork(t *testing.T) {
    logger := &mockLogger{}
    cidr := "192.168.1.0/30" // Small range for testing

    // Mock implementation of IsCouchDBRunning
    mockIsCouchDBRunning := func(ip, port string) bool {
        return ip == "192.168.1.1"
    }

    count := ScanNetwork(cidr, "5984", log.New(logger, "", 0), mockIsCouchDBRunning)
    expectedCount := 1

    if count != expectedCount {
        t.Errorf("Expected %d CouchDB instances, found %d", expectedCount, count)
    }
}