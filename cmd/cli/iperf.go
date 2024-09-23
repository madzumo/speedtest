package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/BGrewell/go-iperf"
	"github.com/madzumo/speedtest/internal/helpers"
)

func runIperf(serverIP string, doDownloadTest bool, portNumber int, transmissionMSS int) bool {
	if !helpers.IsPortOpen(serverIP, portNumber) {
		cp := helpers.NewPromptColor()
		cp.Error.Println("Server unavailable. Iperf Server Client could be turned off.")
		return false
	}
	direction := "🖥️Client->🐒Server (Upload)"
	c := iperf.NewClient(serverIP)
	c.SetJSON(true)
	c.SetIncludeServer(false) //true
	c.SetTimeSec(10)
	c.SetInterval(1)
	c.SetPort(portNumber) //5201
	if transmissionMSS != 0 {
		c.SetMSS(transmissionMSS) //0
	}
	if doDownloadTest {
		c.SetReverse(true)
		c.SetStreams(4) //
		direction = "🐒Server->Client🖥️ (Download)"
	}
	err := c.Start()
	if err != nil {
		fmt.Printf("failed to start client: %v\n", err)
		os.Exit(-1)
	}

	<-c.Done

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
				logX := fmt.Sprintf("%s: %.2f Mbps (MSS:%d)", direction, mbps, transmissionMSS)
				writeLogFile(logX)
			}
		}
	}
	return true
}
