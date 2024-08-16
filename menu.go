package main

import (
	"fmt"
)

var lightningBolt = `
.----------------.  .----------------.  .----------------.  .----------------.  .----------------. 
| .--------------. || .--------------. || .--------------. || .--------------. || .--------------. |
| |    _______   | || |   ______     | || |  _________   | || |  _________   | || |  ________    | |
| |   /  ___  |  | || |  |_   __ \   | || | |_   ___  |  | || | |_   ___  |  | || | |_   ___  .  | |
| |  |  (__ \_|  | || |    | |__) |  | || |   | |_  \_|  | || |   | |_  \_|  | || |   | |    . \ | |
| |   '.___ -.   | || |    |  ___/   | || |   |  _|  _   | || |   |  _|  _   | || |   | |    | | | |
| |  | \____) |  | || |   _| |_      | || |  _| |___/ |  | || |  _| |___/ |  | || |  _| |___.' / | |
| |  |_______.'  | || |  |_____|     | || | |_________|  | || | |_________|  | || | |________.'  | |
| |              | || |              | || |              | || |              | || |              | |
| '--------------' || '--------------' || '--------------' || '--------------' || '--------------' |
'----------------'  '----------------'  '----------------'  '----------------'  '----------------' 
`

func printMenu() int {
	fmt.Println("==========================================")
	fmt.Println(lightningBolt)
	fmt.Println("==========================================")
	fmt.Println("1. Change Server IP")
	fmt.Println("2. Change Block Time (1-12)")
	fmt.Println("3. Change Port Number")
	fmt.Println("4. Change Test Interval")
	fmt.Println("5. RUN Client")
	fmt.Println("6. QUIT")
	fmt.Println("==========================================")
	fmt.Printf("Server IP: %s\n", serverIP)
	fmt.Printf("Block Time: %d\n", blockSelect)
	fmt.Printf("Port Number: %d\n", portNumber)
	fmt.Printf("Test Interval: %d min\n", testInterval)
	fmt.Println("==========================================")

	menuOption := 0
	fmt.Print("Enter Menu Option: ")
	fmt.Scan(&menuOption)
	return menuOption
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
