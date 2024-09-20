package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/fatih/color"
)

var (
	serverIP = "0.0.0.0"
	// blockSelect     int
	testInterval                     = 0
	portNumber                       = 5201
	transmissionMSS                  = 0
	cPrompt                          = color.New(color.BgMagenta)
	cError                           = color.New(color.BgBlue)
	cPromptFG                        = color.New(color.FgMagenta)
	x                    interface{} = 42
	internetTestProvider             = "CloudFlare"
)

func main() {
	speedApp()
}

func speedApp() {
	for {
		clearScreen()
		userSelect := printMenu()
		switch userSelect {
		case 1:
			cPrompt.Print("Enter Server IP: ")
			fmt.Scan(&serverIP)
			if net.ParseIP(serverIP) == nil {
				serverIP = ""
				fmt.Println("Invalid IP address.")
				pauseScreen()
			}
		case 2:
			cPrompt.Print("Enter Port Number: ")
			fmt.Scan(&portNumber)
			x = portNumber
			if _, ok := x.(int); !ok {
				portNumber = 5201
				fmt.Print("Invalid Value. Must be numeric.")
				pauseScreen()
			}
		case 3:
			cPrompt.Print("Enter Test Interval in Minutes: ")
			fmt.Scan(&testInterval)
			x = testInterval
			if _, ok := x.(int); !ok {
				testInterval = 10
				fmt.Print("Invalid Value. Must be numeric.")
				pauseScreen()
			}
		case 4:
			cPrompt.Print("Enter Max Segment Size: ")
			fmt.Scan(&transmissionMSS)
			x = transmissionMSS
			if _, ok := x.(int); !ok {
				transmissionMSS = 1460
				fmt.Print("Invalid Value. Must be numeric.")
				pauseScreen()
			}
		case 5:
			if internetTestProvider == "CloudFlare" {
				internetTestProvider = "SpeedTest.net"
			} else {
				internetTestProvider = "CloudFlare"
			}
		case 6, 8:
			setupTestsHeader()
			for {
				if userSelect == 8 {
					//internet speed test
					var testResult string
					if internetTestProvider == "CloudFlare" {
						testResult = cfTest()
					} else {
						testResult = runSpeedTestNet()
					}
					writeLogFile(testResult)
				}
				if !isPortOpen(serverIP, portNumber, 5*time.Second) {
					cError.Printf("Cannot connect to iperf3 server on: %s Server may not be running\n", serverIP)
					pauseScreen()
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
					pauseScreen()
					break
				}
			}
		case 7:
			setupTestsHeader()
			//internet speed test
			var testResult string
			if internetTestProvider == "CloudFlare" {
				testResult = cfTest()
			} else {
				testResult = runSpeedTestNet()
			}
			writeLogFile(testResult)
			pauseScreen()
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

func setupTestsHeader() {
	clearScreen()
	cPromptFG.Println("==============================================")
	cPrompt.Println("Running Speed Tests...")
	// fmt.Println("(your work here is done✅ go get some coffee☕)")
	cPromptFG.Println("==============================================")
}

func pauseScreen() {
	cPrompt.Println("Enter 'q' to continue....")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('q')
}
