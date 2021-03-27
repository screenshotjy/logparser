package logquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Would make these tests table tests in a real setting.
// These tests are helpful to check to see if the logic is correct in my debugger

func TestProcessLine(t *testing.T) {
	assert := assert.New(t)
	testLog := "[02/28/2020 5:20:57.35][error] Could not create database my_db7. Database server rejected request."
	_, err := processLine(testLog)
	assert.NoError(err)

}

func TestProcessFile(t *testing.T) {
	assert := assert.New(t)
	testFilePath := "../../logs/server1.log"
	_, err := processFile(testFilePath)
	assert.NoError(err)
}
