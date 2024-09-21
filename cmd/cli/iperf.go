package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/BGrewell/go-iperf"
)

func runClient(serverIP string, doDownloadTest bool) bool {
	// cPrompt2.Println("Running Iperf Test")
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

func isPortOpen(serverIP string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", serverIP, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
