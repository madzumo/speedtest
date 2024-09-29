package tests

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/BGrewell/go-iperf"
	hp "github.com/madzumo/speedtest/internal/helpers"
)

func IperfTest(serverIP string, doDownloadTest bool, portNumber int, transmissionMSS int) (bool, string) {

	if !hp.IsPortOpen(serverIP, portNumber) {
		time.Sleep(1 * time.Second)
		return false, "Server unreachable. Iperf on Server could be turned off. Retry in 10 seconds."
	}
	direction := "Iperf PC->Server (Upload)"
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
		direction = "Iperf Server->PC (Download)"
	}
	err := c.Start()
	if err != nil {
		return false, fmt.Sprintf("failed to start client: %v\n", err)
	}

	<-c.Done

	reportX := c.Report()
	var reportData map[string]interface{}
	err = json.Unmarshal([]byte(reportX.String()), &reportData)
	if err != nil {
		return false, fmt.Sprintf("failed to parse JSON report: %v\n", err)
	}

	var testResult string
	if end, ok := reportData["end"].(map[string]interface{}); ok {
		if sumSent, ok := end["sum_sent"].(map[string]interface{}); ok {
			if bitsPerSecond, ok := sumSent["bits_per_second"].(float64); ok {
				mbps := bitsPerSecond / (1024 * 1024)
				if mbps <= 0 {
					return false, "Server is busy. Retry in 10 Seconds."
				}
				testResult = fmt.Sprintf("%s: %.2f Mbps (MSS:%d)", direction, mbps, transmissionMSS)
				hp.WriteLogFile(fmt.Sprintf("🐒%s", testResult))
			}
		}
	}
	return true, testResult
}
