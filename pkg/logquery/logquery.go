package logquery

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

type LogLevel int

const (
	Undefined LogLevel = iota
	Debug
	Info
	Warn
	Error
	Fatal

	// 02/28/2020 5:20:57.45
	logFormat = "01/02/2006 3:4:5.00"
)

var (
	logLineRegex = regexp.MustCompile("(\\[.*\\])(\\[.*\\]) (.*)")
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
func processFiles(logMapping map[string]string) map[string][]*Log {
	return map[string][]*Log{}
}

// processFile process the logs for an individual file and return an array of logs
func processFile(filePath string) ([]*Log, error) {
	// Opens a file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Creates a scanner that will let us itereate over each line
	scanner := bufio.NewScanner(file)
	logs := []*Log{}
	for scanner.Scan() {
		log, err := processLine(scanner.Text())
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// process a single line
func processLine(rawLog string) (*Log, error) {
	matches := logLineRegex.FindStringSubmatch(rawLog)
	if len(matches) != 4 {
		return nil, fmt.Errorf("log does not have proper structure")
	}

	// parse time
	time, err := time.Parse(logFormat, matches[1][1:len(matches[1])-1])
	if err != nil {
		return nil, fmt.Errorf("timestamp was not parseable")
	}

	// parse severity
	severity := Undefined
	switch strings.ToLower(matches[2]) {
	case "[debug]":
		severity = Debug
	case "[info]":
		severity = Info
	case "[warn]":
		severity = Warn
	case "[error]":
		severity = Error
	case "[fatal]":
		severity = Fatal
	}
	if severity == Undefined {
		return nil, fmt.Errorf("severity was not parseable")
	}

	// return single log
	return &Log{
		Time:     time,
		Severity: severity,
		Log:      matches[3],
	}, nil
}

// Query will get a range of logs from multiple files and interpolates them based on severity
func (l *LogQuery) Query(start time.Time, entries int, keys []string, minSeverity LogLevel) []Log {
	return []Log{}
}
