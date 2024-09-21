package main

import (
	"fmt"

	"github.com/fatih/color"
)

var Red = "\033[31m"

var menuText = `
                                                                  
 ░▒▓███████▓▒░▒▓███████▓▒░░▒▓████████▓▒░▒▓████████▓▒░▒▓███████▓▒░  
░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
 ░▒▓██████▓▒░░▒▓███████▓▒░░▒▓██████▓▒░ ░▒▓██████▓▒░ ░▒▓█▓▒░░▒▓█▓▒░ 
       ░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
       ░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓███████▓▒░░▒▓█▓▒░      ░▒▓████████▓▒░▒▓████████▓▒░▒▓███████▓▒░  
                                                                  
`

func printMenu() int {
	c1 := color.New(color.BgRed)
	c2 := color.New(color.FgGreen).Add(color.Bold)
	c3 := color.New(color.FgBlue).Add(color.Bold)
	c4 := color.New(color.FgHiYellow)
	c5 := color.New(color.FgRed)
	// c6 := color.New(color.FgHiWhite).Add(color.Bold)
	c1.Println(menuText)
	fmt.Println("==========================================")
	c2.Printf("Iperf Server: %s\n", serverIP)
	c2.Printf("Port Number: %d\n", portNumber)
	c2.Printf("Repeat Test Every: %d min\n", testInterval)
	if transmissionMSS == 0 {
		c2.Printf("MSS (max segment size): Auto\n")
	} else {
		c2.Printf("MSS (max segment size): %d\n", transmissionMSS)
	}
	c2.Printf("Internet Test Provider: %s\n", internetTestProvider)
	fmt.Println("==========================================")
	c3.Println("1. Change Iperf Server")
	c3.Println("2. Change Port Number")
	c3.Println("3. Change Repeat Test Interval")
	c3.Println("4. Change MSS")
	c3.Println("5. Change Internet Test Provider")
	c4.Println("6. RUN Iperf Test Only")
	c4.Println("7. RUN Internet Speed Test Only")
	c2.Println("8. RUN ALL Tests")
	c5.Println("9. QUIT")
	fmt.Println("==========================================")

	menuOption := 0
	fmt.Print("Enter Menu Option: ")
	fmt.Scan(&menuOption)
	return menuOption
}
