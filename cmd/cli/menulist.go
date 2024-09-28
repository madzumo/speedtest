package main

import (
	"fmt"
	"io"
	"os"
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
	textJobOutcomeFront = "223"
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
		{"Set Repeat Test Interval in Minutes", "Enter Repeat Test Interval in Minutes:"},
		{"Set MSS Size", "Enter Maximum Segment Size (MSS) for Iperf Test:"},
		{"Configure Email Settings", ""},
		{"Toggle CloudFlare Test", ""},
		{"Toggle M-Labs Test", ""},
		{"Toggle Speedtest.net", ""},
		{"Toggle Show Browser on Speed Tests", ""},
	}
	menuSMTP = [][]string{ //menu Text + prompt text
		{"Toggle Email Method", ""},
		{"Set SMTP Host", "Enter SMTP Host Address:"},
		{"Set SMTP Port", "Enter SMTP Port Number:"},
		{"Set Auth Username", "Enter Username to Authenticate SMTP:"},
		{"Set Auth Password", "Enter Password to Authenticate SMTP:"},
		{"Set From: Address", "Enter Sender address (From:) for sending reports:"},
		{"Set To: Address", "Enter Recipient address (To:) for sending reports:"},
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
)

type backgroundJobMsg struct {
	result string
}

type continueJobs struct {
	jobResult string
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
	iperfRepeat         int
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
		return m.updateViewResultDisplay(msg)
	case StateSMTPMenu:
		return m.updateSMTPMenu(msg)
	case StateTextInput:
		return m.updateTextInput(msg)
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			inputValue := m.textInput.Value() // User pressed enter, save the input
			m.textInputError = false

			switch m.inputPrompt {
			case menuSettings[1][1]:
				if i, err := strconv.Atoi(inputValue); err != nil { //validate port number
					m.backgroundJobResult = fmt.Sprintf("Invalid port number: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.IperfP = i
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					m.backgroundJobResult = fmt.Sprintf("Iperf Port Number set to %s.", inputValue)
				}
			case menuSettings[0][1]:
				m.backgroundJobResult = fmt.Sprintf("Iperf Server IP set to %s.", inputValue)
				m.configSettings.IperfS = inputValue
				m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)

			case menuSettings[2][1]:
				if i, err := strconv.Atoi(inputValue); err != nil {
					m.backgroundJobResult = fmt.Sprintf("Invalid entry for Minutes: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.Interval = i
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					m.backgroundJobResult = fmt.Sprintf("Repeat Test Interval set to %s minutes.", inputValue)
				}
			case menuSettings[3][1]:
				if i, err := strconv.Atoi(inputValue); err != nil {
					m.backgroundJobResult = fmt.Sprintf("Invalid entry for MSS: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.MSS = i
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					m.backgroundJobResult = fmt.Sprintf("MSS set to %s", inputValue)
				}
			case menuSMTP[1][1]:
				m.configSettings.EmailSettings.SMTPHost = inputValue
				m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
				m.backgroundJobResult = fmt.Sprintf("SMTP Host set to: %s", inputValue)

			case menuSMTP[2][1]:
				if _, err := strconv.Atoi(inputValue); err != nil { //validate port number
					m.backgroundJobResult = fmt.Sprintf("Invalid port number: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.EmailSettings.SMTPPort = inputValue
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
					m.backgroundJobResult = fmt.Sprintf("SMTP Port Number set to: %s", inputValue)
				}
			case menuSMTP[3][1]:
				m.configSettings.EmailSettings.UserName = inputValue
				m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
				m.backgroundJobResult = fmt.Sprintf("SMTP Auth Username set to: %s", inputValue)

			case menuSMTP[4][1]:
				m.configSettings.EmailSettings.PassWord = inputValue
				m.backgroundJobResult = fmt.Sprintf("SMTP Auth Password set to: %s", inputValue)

			case menuSMTP[5][1]:
				if err := checkmail.ValidateFormat(inputValue); err != nil {
					m.backgroundJobResult = fmt.Sprintf("Not a valid E-mail Address: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.EmailSettings.From = inputValue
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
					m.backgroundJobResult = fmt.Sprintf("Sender (From) address set to: %s", inputValue)
				}

			case menuSMTP[6][1]:
				if err := checkmail.ValidateFormat(inputValue); err != nil {
					m.backgroundJobResult = fmt.Sprintf("Not a valid E-mail Address: %s", inputValue)
					m.textInputError = true
				} else {
					m.configSettings.EmailSettings.To = inputValue
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
					m.backgroundJobResult = fmt.Sprintf("Recipient (To) address set to: %s", inputValue)
				}
			}
			m.prevState = m.state
			m.state = StateResultDisplay
			return m, nil
		case tea.KeyEsc:
			m.state = StateSettingsMenu
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
				case menuSettings[2][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSettings[2][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., 10"
					m.textInput.Focus()
					m.textInput.CharLimit = 5
					m.textInput.Width = 20
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSettings[3][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSettings[3][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., 1450"
					m.textInput.Focus()
					m.textInput.CharLimit = 5
					m.textInput.Width = 20
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSettings[4][0]:
					m.list.Title = "Main Menu->Settings->SMTP"
					selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(menuSMTPcolor))
					m.list.Styles.Title = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(menuSMTPcolor))
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
					m.prevState = m.state
					m.state = StateSMTPMenu
					m.updateListItems()
					return m, nil
				case menuSettings[5][0]:
					if m.configSettings.CloudFrontTest {
						m.configSettings.CloudFrontTest = false
					} else {
						m.configSettings.CloudFrontTest = true
					}
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					return m, nil
				case menuSettings[6][0]:
					if m.configSettings.MLabTest {
						m.configSettings.MLabTest = false
					} else {
						m.configSettings.MLabTest = true
					}
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					return m, nil
				case menuSettings[7][0]:
					if m.configSettings.NetTest {
						m.configSettings.NetTest = false
					} else {
						m.configSettings.NetTest = true
					}
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, true, false)
					return m, nil
				case menuSettings[8][0]:
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
		case "q", "esc":
			m.backgroundJobResult = "Job Cancelled"
			m.state = StateResultDisplay
			return m, nil
		default:
			// For other key presses, update the spinner
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case backgroundJobMsg:
		// if len(m.jobList) <= 0 {
		m.backgroundJobResult = lipgloss.NewStyle().Foreground(lipgloss.Color(textJobOutcomeFront)).Bold(true).Render(m.jobOutcome) + "\n\n" + // msg.result
			lipgloss.NewStyle().Foreground(lipgloss.Color(m.backgroundJobResult)).Render(msg.result)

		m.state = StateResultDisplay
		return m, nil
	case continueJobs:
		return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob())
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m *MenuList) updateViewResultDisplay(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					m.textInput.CharLimit = 20
					m.textInput.Width = 20
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
				case menuSMTP[3][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSMTP[3][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., Jabuticaba"
					m.textInput.Focus()
					m.textInput.CharLimit = 100
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
					m.textInput.CharLimit = 100
					m.textInput.Width = 20
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSMTP[5][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSMTP[5][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., Its_A_Me@domain.com"
					m.textInput.Focus()
					m.textInput.CharLimit = 100
					m.textInput.Width = 50
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSMTP[6][0]:
					m.prevMenuState = m.state
					m.prevState = m.state
					m.state = StateTextInput
					m.inputPrompt = menuSMTP[6][1]
					m.textInput = textinput.New()
					m.textInput.Placeholder = "e.g., Joe_Mama@domain.com"
					m.textInput.Focus()
					m.textInput.CharLimit = 100
					m.textInput.Width = 50
					m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor))
					m.textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(textInputColor))
					return m, nil
				case menuSMTP[0][0]:
					if m.configSettings.EmailSettings.UseOutlook {
						m.configSettings.EmailSettings.UseOutlook = false
					} else {
						m.configSettings.EmailSettings.UseOutlook = true
					}
					m.header, _ = showHeaderPlusConfigPlusIP(m.configSettings, false, true)
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

func (m *MenuList) startBackgroundJob() tea.Cmd {
	return func() tea.Msg {
		var continueResult string
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

				hp.InstallPlaywright()

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

				passed, result := t.IperfTest(m.configSettings.IperfS, true, m.configSettings.IperfP, m.configSettings.MSS)
				continueResult = result
				if passed {
					delete(m.jobsList, 4)
				} else {
					time.Sleep(10 * time.Second)
				}
			} else if m.jobsList[5] != "" {
				m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
				m.spinnerMsg = "Saving Settings"
				m.spinner.Tick()
				time.Sleep(1 * time.Second)
				saveConfig(m.configSettings)
				delete(m.jobsList, 5)
			}
			return continueJobs{jobResult: continueResult}
		} else {
			return backgroundJobMsg{result: "Completed successfully!"}
		}
		// if len(m.jobsList) > 0 {
		// 	return continueJobs{jobResult: continueResult}
		// } else {
		// 	return backgroundJobMsg{result: "Completed successfully!"}
		// }
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
}

func (m MenuList) viewTextInput() string {
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(textPromptColor)).Bold(true)
	return fmt.Sprintf("\n\n%s\n\n%s", promptStyle.Render(m.inputPrompt), m.textInput.View())

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
	m.iperfRepeat = 0
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
