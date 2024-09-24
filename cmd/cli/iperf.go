package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BGrewell/go-iperf"
	"github.com/madzumo/speedtest/internal/bubbles"
	"github.com/madzumo/speedtest/internal/helpers"
)

func runIperf(serverIP string, doDownloadTest bool, portNumber int, transmissionMSS int) (bool, int) {
	var bubbleText string
	if doDownloadTest {
		bubbleText = "Iperf test Download..."
	} else {
		bubbleText = "Iperf test Upload..."
	}
	quit := make(chan struct{})
	go bubbles.ShowSpinner(quit, bubbleText, "46") // Run spinner in a goroutine

	if !helpers.IsPortOpen(serverIP, portNumber) {
		close(quit)
		time.Sleep(2 * time.Second)
		fmt.Println(lipErrorStyle.Render("Server unavailable. Iperf Server Client could be turned off."))
		return false, 0
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
		// os.Exit(-1)
		return false, 0
	}

	<-c.Done

	reportX := c.Report()
	var reportData map[string]interface{}
	err = json.Unmarshal([]byte(reportX.String()), &reportData)
	if err != nil {
		fmt.Printf("failed to parse JSON report: %v\n", err)
		return false, 0
	}

	close(quit)
	time.Sleep(2 * time.Second)

	if end, ok := reportData["end"].(map[string]interface{}); ok {
		if sumSent, ok := end["sum_sent"].(map[string]interface{}); ok {
			if bitsPerSecond, ok := sumSent["bits_per_second"].(float64); ok {
				mbps := bitsPerSecond / (1024 * 1024)
				if mbps <= 0 {
					fmt.Println("Server is busy. Wait for 10 seconds")
					return false, 1
				}
				resultData := fmt.Sprintf("%s: %.2f Mbps (MSS:%d)", direction, mbps, transmissionMSS)
				fmt.Println(resultData)
				writeLogFile(resultData)
			}
		}
	}
	return true, 0
}
