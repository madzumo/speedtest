package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("111"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

var menuTOP = []string{
	"Run ALL Tests",
	"Run Internet Speed Tests Only",
	"Run Iperf Test Only",
	"Change Settings",
	"Save Settings",
}

var menuSettings = []string{
	"Set Iperf Server IP",
	"Set Iperf Port number",
	"Set Repeat Test Interval in Minutes",
	"Set MSS Size",
	"Toggle: Use CloudFlare",
	"Toggle: Use M-Labs",
	"Toggle: Use Speedtest.net",
	"Toggle: Show Browser on Speed Tests",
	"Back to Main Menu",
}

var menuSMTP = []string{
	"Set SMTP Server",
	"Set SMTP Port",
	"Set SMTP Username",
	"Set SMTP Password",
	"Set E-Mail Subject",
	"Set E-Mail Message",
	"Back to Settings Menu",
}

type MenuState int

const (
	StateMainMenu MenuState = iota
	StateSettingsMenu
	StateSpinner
	StateResultDisplay
	StateTextInput
)

type backgroundJobTypes int

const (
	JobCloudFlareTest backgroundJobTypes = iota
	JobMLabTest
	JobSTnetTest
	JobIpefTest
	JobSaveSettings
)

type backgroundJobMsg struct {
	result string
}

type continueJobs struct{}

type MenuList struct {
	list                list.Model
	choice              string
	header              string
	headerIP            string
	state               MenuState
	prevState           MenuState
	spinner             spinner.Model
	spinnerMsg          string
	backgroundJobResult string
	selectColor         string
	menuTitle           string
	jobList             map[int]string
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
				case "Run ALL Tests", "Run Internet Speed Tests Only", "Run Iperf Test Only", "Save Settings":
					BuildJobList(m)
					m.prevState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob())
				case "Change Settings":
					m.prevState = m.state
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

func (m *MenuList) updateSettingsMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc", "ctrl+c":
			m.state = StateMainMenu
			m.updateListItems()
			return m, nil
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				choice := string(i)
				switch choice {
				case "Back to Main Menu":
					m.state = StateMainMenu
					m.updateListItems()
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
		// case "q", "esc", "ctrl+c":
		// 	return m, tea.Quit
		default:
			// For other key presses, update the spinner
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case backgroundJobMsg:
		// if len(m.jobList) <= 0 {
		m.backgroundJobResult = msg.result
		m.state = StateResultDisplay
		return m, nil
		// } else {
		// 	return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob())
		// }
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
			m.state = m.prevState
			m.updateListItems()
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *MenuList) startBackgroundJob() tea.Cmd {
	return func() tea.Msg {
		if len(m.jobList) == 0 {
			return backgroundJobMsg{result: "All jobs completed successfully!"}
		}

		// Grab the first job in the list
		for i := range m.jobList {
			runJob := m.jobList[i]
			switch runJob {
			case "CloudFlare":
				m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
				m.spinnerMsg = "Running Cloudflare Speed test"
				time.Sleep(3 * time.Second)
			case "MLab":
				m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("57"))
				m.spinnerMsg = "Running MLab Speed test"
				time.Sleep(3 * time.Second)
			case "Iperf":
				m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("200"))
				m.spinnerMsg = "Running Iperf Speed test"
				time.Sleep(3 * time.Second)
			}

			// Remove the job after running
			delete(m.jobList, i)
			break // Exit the loop after the first job is run
		}

		// Continue running the next job if there are any left
		if len(m.jobList) > 0 {
			return continueJobs{}
		} else {
			return backgroundJobMsg{result: "Completed successfully!"}
		}
	}
}

func (m MenuList) View() string {
	switch m.state {
	case StateMainMenu, StateSettingsMenu:
		return m.header + "\n" + m.list.View()
	case StateSpinner:
		return fmt.Sprintf("\n\n   %s %s\n\n%v\nLength:%v", m.spinner.View(), m.spinnerMsg, m.jobList, len(m.jobList))
	case StateResultDisplay:
		return m.viewResultDisplay()
	default:
		return "Unknown state"
	}
}

func (m MenuList) viewResultDisplay() string {
	return fmt.Sprintf("\n\n%s\n\nPress 'esc' to return.", m.backgroundJobResult)
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
			items = append(items, item(value))
		}
		m.list.SetItems(items)
	}
	m.list.ResetSelected()
}

func BuildJobList(m *MenuList) {
	switch m.choice {
	case "Run ALL Tests":
		m.jobList = map[int]string{
			0: "CloudFlare",
			1: "MLab",
			2: "Iperf",
		}
	case "Run Internet Speed Tests Only":
		m.jobList = map[int]string{
			0: "CloudFlare",
			1: "MLab",
			2: "Iperf",
		}
	case "Run Iperf Test Only":
		m.jobList = map[int]string{
			0: "CloudFlare",
			1: "MLab",
			2: "Iperf",
		}
	case "Save Settings":
		m.jobList = map[int]string{
			0: "CloudFlare",
			1: "MLab",
			2: "Iperf",
		}
	}
}

func ShowMenuList(selectColor string, header string, headerIP string) {
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(selectColor))
	titleStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(selectColor))

	const defaultWidth = 20

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

	// Initialize the spinner
	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(selectColor))

	m := MenuList{
		list:        l,
		header:      header,
		headerIP:    headerIP,
		state:       StateMainMenu,
		spinner:     s,
		selectColor: selectColor,
		spinnerMsg:  "Perro sucio",
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
