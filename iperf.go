package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/BGrewell/go-iperf"
)

var logFileName = "iperf3_report.txt"

func runClient(serverIP string, doDownloadTest bool) bool {
	direction := "🖥️Client->Server (Upload)"
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
		direction = "Server->Client🖥️ (Download)"
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
				if _, err := fmt.Fprintf(fileWriter, "[%s] %s Rate: %.2f Mbps (MSS:%d)\n", currentTime, direction, mbps, transmissionMSS); err != nil {
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

func isPortOpen(serverIP string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", serverIP, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
