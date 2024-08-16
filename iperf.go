package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/BGrewell/go-iperf"
)

var logFileName = "iperf3_report.txt"

var blockWindow = map[int][]int{
	0:  {1, 60},
	1:  {1, 4},
	2:  {5, 9},
	3:  {10, 14},
	4:  {15, 19},
	5:  {20, 24},
	6:  {25, 29},
	7:  {30, 34},
	8:  {35, 39},
	9:  {40, 44},
	10: {45, 49},
	11: {50, 54},
	12: {55, 60},
}

func runClient(serverIP string, doUploadTest bool) bool {
	direction := "DOWNload"
	c := iperf.NewClient(serverIP)
	c.SetJSON(true)
	c.SetIncludeServer(false) //true
	c.SetStreams(1)           // 4
	c.SetTimeSec(30)
	c.SetInterval(1)
	if doUploadTest {
		c.SetReverse(true)
		direction = "UPload"
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
		if sumSent, ok := end["sum_sent"].(map[string]interface{}); ok {
			if bitsPerSecond, ok := sumSent["bits_per_second"].(float64); ok {
				mbps := bitsPerSecond / (1024 * 1024)
				if mbps <= 0 {
					return false
				}
				if _, err := fmt.Fprintf(fileWriter, "[%s] %s Rate: %.2f Mbps\n", currentTime, direction, mbps); err != nil {
					fmt.Printf("failed to write to file: %v\n", err)
				} else {
					fmt.Printf("[%s] %s Rate: %.2f Mbps\n", currentTime, direction, mbps)
					// fmt.Print(c.Report().String())
				}
			}
		}
	}
	return true
}

func getBlockSelectWindow(blockSelect int) bool {
	currentTime := time.Now()
	if currentTime.Minute() >= blockWindow[blockSelect][0] && currentTime.Minute() <= blockWindow[blockSelect][1] {
		return true
	}
	return false
}
