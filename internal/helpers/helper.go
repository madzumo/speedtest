package helpers

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
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
	LipStandardStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Bold(true)
	LipStandard2Style      = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	LipHeaderStyle         = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("127"))
	LipConfigStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("112"))
	LipConfigSettingsStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	LipConfigSMTPStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("184"))
	LipOutputStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(lipgloss.Color("22"))
	LipErrorStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(lipgloss.Color("196")) //231 white
	LipSystemMsgStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("170")) //232 black
	LipFooterStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	LipResetStyle          = lipgloss.NewStyle()
)

type EmailJob struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Subject    string `json:"subject"`
	Body       string `json:"body"`
	SMTPHost   string `json:"smtpHost"`
	SMTPPort   string `json:"smtpPort"`
	UserName   string `json:"userName"`
	PassWord   string `json:"passWord"`
	UseOutlook bool   `json:"useOutlook"`
	Attachment string `json:"attachment"`
}

func NewEmailJob() *EmailJob {
	return &EmailJob{
		To: "",
	}
}

func (e *EmailJob) SendSMTP() string {
	// Set up authentication information.
	auth := smtp.PlainAuth("", e.UserName, e.PassWord, e.SMTPHost)

	// Define the email headers and body.
	from := e.From
	to := []string{e.To}
	subject := "Speed Test Report\n"
	body := "Speed Test Report Incoming!.\n"

	// Compose the message.
	message := []byte(subject + "\n" + body)

	// Set up the SMTP server and port.
	smtpAddr := fmt.Sprintf("%s:%s", e.SMTPHost, e.SMTPPort)

	// Send the email.
	err := smtp.SendMail(smtpAddr, auth, from, to, message)
	if err != nil {
		return fmt.Sprintf("Failed to send email: %v", err)
	}

	return "Email sent successfully!"
}

func (e *EmailJob) SentOutlook(attachmentPath, sendTO string) string {
	// // Try to start Outlook programmatically if it's not open
	// err := exec.Command("outlook.exe").Start()
	// if err != nil {
	// 	return fmt.Sprintf("Failed to start Outlook: %v", err)
	// }

	// Initialize COM
	err := ole.CoInitialize(0)
	if err != nil {
		return fmt.Sprintf("COM initialization failed: %v", err)
	}
	defer ole.CoUninitialize()

	// Create a new COM object for Outlook
	unknown, err := oleutil.CreateObject("Outlook.Application")
	if err != nil {
		return fmt.Sprintf("Failed to create Outlook object: %v", err)
	}
	defer unknown.Release()

	// Get the Outlook Application interface
	outlookApp, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Sprintf("Failed to get IDispatch for Outlook: %v", err)
	}
	defer outlookApp.Release()

	// Create a new MailItem
	mailItem, err := oleutil.CallMethod(outlookApp, "CreateItem", 0) // 0 means olMailItem
	if err != nil {
		return fmt.Sprintf("Failed to create MailItem: %v", err)
	}
	mail := mailItem.ToIDispatch()
	defer mail.Release()

	// Set the email properties
	oleutil.PutProperty(mail, "Subject", "Speed Test Report")
	// oleutil.PutProperty(mail, "Body", "Speed Test Report Incoming!")
	oleutil.PutProperty(mail, "To", sendTO)

	if attachmentPath != "" {
		// //1.Embed contents of a text file
		// fileContent, err := os.ReadFile(attachmentPath)
		// if err != nil {
		// 	return fmt.Sprintf("Failed to read the text file: %v", err)
		// }
		// oleutil.PutProperty(mail, "Body", string(fileContent))

		// Option 2: Add an attachment
		attachments := oleutil.MustGetProperty(mail, "Attachments").ToIDispatch()
		defer attachments.Release()
		// Add the file as an attachment
		_, err = oleutil.CallMethod(attachments, "Add", attachmentPath)
		if err != nil {
			return fmt.Sprintf("Failed to add attachment: %v", err)
		}
	}

	// Send the email
	_, err = oleutil.CallMethod(mail, "Send")
	if err != nil {
		return fmt.Sprintf("Failed to send email: %v", err)
	}

	return "Email sent successfully!"
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
	// ClearTerminalScreen()
	return greatSuccess
}

func SetLogFileName() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname of client:", err)
	} else {
		return fmt.Sprintf("Speed_%s.txt", hostname)
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
