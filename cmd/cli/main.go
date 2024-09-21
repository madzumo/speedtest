package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/madzumo/speedtest/internal/helpers"
	"github.com/playwright-community/playwright-go"
)

var (
	serverIP = "0.0.0.0"
	// blockSelect     int
	testInterval                     = 0
	portNumber                       = 5201
	transmissionMSS                  = 0
	x                    interface{} = 42
	internetTestProvider             = "CloudFlare"
)

func main() {
	speedApp()
}

func speedApp() {
	for {

		helpers.ClearTerminalScreen()
		userSelect := printMenu()
		cp := helpers.NewPromptColor()
		switch userSelect {
		case 1:
			cp.Normal.Print("Enter Server IP: ")
			fmt.Scan(&serverIP)
			if net.ParseIP(serverIP) == nil {
				serverIP = ""
				fmt.Println("Invalid IP address.")
				helpers.PauseTerminalScreen()
			}
		case 2:
			cp.Normal.Print("Enter Port Number: ")
			fmt.Scan(&portNumber)
			x = portNumber
			if _, ok := x.(int); !ok {
				portNumber = 5201
				fmt.Print("Invalid Value. Must be numeric.")
				helpers.PauseTerminalScreen()
			}
		case 3:
			cp.Normal.Print("Enter Test Interval in Minutes: ")
			fmt.Scan(&testInterval)
			x = testInterval
			if _, ok := x.(int); !ok {
				testInterval = 10
				fmt.Print("Invalid Value. Must be numeric.")
				helpers.PauseTerminalScreen()
			}
		case 4:
			cp.Normal.Print("Enter Max Segment Size: ")
			fmt.Scan(&transmissionMSS)
			x = transmissionMSS
			if _, ok := x.(int); !ok {
				transmissionMSS = 1460
				fmt.Print("Invalid Value. Must be numeric.")
				helpers.PauseTerminalScreen()
			}
		case 5:
			if internetTestProvider == "CloudFlare" {
				internetTestProvider = "SpeedTest.net"
			} else {
				internetTestProvider = "CloudFlare"
			}
		case 6, 8:
			// setupTestsHeader()
			for {
				if userSelect == 8 {
					//internet speed test
					var testResult string
					if internetTestProvider == "CloudFlare" {
						testResult = cfTest(false)
					} else {
						testResult = runSpeedTestNet()
					}
					writeLogFile(testResult)
				}
				if !isPortOpen(serverIP, portNumber, 5*time.Second) {
					cp.Error.Printf("Cannot connect to iperf3 server on: %s Server may not be running\n", serverIP)
					helpers.PauseTerminalScreen()
					break
				} else {
					//iperf
					if runClient(serverIP, false) {
						if runClient(serverIP, true) {
							time.Sleep(time.Duration(testInterval) * time.Minute)
						}
					} else {
						fmt.Println("iperf server busy. retry in 10 seconds")
						time.Sleep(10 * time.Second)
					}
				}
				if testInterval == 0 {
					helpers.PauseTerminalScreen()
					break
				}
			}
		case 7:
			speedTests()
			helpers.PauseTerminalScreen()
		case 9:
			return
		}
	}
}

func setLogFileName() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error gettign hostname of client:", err)
	} else {
		return fmt.Sprintf("iperf3_%s.txt", hostname)
	}
	return ""
}

func writeLogFile(logData string) {
	logFileName := setLogFileName()
	fileWriter, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("failed to create/open Log file: %v\n", err)
	}
	defer fileWriter.Close()

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	if _, err := fmt.Fprintf(fileWriter, "[%s]%s\n", currentTime, logData); err != nil {
		fmt.Printf("failed to write to Log file: %v\n", err)
	}

	fmt.Printf("[%s]%s\n", currentTime, logData)
}

func speedTests() {
	helpers.ClearTerminalScreen()
	// Install Playwright if not already installed
	if err := playwright.Install(); err != nil {
		log.Fatalf("could not install Playwright: %v", err)
	}
	helpers.ClearTerminalScreen()
	writeLogFile(cfTest(false))
	writeLogFile(mlTest(false))
}
