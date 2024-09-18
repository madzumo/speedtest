package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/BGrewell/go-iperf"
)

var logFileName = "iperf3_report.txt"

func runClient(serverIP string, doDownloadTest bool) bool {
	direction := "Upload"
	c := iperf.NewClient(serverIP)
	c.SetJSON(true)
	c.SetIncludeServer(true) //true
	c.SetTimeSec(10)
	c.SetInterval(1)
	c.SetPort(portNumber) //5201
	// c.SetMSS(transmissionMSS) //0
	if doDownloadTest {
		c.SetReverse(true)
		c.SetStreams(4) //
		direction = "Download"
	}
	err := c.Start()
	if err != nil {
		fmt.Printf("failed to start client: %v\n", err)
		os.Exit(-1)
	}

	<-c.Done

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	fileWriter, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("failed to open Log file: %v\n", err)
		return false
	}
	defer fileWriter.Close()

	reportX := c.Report()
	var reportData map[string]interface{}
	err = json.Unmarshal([]byte(reportX.String()), &reportData)
	if err != nil {
		fmt.Printf("failed to parse JSON report: %v\n", err)
		return false
	}

	if end, ok := reportData["end"].(map[string]interface{}); ok {
		var bitsPerSecond float64
		if doDownloadTest {
			if sumReceived, ok := end["sum_received"].(map[string]interface{}); ok {
				bitsPerSecond = sumReceived["bits_per_second"].(float64)
			}
		} else {
			if sumSent, ok := end["sum_sent"].(map[string]interface{}); ok {
				bitsPerSecond = sumSent["bits_per_second"].(float64)
			}
		}

		if bitsPerSecond > 0 {
			mbps := bitsPerSecond / (1024 * 1024)
			if _, err := fmt.Fprintf(fileWriter, "[%s] %s Bitrate: %.2f Mbps (MSS:%d)\n", currentTime, direction, mbps, transmissionMSS); err != nil {
				fmt.Printf("failed to write to file: %v\n", err)
			} else {
				fmt.Printf("[%s] %s Bitrate: %.2f Mbps\n", currentTime, direction, mbps)
			}
		}
	}
	return true
}
