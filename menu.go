package main

import (
	"fmt"

	"github.com/fatih/color"
)

var Red = "\033[31m"
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
	c1 := color.New(color.BgRed)
	c2 := color.New(color.FgGreen).Add(color.Bold)
	c3 := color.New(color.FgHiBlue).Add(color.Bold)
	c4 := color.New(color.FgHiYellow)
	c5 := color.New(color.FgRed)
	c1.Println(lightningBolt)
	fmt.Println("==========================================")
	c2.Printf("Server IP: %s\n", serverIP)
	c2.Printf("Block Time: %d\n", blockSelect)
	c2.Printf("Port Number: %d\n", portNumber)
	c2.Printf("Test Interval: %d min\n", testInterval)
	fmt.Println("==========================================")
	c3.Println("1. Change Server IP")
	c3.Println("2. Change Block Time (1-12)")
	c3.Println("3. Change Port Number")
	c3.Println("4. Change Test Interval")
	c4.Println("5. RUN Client")
	c5.Println("6. QUIT")
	fmt.Println("==========================================")

	menuOption := 0
	fmt.Print("Enter Menu Option: ")
	fmt.Scan(&menuOption)
	return menuOption
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
