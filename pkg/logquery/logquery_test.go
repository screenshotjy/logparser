package logquery

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Would make these tests table tests in a real setting.
// These tests are helpful to check to see if the logic is correct in my debugger

func TestProcessLine(t *testing.T) {
	assert := assert.New(t)
	testLog := "[02/28/2020 5:20:57.35][error] Could not create database my_db7. Database server rejected request."
	_, err := processLine(testLog, "hi")
	assert.NoError(err)

}

func TestProcessFile(t *testing.T) {
	assert := assert.New(t)
	testFilePath := "../../logs/server1.log"
	_, err := processFile(testFilePath, "hi")
	assert.NoError(err)
}

func TestProcessFiles(t *testing.T) {
	testFileMappings := map[string]string{
		"server1": "../../logs/server1.log",
		"db":      "../../logs/db_server.log",
	}
	_ = processFiles(testFileMappings)
}

func TestQuery(t *testing.T) {
	testFileMappings := map[string]string{
		"server1": "../../logs/server1.log",
		"db":      "../../logs/db_server.log",
	}

	testQuery, _ := NewLogQuery(testFileMappings)
	logs := testQuery.Query(time.Time{}, 100, []string{"server1", "db"}, Debug)
	fmt.Printf(logs)
}
