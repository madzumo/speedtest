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
	"github.com/playwright-community/playwright-go"
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
	LipStandardStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("117"))
	LipHeaderStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("127"))
	LipConfigStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("112"))
	LipOutputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(lipgloss.Color("22"))
	LipErrorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(lipgloss.Color("196")) //231 white
	LipSystemMsgStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("170")) //232 black
	LipResetStyle     = lipgloss.NewStyle()
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
	fmt.Println(LipResetStyle.Render("\n"))
	fmt.Println(LipStandardStyle.Render("Press 'Enter' to continue...."))
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

func InstallPlaywright() (greatSuccess bool) {
	greatSuccess = true
	if err := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}}); err != nil {
		fmt.Println(LipErrorStyle.Render(fmt.Sprintf("could not install Playwright: %v\n", err)))
		PauseTerminalScreen()
		greatSuccess = false
	}
	ClearTerminalScreen()
	return greatSuccess
}

func SetLogFileName() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname of client:", err)
	} else {
		return fmt.Sprintf("SpeedTest_%s.txt", hostname)
	}
	return ""
}

func WriteLogFile(logData string) {
	logFileName := SetLogFileName()
	fileWriter, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("failed to create/open Log file: %v\n", err)
	}
	defer fileWriter.Close()

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	if _, err := fmt.Fprintf(fileWriter, "[%s]%s\n", currentTime, logData); err != nil {
		fmt.Printf("failed to write to Log file: %v\n", err)
	}
}
