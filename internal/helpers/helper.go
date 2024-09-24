package helpers

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	MenuHeader = `
                                                                  
 ░▒▓███████▓▒░▒▓███████▓▒░░▒▓████████▓▒░▒▓████████▓▒░▒▓███████▓▒░  
░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
 ░▒▓██████▓▒░░▒▓███████▓▒░░▒▓██████▓▒░ ░▒▓██████▓▒░ ░▒▓█▓▒░░▒▓█▓▒░ 
       ░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
       ░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓███████▓▒░░▒▓█▓▒░      ░▒▓████████▓▒░▒▓████████▓▒░▒▓███████▓▒░  
                                                                  
`
	lipStandardStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("117"))
	lipResetStyle    = lipgloss.NewStyle() // No styling
)

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
	fmt.Print("\033[H\033[2J")
	// fmt.Println("Have cleared")
	// cmd := exec.Command("clear") // works on Linux/macOS
	// cmd.Stdout = os.Stdout
	// cmd.Run()
}

func PauseTerminalScreen() {
	fmt.Println(lipResetStyle.Render("\n"))
	fmt.Println(lipStandardStyle.Render("Press 'Enter' to continue...."))
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

func SetPEMfiles() {
	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	// Find .pem files in the directory
	matches, err := filepath.Glob(filepath.Join(dir, "*.pem"))
	if err != nil {
		// fmt.Println("Error searching for .pem files:", err)
		return
	}

	// If a .pem file was found, set the environment variable
	if len(matches) > 0 {
		err = os.Setenv("NODE_EXTRA_CA_CERTS", matches[0]) // Use the first .pem file found
		if err != nil {
			// fmt.Println("Error setting environment variable:", err)
		} else {
			// fmt.Println("Environment variable set:", os.Getenv("NODE_EXTRA_CA_CERTS"))
		}
	} else {
		// fmt.Println("No .pem files found.")
	}
}
