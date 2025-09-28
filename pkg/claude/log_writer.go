package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"
	"umami/pkg/db"
)

type LogWriter struct {
	dbClient db.DB
	taskID   string
	buffer   strings.Builder
}

func NewLogWriter(dbClient db.DB, taskID string) *LogWriter {
	return &LogWriter{
		dbClient: dbClient,
		taskID:   taskID,
	}
}

func (l *LogWriter) Write(p []byte) (n int, err error) {
	// Add incoming bytes to buffer
	l.buffer.Write(p)

	// Process complete lines from buffer
	l.processCompleteLines()

	return len(p), nil
}

func (l *LogWriter) processCompleteLines() {
	bufferContent := l.buffer.String()
	scanner := bufio.NewScanner(strings.NewReader(bufferContent))

	var processedLines []string
	var lastIncomplete string

	// Process each line
	for scanner.Scan() {
		line := scanner.Text()
		if l.processJSONLine(line) {
			processedLines = append(processedLines, line)
		} else {
			// This line might be incomplete JSON, keep it for next time
			lastIncomplete = line
		}
	}

	// Reset buffer with any incomplete line
	l.buffer.Reset()
	if lastIncomplete != "" {
		l.buffer.WriteString(lastIncomplete)
	}
}

func (l *LogWriter) processJSONLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return true // Empty lines are "processed"
	}

	log.Printf("LogWriter: Processing line: %s", line)

	ctx := context.Background()
	u := Update{}
	err := json.Unmarshal([]byte(line), &u)
	if err != nil {
		log.Printf("LogWriter: Unable to unmarshal update (might be partial): %s", err)
		return false // Not processed, might be incomplete
	}

	messages := []map[string]string{}

	for _, c := range u.Message.Content {
		switch c.Type {
		case "text":
			messages = append(messages, map[string]string{
				"time":  time.Now().Format(time.RFC3339),
				"title": "update",
				"text":  c.Text,
			})
		case "tool_use":
			messages = append(messages, map[string]string{
				"title": "tool",
				"text":  c.Name,
			})
		}
	}

	if len(messages) > 0 {
		err = l.dbClient.InsertLog(ctx, l.taskID, messages)
		if err != nil {
			log.Printf("LogWriter: Unable to insert log %s", err)
			return false
		}
	}

	return true // Successfully processed
}

// Flush processes any remaining buffered content
// Call this when you're done writing to ensure no data is lost
func (l *LogWriter) Flush() error {
	remaining := strings.TrimSpace(l.buffer.String())
	if remaining != "" {
		log.Printf("LogWriter: Flushing remaining buffer: %s", remaining)
		l.processJSONLine(remaining)
	}
	l.buffer.Reset()
	return nil
}
