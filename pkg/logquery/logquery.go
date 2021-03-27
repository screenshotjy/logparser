package logquery

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
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
	Key      string

	TimeString     string
	SeverityString string
}

func (l Log) String() string {
	return fmt.Sprintf("%s%s[%s] %s", l.TimeString, l.SeverityString, l.Key, l.Log)
}

// Queryier is the interface that calls the Query. This is nice if we ever want to change
// out the underlying implementation
type Queryier interface {
	Query(start time.Time, end time.Time, entries int, keys []string, minSeverity LogLevel) string
}

// LogQuery implements Queryier and will process the logs on creation
type LogQuery struct {
	processedLogs map[string][]*Log
}

// NewLogQuery return a new LogQuery object
func NewLogQuery(logMapping map[string]string) (*LogQuery, error) {
	return &LogQuery{
		processedLogs: processFiles(logMapping),
	}, nil
}

// processLogs processes the logMapping and returns a map of file name to logs
func processFiles(logMapping map[string]string) map[string][]*Log {
	rv := map[string][]*Log{}
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}

	// Concurrently parse files in different go routines for better efficiency
	for fileKey, path := range logMapping {
		// Wait groups help us initiate a bunch of work and then wait for it to finish before returning to execution
		wg.Add(1)
		go func(fileKey, path string) {
			defer wg.Done()
			logs, err := processFile(path, fileKey)
			if err != nil {
				fmt.Printf("error processing log file %s, %s \n", path, err)
				return
			}

			mutex.Lock()
			defer mutex.Unlock()
			rv[fileKey] = logs
		}(fileKey, path)
	}
	wg.Wait()

	return rv
}

// processFile process the logs for an individual file and return an array of logs
func processFile(filePath string, key string) ([]*Log, error) {
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
		log, err := processLine(scanner.Text(), key)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// process a single line
func processLine(rawLog string, key string) (*Log, error) {
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
		Time:           time,
		Severity:       severity,
		Log:            matches[3],
		Key:            key,
		TimeString:     matches[1],
		SeverityString: matches[2],
	}, nil
}

// Query will get a range of logs from multiple files and interpolates them based on severity
func (l *LogQuery) Query(start time.Time, entries int, logKeys []string, minSeverity LogLevel) string {
	wg := sync.WaitGroup{}
	processedFiles := map[string][]Log{}
	mutex := sync.Mutex{}

	// Filter logs for all files
	for _, logKey := range logKeys {
		if logs, ok := l.processedLogs[logKey]; ok {
			wg.Add(1)
			go func(logKey string, logs []*Log) {
				defer wg.Done()
				rv := []Log{}
				for i, log := range logs {
					// If we processed the max logs here, we don't need to iterate further
					if i == entries {
						break
					}
					// Future optimization, we dont need to start our iteration at the beginning. We can
					// do a search for the first time
					if log.Time.After(start) && log.Severity >= minSeverity {
						rv = append(rv, *log)
					}
				}

				mutex.Lock()
				defer mutex.Unlock()
				processedFiles[logKey] = rv
			}(logKey, logs)
		}
	}
	wg.Wait()

	inOrderLogs := logMerge(processedFiles, entries)
	rv := make([]string, len(inOrderLogs))
	for i, log := range inOrderLogs {
		rv[i] = log.String()
	}
	return strings.Join(rv, "\n")
}

// ByTime fufills the sort.Interface so we can sort an array of logs by time using the sort package
type ByTime []Log

func (b ByTime) Len() int           { return len(b) }
func (b ByTime) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByTime) Less(i, j int) bool { return b[i].Time.Before(b[j].Time) }

// logMerge interpolates multiple file logs in order by time
func logMerge(logsByKey map[string][]Log, limit int) []Log {
	fileOrderByFirstLog := []Log{}

	// Get the first log from each logs array
	for _, logs := range logsByKey {
		if len(logs) == 0 {
			continue
		}
		firstLog := logs[0]
		fileOrderByFirstLog = append(fileOrderByFirstLog, firstLog)
	}

	// Sort the order of the first logs
	sort.Sort(ByTime(fileOrderByFirstLog))

	rv := []Log{}
	for len(fileOrderByFirstLog) > 0 {
		// Get the known earliest log
		firstLog := fileOrderByFirstLog[0]

		// Get the next log file's earliest time
		var rangeTime *time.Time
		if len(fileOrderByFirstLog) > 1 {
			rangeTime = &fileOrderByFirstLog[1].Time
		}

		// Get the range of logs from a file up till the next end time
		logsToAdd, endIndex := getRangeLogs(logsByKey[firstLog.Key], rangeTime, limit-len(rv))

		// Append the logs from the file
		rv = append(rv, logsToAdd...)

		// Remove the logs from the slice that were just added
		logsByKey[firstLog.Key] = logsByKey[firstLog.Key][endIndex:]
		if len(rv) == limit {
			return rv
		}

		// Get the next log from the file we just took logs out of and add to the FileOrderByFirstLog
		if len(logsByKey[firstLog.Key]) == 0 {
			fileOrderByFirstLog = fileOrderByFirstLog[1:]
			delete(logsByKey, firstLog.Key)
		} else {
			nextLog := logsByKey[firstLog.Key][0]
			fileOrderByFirstLog[0] = nextLog
			sort.Sort(ByTime(fileOrderByFirstLog))
		}

	}
	return rv
}

func getRangeLogs(logs []Log, endTime *time.Time, limit int) ([]Log, int) {
	if endTime == nil {
		endIndex := len(logs)
		if limit < endIndex {
			endIndex = limit
		}
		return logs[:endIndex], endIndex
	}

	i := 1
	for i < len(logs) {
		log := logs[i]
		if !endTime.After(log.Time) {
			break
		}
		i++
	}
	return logs[:i], i
}
