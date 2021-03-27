package logquery

import "time"

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
	Fatal
)

// Single log of a log file
type Log struct {
	Time     time.Time
	Severity LogLevel
	Log      string
}

// Queryier is the interface that calls the Query. This is nice if we ever want to change
// out the underlying implementation
type Queryier interface {
	Query(start time.Time, end time.Time, entries int, keys []string, minSeverity LogLevel) []Log
}

// LogQuery implements Queryier and will process the logs on creation
type LogQuery struct {
	processedLogs map[string][]string
}

// NewLogQuery return a new LogQuery object
func NewLogQuery(logMapping map[string]string) (*LogQuery, error) {
	return &LogQuery{}, nil
}

// processLogs processes the logMapping and returns a map of file name to logs
func processLogs(logMapping map[string]string) map[string][]*Log {
	return map[string][]*Log{}
}

// processFile process the logs for an individual file and return an array of logs
func processFile(fileName string) []*Log {
	return []*Log{}
}

// Query will get a range of logs from multiple files and interpolates them based on severity
func (l *LogQuery) Query(start time.Time, entries int, keys []string, minSeverity LogLevel) []Log {
	return []Log{}
}
