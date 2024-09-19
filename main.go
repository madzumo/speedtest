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
	testInterval                = 10
	portNumber                  = 5201
	transmissionMSS             = 0
	cPrompt                     = color.New(color.BgMagenta)
	cError                      = color.New(color.BgBlue)
	cPromptFG                   = color.New(color.FgMagenta)
	x               interface{} = 42
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
				fmt.Println("Invalid IP address...Press Enter to continue.")
				reader := bufio.NewReader(os.Stdin)
				_, _ = reader.ReadString('\n')
			}
		case 2:
			cPrompt.Print("Enter Port Number: ")
			fmt.Scan(&portNumber)
			x = portNumber
			if _, ok := x.(int); !ok {
				portNumber = 5201
				fmt.Print("Invalid Value. Must be numeric...Press Enter to continue.")
				reader := bufio.NewReader(os.Stdin)
				_, _ = reader.ReadString('\n')
			}
		case 3:
			cPrompt.Print("Enter Test Interval in Minutes: ")
			fmt.Scan(&testInterval)
			x = testInterval
			if _, ok := x.(int); !ok {
				testInterval = 10
				fmt.Print("Invalid Value. Must be numeric...Press Enter to continue.")
				reader := bufio.NewReader(os.Stdin)
				_, _ = reader.ReadString('\n')
			}
		case 4:
			cPrompt.Print("Enter Max Segment Size: ")
			fmt.Scan(&transmissionMSS)
			x = transmissionMSS
			if _, ok := x.(int); !ok {
				transmissionMSS = 1460
				fmt.Print("Invalid Value. Must be numeric...Press Enter to continue.")
				reader := bufio.NewReader(os.Stdin)
				_, _ = reader.ReadString('\n')
			}
		case 5, 7:
			setupTestsHeader()
			if !isPortOpen(serverIP, portNumber, 5*time.Second) {
				cError.Printf("Cannot connect to iperf3 server on: %s Server may not be running\nEnter 'q' to return\n", serverIP)
				reader := bufio.NewReader(os.Stdin)
				_, _ = reader.ReadString('q')
			} else {
				for {
					if userSelect == 7 {
						//internet speed test
						testResult := runSpeedTestNet()
						writeLogFile(testResult)
					}
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
			}
		case 6:
			setupTestsHeader()
			for {
				testResult := runSpeedTestNet()
				writeLogFile(testResult)
				time.Sleep(time.Duration(testInterval) * time.Minute)
			}
		// case 7:
		// 	var urlTest string
		// 	fmt.Print("Input URL to test:")
		// 	fmt.Scan(&urlTest)
		// 	doPlay(urlTest)
		// 	reader2 := bufio.NewReader(os.Stdin)
		// 	_, _ = reader2.ReadString('q')
		case 8:
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
