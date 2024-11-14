package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/badoux/checkmail"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	hp "github.com/madzumo/speedtest/internal/helpers"
	t "github.com/madzumo/speedtest/internal/tests"
)

const listHeight = 14

var (
	titleStyle          = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("111"))
	itemStyle           = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle   = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle     = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle           = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	menuMainColor       = "205"
	menuSettingsColor   = "111"
	menuSMTPcolor       = "184"
	textPromptColor     = "141" //"100" //nice: 141
	textInputColor      = "193" //"40" //nice: 193
	textErrorColorBack  = "1"
	textErrorColorFront = "15"
	textResultJob       = "141" //PINK"205"
	textJobOutcomeFront = "216"
	// txtJobOutcomeBack   = "205"

	menuTOP = []string{
		"Run ALL Tests",
		"Run Internet Speed Tests Only",
		"Run Iperf Test Only",
		"Configure Settings",
		"Save Settings",
	}
	menuSettings = [][]string{ //menu Text + prompt text
		{"Set Iperf Server IP", "Enter Iperf Server IP:"},
		{"Set Iperf Port number", "Enter Iperf Port Number:"},
		{"Configure Email Settings", ""},
		{"Toggle CloudFlare Test", ""},
		{"Toggle M-Labs Test", ""},
		{"Toggle Speedtest.net", ""},
		{"Toggle Show Browser on Speed Tests", ""},
		{"Set Repeat Test Interval", "Enter Repeat Interval in Seconds:"},
		{"Set MSS Size", "Enter Maximum Segment Size (MSS) for Iperf Test:"},
		{"Set Iperf Retry Timeout", "Enter Max seconds Iperf will Retry before cancelling:"},
	}
	menuSMTP = [][]string{ //menu Text + prompt text
		{"Toggle Email Service (SMTP/Outlook/OFF)", ""},
		{"Set SMTP Host", "Enter SMTP Host Address:"},
		{"Set SMTP Port", "Enter SMTP Port Number:"},
		// {"Set Auth Username", "Enter Username to Authenticate SMTP:"},
		{"Set From: Address", "Enter Sender address (From:) for sending reports:"},
		{"Set Auth Password or Open Relay", "Input Password to Authenticate SMTP. Leave blank for Open Relay:"},
		{"Set To: Address", "Enter Recipient address (To:) for sending reports:"},
		{"Send Test Email", ""},
		// {"Set E-Mail Subject", "Enter Subject title in outgoing report E-mail:"},
		// {"Set E-Mail Message","Enter Message for E-mail report:",},
	}
)

type MenuState int

const (
	StateMainMenu MenuState = iota
	StateSettingsMenu
	StateSpinner
	StateResultDisplay
	StateSMTPMenu
	StateTextInput
	StateInterval
)

type backgroundJobMsg struct {
	result string
}

type continueJobs struct {
	jobResult  string
	iperfError bool
}

type JobList int

const (
	CFtest JobList = iota
	MLTest
	NETtest
	IperfTest
	Settings
)

type MenuList struct {
	list                list.Model
	choice              string
	header              string
	headerIP            string
	state               MenuState
	prevState           MenuState
	prevMenuState       MenuState
	spinner             spinner.Model
	spinnerMsg          string
	backgroundJobResult string
	textInput           textinput.Model
	inputPrompt         string
	textInputError      bool
	configSettings      *configSettings
	jobOutcome          string
	jobsList            map[int]string
	iperfError          bool
	iperfErrorCount     int
	sendEmailActive     bool
}

func (m MenuList) Init() tea.Cmd {
	return nil
}

func (m MenuList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateMainMenu:
		return m.updateMainMenu(msg)
	case StateSettingsMenu:
		return m.updateSettingsMenu(msg)
	case StateSpinner:
		return m.updateSpinner(msg)
	case StateResultDisplay:
		return m.updateResultDisplay(msg)
	case StateSMTPMenu:
		return m.updateSMTPMenu(msg)
	case StateTextInput:
		return m.updateTextInput(msg)
	case StateInterval:
		return m.updateInterval(msg)
	default:
		return m, nil
	}
}

