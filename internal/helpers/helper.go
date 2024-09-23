package helpers

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/fatih/color"
)

var MenuHeader = `
                                                                  
 ░▒▓███████▓▒░▒▓███████▓▒░░▒▓████████▓▒░▒▓████████▓▒░▒▓███████▓▒░  
░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
 ░▒▓██████▓▒░░▒▓███████▓▒░░▒▓██████▓▒░ ░▒▓██████▓▒░ ░▒▓█▓▒░░▒▓█▓▒░ 
       ░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
       ░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓███████▓▒░░▒▓█▓▒░      ░▒▓████████▓▒░▒▓████████▓▒░▒▓███████▓▒░  
                                                                  
`

// Holds different color configuration to colorize Terminal Prompts
type PromptColor struct {
	//WHITE on -> BLUE
	Normal *color.Color
	//WHITE on -> RED + Bold
	Error *color.Color
	//MAGENTA + Bold
	Notify1 *color.Color
	//BLUE + Bold
	Notify2 *color.Color
	//RED + Bold
	Notify3 *color.Color
	//GREEN + Bold
	Notify4 *color.Color
	//BLACK on -> Yellow + Bold
	Special *color.Color
}

// Initialize Prompt Color
func NewPromptColor() *PromptColor {
	return &PromptColor{
		Normal:  color.New(color.BgBlue).Add(color.FgWhite),
		Error:   color.New(color.BgRed).Add(color.FgWhite).Add(color.Bold),
		Notify1: color.New(color.FgMagenta).Add(color.Bold),
		Notify2: color.New(color.FgBlue).Add(color.Bold),
		Special: color.New(color.BgYellow).Add(color.FgBlack).Add(color.Bold),
		Notify3: color.New(color.FgRed).Add(color.Bold),
		Notify4: color.New(color.FgHiGreen).Add(color.Bold),
	}
}

func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func ClearTerminalScreen() {
	// fmt.Println("Going to clear")
	// fmt.Print("\033[H\033[2J")
	// fmt.Println("Have cleared")
	// cmd := exec.Command("clear") // works on Linux/macOS
	// cmd.Stdout = os.Stdout
	// cmd.Run()
}

func PauseTerminalScreen() {
	pc := NewPromptColor()
	pc.Notify2.Printf("\nPress 'Enter' to continue....")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func IsPortOpen(serverIP string, port int) bool {
	address := fmt.Sprintf("%s:%d", serverIP, port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func ScanCACerts() (foundCert bool) {

	return false
}

func UseCACerts() bool {

	return false
}
