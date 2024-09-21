package helpers

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/fatih/color"
)

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
	fmt.Print("\033[H\033[2J")
}

func PauseTerminalScreen() {
	pc := NewPromptColor()
	pc.Normal.Println("Enter 'q' to continue....")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('q')
}