func (m *MenuList) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			err := clipboard.WriteAll(m.headerIP)
			if err != nil {
				fmt.Println("Failed to copy to clipboard:", err)
			}
		}
		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				switch m.choice {
				case menuTOP[0], menuTOP[1], menuTOP[2], menuTOP[4]:
					BuildJobList(m)
					m.prevState = m.state
					m.prevMenuState = m.state
					m.state = StateSpinner
					if (!m.configSettings.EmailSettings.UseOutlook && !m.configSettings.EmailSettings.UseSMTP) || m.choice == menuTOP[4] {
						m.sendEmailActive = false
					} else {
						m.sendEmailActive = true
					}
					return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob())
				case menuTOP[3]:
					m.prevState = m.state
					m.list.Title = "Main Menu->Settings"
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuSettingsColor))
					m.list.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(menuSettingsColor))
					m.state = StateSettingsMenu
					m.updateListItems()
					return m, nil
				}
			}
			return m, nil
		}
		// case jobListMsg:

		// 	// m.state = StateResultDisplay
		// 	// return m, nil
		// 	m.prevState = m.state
		// 	m.state = StateSpinner
		// 	return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob())
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MenuList) updateTextInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.textInputError = false
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			inputValue := m.textInput.Value() // User pressed enter, save the input

			switch m.inputPrompt {
			case menuSettings[1][1]:
				if i, err := strconv.Atoi(inputValue); err != nil { //validate port number
					m.backgroundJobResult = fmt.Sprintf("Invalid port number: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.IperfP = i
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					m.backgroundJobResult = fmt.Sprintf("Iperf Port Number set -> %s.", inputValue)
				}
			case menuSettings[0][1]:
				m.backgroundJobResult = fmt.Sprintf("Iperf Server IP set -> %s.", inputValue)
				m.configSettings.IperfS = inputValue
				m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)

			case menuSettings[7][1]:
				if i, err := strconv.Atoi(inputValue); err != nil {
					m.backgroundJobResult = fmt.Sprintf("Invalid entry for Seconds: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.RepeatInterval = i
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					m.backgroundJobResult = fmt.Sprintf("Repeat Test Interval -> %s seconds.", inputValue)
				}
			case menuSettings[8][1]:
				if i, err := strconv.Atoi(inputValue); err != nil {
					m.backgroundJobResult = fmt.Sprintf("Invalid entry for MSS: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.MSS = i
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					m.backgroundJobResult = fmt.Sprintf("MSS set -> %s", inputValue)
				}
			case menuSettings[9][1]:
				if i, err := strconv.Atoi(inputValue); err != nil {
					m.backgroundJobResult = fmt.Sprintf("Invalid entry for Iperf Retry Timeout: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.IperfTimeout = i
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					m.backgroundJobResult = fmt.Sprintf("Iperf Retry Timeout -> %s seconds", inputValue)
				}
			case menuSMTP[1][1]:
				m.configSettings.EmailSettings.SMTPHost = inputValue
				m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
				m.backgroundJobResult = fmt.Sprintf("SMTP Host set -> %s", inputValue)

			case menuSMTP[2][1]:
				if _, err := strconv.Atoi(inputValue); err != nil { //validate port number
					m.backgroundJobResult = fmt.Sprintf("Invalid port number: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.EmailSettings.SMTPPort = inputValue
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
					m.backgroundJobResult = fmt.Sprintf("SMTP Port Number set -> %s", inputValue)
				}
			// case menuSMTP[3][1]:
			// 	m.configSettings.EmailSettings.UserName = inputValue
			// 	m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
			// 	m.backgroundJobResult = fmt.Sprintf("SMTP Auth Username set -> %s", inputValue)

			case menuSMTP[4][1]:
				m.configSettings.EmailSettings.PassWord = inputValue
				if inputValue == "" {
					inputValue = "OpenRelay"
				}
				m.backgroundJobResult = fmt.Sprintf("SMTP Auth Password set -> %s", inputValue)

			case menuSMTP[3][1]:
				if err := checkmail.ValidateFormat(inputValue); err != nil {
					m.backgroundJobResult = fmt.Sprintf("Not a valid E-mail Address: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.EmailSettings.From = inputValue
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
					m.backgroundJobResult = fmt.Sprintf("Sender (From) address set -> %s", inputValue)
				}

			case menuSMTP[5][1]:
				if err := checkmail.ValidateFormat(inputValue); err != nil {
					m.backgroundJobResult = fmt.Sprintf("Not a valid E-mail Address: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.EmailSettings.To = inputValue
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
					m.backgroundJobResult = fmt.Sprintf("Recipient (To) address set -> %s", inputValue)
				}
			}
			m.prevState = m.state
			m.state = StateResultDisplay
			return m, nil
		case tea.KeyEsc:
			// m.state = StateSettingsMenu
			m.state = m.prevState
			return m, nil
		}
	}

	return m, cmd
}

func (m *MenuList) updateSettingsMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc", "ctrl+c":
			m.list.Title = "Main Menu"
			m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, false)
			selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuMainColor))
			m.list.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(menuMainColor))
			m.state = StateMainMenu
			m.updateListItems()
			return m, nil
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				choice := string(i)
				switch choice {
				case "Back to Main Menu":
					m.list.Title = "Main Menu"
					selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuMainColor))
					m.list.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(menuMainColor))
					m.state = StateMainMenu
					m.updateListItems()
					return m, nil
				case menuSettings[1][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSettings[1][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., 5201"
					m.textInput.Focus()
					m.textInput.CharLimit = 5 // Port numbers are up to 5 digits
					m.textInput.Width = 20
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSettings[0][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSettings[0][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., 192.168.1.1"
					m.textInput.Focus()
					m.textInput.CharLimit = 20
					m.textInput.Width = 20
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
				case menuSettings[7][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSettings[7][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., 10"
					m.textInput.Focus()
					m.textInput.CharLimit = 5
					m.textInput.Width = 20
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSettings[8][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSettings[8][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., 1450"
					m.textInput.Focus()
					m.textInput.CharLimit = 5
					m.textInput.Width = 20
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSettings[9][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSettings[9][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., 120"
					m.textInput.Focus()
					m.textInput.CharLimit = 5
					m.textInput.Width = 20
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSettings[2][0]:
					m.list.Title = "Main Menu->Settings->SMTP"
					selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuSMTPcolor))
					m.list.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(menuSMTPcolor))
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
					m.prevState = m.state
					m.state = StateSMTPMenu
					m.updateListItems()
					return m, nil
				case menuSettings[3][0]:
					if m.configSettings.CloudFrontTest {
						m.configSettings.CloudFrontTest = false
					} else {
						m.configSettings.CloudFrontTest = true
					}
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					return m, nil
				case menuSettings[4][0]:
					if m.configSettings.MLabTest {
						m.configSettings.MLabTest = false
					} else {
						m.configSettings.MLabTest = true
					}
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					return m, nil
				case menuSettings[5][0]:
					if m.configSettings.NetTest {
						m.configSettings.NetTest = false
					} else {
						m.configSettings.NetTest = true
					}
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					return m, nil
				case menuSettings[6][0]:
					if m.configSettings.ShowBrowser {
						m.configSettings.ShowBrowser = false
					} else {
						m.configSettings.ShowBrowser = true
					}
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					return m, nil
				default:
					// Simulate settings change
					m.prevState = m.state
					m.state = StateResultDisplay
					m.backgroundJobResult = fmt.Sprintf("%s updated successfully.", choice)
					return m, nil
				}
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MenuList) updateSpinner(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// case "q", "esc":
		// 	m.backgroundJobResult = "Job Cancelled"
		// 	m.state = StateResultDisplay
		// 	return m, nil
		default:
			// For other key presses, update the spinner
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case backgroundJobMsg:
		if m.sendEmailActive {
			m.configSettings.EmailSettings.Subject = "Speed Test Report"
			m.configSettings.EmailSettings.Body = "Speed Test report incoming!"
			//send e-mail in a go routine with Lock OS thread for bubble tea compat.
			resultChan := make(chan string)
			go func() {
				runtime.LockOSThread()
				defer runtime.UnlockOSThread()
				emailResult := ""
				if m.configSettings.EmailSettings.UseOutlook {
					emailResult = m.configSettings.EmailSettings.SendOutlook(false)
				}
				if m.configSettings.EmailSettings.UseSMTP {
					emailResult = m.configSettings.EmailSettings.SendSMTP(false)
				}
				resultChan <- emailResult
			}()
			m.backgroundJobResult = m.jobOutcome + "\n\n" + msg.result + "\n" + <-resultChan
		} else {
			m.backgroundJobResult = m.jobOutcome + "\n\n" + msg.result + "\n"
		}

		if m.configSettings.RepeatInterval > 0 {
			m.state = StateInterval
			return m, m.startInverval()
		} else {
			m.state = StateResultDisplay
			return m, nil
		}

	case continueJobs:
		m.jobOutcome += msg.jobResult + "\n"
		if msg.iperfError {
			m.iperfError = msg.iperfError
			m.iperfErrorCount += 1
		}

		return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob())
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m *MenuList) updateInterval(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.state = m.prevMenuState
			m.updateListItems()
			return m, nil
		}
	case continueJobs:
		BuildJobList(m)
		m.prevState = m.state
		// m.prevMenuState = m.state
		m.state = StateSpinner
		return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob())

	}
	return m, nil
}

func (m *MenuList) updateResultDisplay(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.textInputError {
				m.state = m.prevState
			} else {
				m.state = m.prevMenuState
			}
			m.updateListItems()
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *MenuList) updateSMTPMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc", "ctrl+c":
			m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
			m.list.Title = "Main Menu->Settings"
			selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuSettingsColor))
			m.list.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(menuSettingsColor))
			m.state = StateSettingsMenu
			m.updateListItems()
			return m, nil
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				choice := string(i)
				switch choice {
				case "Back to Main Menu":
					m.list.Title = "Main Menu"
					selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuMainColor))
					m.list.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(menuMainColor))
					m.state = StateMainMenu
					m.updateListItems()
					return m, nil
				case menuSMTP[1][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSMTP[1][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., mail.domain.com"
					m.textInput.Focus()
					m.textInput.CharLimit = 50
					m.textInput.Width = 50
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSMTP[2][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSMTP[2][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., 587"
					m.textInput.Focus()
					m.textInput.CharLimit = 5 // Port numbers are up to 5 digits
					m.textInput.Width = 20
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSMTP[4][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSMTP[4][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., pass123"
					m.textInput.Focus()
					m.textInput.CharLimit = 50
					m.textInput.Width = 50
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSMTP[3][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSMTP[3][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., Its_A_Me@domain.com"
					m.textInput.Focus()
					m.textInput.CharLimit = 50
					m.textInput.Width = 50
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSMTP[5][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSMTP[5][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., Joe_Mama@domain.com"
					m.textInput.Focus()
					m.textInput.CharLimit = 50
					m.textInput.Width = 50
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSMTP[0][0]:
					if !m.configSettings.EmailSettings.UseOutlook && !m.configSettings.EmailSettings.UseSMTP {
						m.configSettings.EmailSettings.UseSMTP = true
					} else if m.configSettings.EmailSettings.UseSMTP {
						m.configSettings.EmailSettings.UseSMTP = false
						m.configSettings.EmailSettings.UseOutlook = true
					} else {
						m.configSettings.EmailSettings.UseSMTP = false
						m.configSettings.EmailSettings.UseOutlook = false
					}
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
					return m, nil
				case menuSMTP[6][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateResultDisplay
					m.backgroundJobResult = "Activate Email Service in Email Settings"
					if m.configSettings.EmailSettings.UseOutlook || m.configSettings.EmailSettings.UseSMTP {
						m.configSettings.EmailSettings.Subject = "Testing Email in Speed3"
						m.configSettings.EmailSettings.Body = "Test Message"
						resultChan := make(chan string)
						go func() {
							runtime.LockOSThread()
							defer runtime.UnlockOSThread()

							emailResult := ""
							if m.configSettings.EmailSettings.UseOutlook {
								emailResult = m.configSettings.EmailSettings.SendOutlook(true)
							}
							if m.configSettings.EmailSettings.UseSMTP {
								emailResult = m.configSettings.EmailSettings.SendSMTP(true)
							}
							resultChan <- emailResult
						}()
						m.backgroundJobResult = <-resultChan
					}

					return m, nil
				default:
					// Simulate settings change
					m.prevState = m.state
					m.state = StateResultDisplay
					m.backgroundJobResult = fmt.Sprintf("%s updated successfully.", choice)
					return m, nil
				}
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}
func (m *MenuList) startInverval() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Duration(m.configSettings.RepeatInterval) * time.Second)
		return continueJobs{}
	}
}
func (m *MenuList) startBackgroundJob() tea.Cmd {
	return func() tea.Msg {
		var continueResult string
		iperfOut := false
		if len(m.jobsList) > 0 { //check each job in Order, process then exit for next pass
			if m.jobsList[0] != "" {
				m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("51")) //white = 231
				m.spinnerMsg = "Installing Browser Components"
				// Redirect stdout and stderr to suppress Playwright output
				stdout := os.Stdout
				stderr := os.Stderr
				// Create temporary files to capture the output
				tempOut, _ := os.CreateTemp("", "playwright-out-*")
				tempErr, _ := os.CreateTemp("", "playwright-err-*")
				// Redirect stdout and stderr
				os.Stdout = tempOut
				os.Stderr = tempErr

				pass, result := hp.InstallPlaywright()

				if !pass {
					continueResult = fmt.Sprintf("Error Installing Components need for Tests. Internet is inaccessible:\n%s", result)
					for k := range m.jobsList {
						delete(m.jobsList, k)
					}
				}
				// Reset stdout and stderr after installation
				os.Stdout = stdout
				os.Stderr = stderr
				// Clear the terminal to remove Playwright output
				// tea.ClearScreen()
				delete(m.jobsList, 0)
			} else if m.jobsList[1] != "" {
				m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
				m.spinnerMsg = "Running Cloudflare Speed test"
				continueResult = t.CFTest(m.configSettings.ShowBrowser)
				// time.Sleep(3 * time.Second)
				// continueResult = "CF Job is done!"
				delete(m.jobsList, 1)
			} else if m.jobsList[2] != "" {
				m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("129"))
				m.spinnerMsg = "Running MLab Speed test"
				continueResult = t.MLTest(m.configSettings.ShowBrowser)
				// time.Sleep(3 * time.Second)
				// continueResult = "MLab done"
				delete(m.jobsList, 2)
			} else if m.jobsList[3] != "" {
				m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
				m.spinnerMsg = "Running Speedtest.NET test"
				continueResult = t.NETTest()
				// time.Sleep(3 * time.Second)
				// continueResult = "NET done"
				delete(m.jobsList, 3)
			} else if m.jobsList[4] != "" {
				m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("156"))
				m.spinnerMsg = "Running Iperf test"
				if m.iperfError {
					time.Sleep(10 * time.Second)
				}
				pass, result := t.IperfTest(m.configSettings.IperfS, false, m.configSettings.IperfP, m.configSettings.MSS)
				continueResult = result
				if pass {
					pass, result = t.IperfTest(m.configSettings.IperfS, true, m.configSettings.IperfP, m.configSettings.MSS)
					continueResult += "\n" + result
				}

				if pass || m.iperfErrorCount >= (m.configSettings.IperfTimeout/15) {
					delete(m.jobsList, 4)
				} else {
					iperfOut = true
				}
			} else if m.jobsList[5] != "" {
				m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
				m.spinnerMsg = "Saving Settings"
				// m.spinner.Tick()
				time.Sleep(1 * time.Second)
				saveConfig(m.configSettings)
				delete(m.jobsList, 5)
			}
			return continueJobs{jobResult: continueResult, iperfError: iperfOut}
		} else {
			return backgroundJobMsg{result: "Job Complete!"}
		}
	}
}

func (m MenuList) View() string {
	switch m.state {
	case StateMainMenu, StateSettingsMenu, StateSMTPMenu:
		return m.header + "\n" + m.list.View()
	case StateSpinner:
		return m.viewSpinner()
	case StateResultDisplay:
		return m.viewResultDisplay()
	case StateTextInput:
		return m.viewTextInput()
	case StateInterval:
		return m.viewInterval()
	default:
		return "Unknown state"
	}
}

func (m MenuList) viewSpinner() string {
	// tea.ClearScreen()
	spinnerBase := fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.spinnerMsg)

	// return spinnerBase + m.jobOutcome
	return spinnerBase + lipgloss.NewStyle().Foreground(lipgloss.Color(textJobOutcomeFront)).Bold(true).Render(m.jobOutcome)
}

func (m MenuList) viewResultDisplay() string {
	outro := "Press 'esc' to return."
	outroRender := hp.LipStandardStyle.Render(outro)
	//lipgloss.NewStyle().Foreground(lipgloss.Color(textJobOutcomeFront)).Bold(true).Render(m.jobOutcome)
	if m.textInputError {
		m.backgroundJobResult = lipgloss.NewStyle().Foreground(lipgloss.Color(textErrorColorFront)).Background(lipgloss.Color(textErrorColorBack)).Bold(true).Render(m.backgroundJobResult)
	} else {
		m.backgroundJobResult = lipgloss.NewStyle().Foreground(lipgloss.Color(textResultJob)).Render(m.backgroundJobResult)
	}
	return fmt.Sprintf("\n\n%s\n\n%s", m.backgroundJobResult, outroRender)

	// //repeat interval
	// if m.configSettings.Interval > 0 {

	// }
}

func (m MenuList) viewTextInput() string {
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor)).Bold(true)
	return fmt.Sprintf("\n\n%s\n\n%s", promptStyle.Render(m.inputPrompt), m.textInput.View())

}

