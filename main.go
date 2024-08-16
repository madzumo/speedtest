package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

var (
	serverIP     = "10.132.132.190"
	blockSelect  int
	testInterval = 10
	portNumber   = 5201
)

func main() {
	for {
		clearScreen()
		switch printMenu() {
		case 1:
			fmt.Print("Enter Server IP: ")
			fmt.Scan(&serverIP)
			if net.ParseIP(serverIP) == nil {
				serverIP = ""
				fmt.Println("Invalid IP address...Press Enter to continue.")
				reader := bufio.NewReader(os.Stdin)
				_, _ = reader.ReadString('\n')
			}
		case 2:
			fmt.Print("Enter Time Block #: ")
			fmt.Scan(&blockSelect)
			_, exists := blockWindow[blockSelect]
			if !exists {
				blockSelect = 0
				fmt.Print("Invalid Time block #...Press Enter to continue.")
				reader := bufio.NewReader(os.Stdin)
				_, _ = reader.ReadString('\n')
			}
		case 3:
			fmt.Print("Enter Port Number: ")
			fmt.Scan(&portNumber)
		case 4:
			fmt.Print("Enter Test Interval in Minutes: ")
			fmt.Scan(&testInterval)
			if testInterval > 60 {
				testInterval = 15
				fmt.Print("Invalid Test Interval Minutes...Press Enter to continue.")
				reader := bufio.NewReader(os.Stdin)
				_, _ = reader.ReadString('\n')
			}
		case 5:
			clearScreen()
			fmt.Println("Running Speed Tests...")
			fmt.Println("(your work is done. go get some coffee)")
			for {
				if getBlockSelectWindow(blockSelect) {
					if runClient(serverIP, false) {
						if runClient(serverIP, true) {
							time.Sleep(time.Duration(testInterval) * time.Minute)
						}
					} else {
						fmt.Println("busy. retry in 10 seconds")
						time.Sleep(10 * time.Second)
					}
				}
			}
		case 6:
			return
		}
	}
}
