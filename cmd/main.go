package main

import (
	"fmt"
	"os"
	"time"

	"github.com/screenshotjy/logquery/pkg/logquery"
)

func main() {
	logQuery, err := logquery.NewLogQuery(map[string]string{
		"server1":   "./logs/server1.log",
		"db_server": "./logs/db_server.log",
	})

	if err != nil {
		fmt.Printf("Error in creating LogQuery, %s", err)
		os.Exit(1)
	}

	logQuery.Query(time.Now().Add(-time.Hour*24), 100, []string{"server1", "db_server"}, logquery.Info)

}