func (m MenuList) viewInterval() string {
	outro := fmt.Sprintf("\n\n%s\n\nWaiting...Next interval %v seconds", m.backgroundJobResult, m.configSettings.RepeatInterval)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(textResultJob)).Render(outro)
}

func (m *MenuList) updateListItems() {
	switch m.state {
	case StateMainMenu:
		items := []list.Item{}
		for _, value := range menuTOP {
			items = append(items, item(value))
		}
		m.list.SetItems(items)
	case StateSettingsMenu:
		items := []list.Item{}
		for _, value := range menuSettings {
			items = append(items, item(value[0]))
		}
		m.list.SetItems(items)
	case StateSMTPMenu:
		items := []list.Item{}
		for _, value := range menuSMTP {
			items = append(items, item(value[0]))
		}
		m.list.SetItems(items)
	}

	m.list.ResetSelected()
}

func BuildJobList(m *MenuList) {
	m.iperfError = false
	m.iperfErrorCount = 0
	m.jobOutcome = ""
	m.jobsList = map[int]string{}
	switch m.choice {
	case menuTOP[0]:
		m.jobsList[0] = "PLAY"
		if m.configSettings.CloudFrontTest {
			m.jobsList[1] = "CF"
		}
		if m.configSettings.MLabTest {
			m.jobsList[2] = "ML"
		}
		if m.configSettings.NetTest {
			m.jobsList[3] = "NET"
		}
		m.jobsList[4] = "IP"

	case menuTOP[1]:
		m.jobsList[0] = "PLAY"
		if m.configSettings.CloudFrontTest {
			m.jobsList[1] = "CF"
		}
		if m.configSettings.MLabTest {
			m.jobsList[2] = "ML"
		}
		if m.configSettings.NetTest {
			m.jobsList[3] = "NET"
		}
	case menuTOP[2]:
		m.jobsList[4] = "IP"
	case menuTOP[4]:
		m.jobsList[5] = "SETTINGS"
	}
}

func ShowMenuList(header string, headerIP string, cs *configSettings) {
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuMainColor))
	titleStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(menuMainColor))

	const defaultWidth = 90

	// Initialize the list with empty items; items will be set in updateListItems
	l := list.New([]list.Item{}, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Main Menu"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.KeyMap.ShowFullHelp = key.NewBinding() // remove '?' help option

	s := spinner.New()
	s.Spinner = spinner.Pulse

	m := MenuList{
		list:           l,
		header:         header,
		headerIP:       headerIP,
		state:          StateMainMenu,
		spinner:        s,
		spinnerMsg:     "Action Performing",
		configSettings: cs,
	}

	m.updateListItems()

	m.list.KeyMap.Quit = key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	)

	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
