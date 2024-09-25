package bubbles

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
	"Toggle: Use CloudFront",
	"Toggle: Use M-Labs",
	"Toggle: Use Speedtest.net",
	"Toggle: Show Browser on Speed Tests",
	"Back to Main Menu",
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

type MenuState int

const (
	StateMainMenu MenuState = iota
	StateSettingsMenu
	StateSpinner
	StateResultDisplay
)

type backgroundJobMsg struct {
	result string
}

type MenuList struct {
	list                list.Model
	choice              string
	header              string
	headerIP            string
	state               MenuState
	prevState           MenuState
	spinner             spinner.Model
	backgroundJobResult string
	selectColor         string
	menuTitle           string
	showTitle           bool
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
				case "Run ALL Tests":
					m.prevState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob("Run ALL Tests"))
				case "Run Internet Speed Tests Only":
					m.prevState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob("Run Internet Speed Tests Only"))
				case "Run Iperf Test Only":
					m.prevState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob("Run Iperf Test Only"))
				case "Change Settings":
					m.prevState = m.state
					m.state = StateSettingsMenu
					m.updateListItems()
					return m, nil
				case "Save Settings":
					m.prevState = m.state
					m.state = StateSpinner
					return m, tea.Batch(m.spinner.Tick, m.startBackgroundJob("Save Settings"))
				}
			}
			return m, nil
		}
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
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		default:
			// For other key presses, update the spinner
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case backgroundJobMsg:
		// Remove this line to prevent overwriting m.prevState
		// m.prevState = m.state
		m.backgroundJobResult = msg.result
		m.state = StateResultDisplay
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m *MenuList) updateResultDisplay(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *MenuList) startBackgroundJob(jobType string) tea.Cmd {
	return func() tea.Msg {
		// Simulate a background job
		time.Sleep(3 * time.Second)
		// Return the result
		return backgroundJobMsg{result: fmt.Sprintf("%s completed successfully!", jobType)}
	}
}

func (m MenuList) View() string {
	switch m.state {
	case StateMainMenu, StateSettingsMenu:
		return m.header + "\n" + m.list.View()
	case StateSpinner:
		return fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), "Processing...")
	case StateResultDisplay:
		return m.resultDisplayView()
	default:
		return "Unknown state"
	}
}

func (m MenuList) resultDisplayView() string {
	return fmt.Sprintf("\n\n%s\n\nPress 'q' or 'esc' to return.", m.backgroundJobResult)
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

func ShowMenuList(menuTitle string, showtitle bool, selectColor string, header string, headerIP string) string {
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(selectColor))
	titleStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(selectColor))

	const defaultWidth = 20

	// Initialize the list with empty items; items will be set in updateListItems
	l := list.New([]list.Item{}, itemDelegate{}, defaultWidth, listHeight)
	l.Title = menuTitle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(showtitle)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.KeyMap.ShowFullHelp = key.NewBinding() // remove '?' help

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
		menuTitle:   menuTitle,
		showTitle:   showtitle,
	}

	m.updateListItems()

	m.list.KeyMap.Quit = key.NewBinding(
		key.WithKeys("esc", "ctrl+c"), // you can add 'q' to escape here
		key.WithHelp("esc", "quit"),
	)

	finalM, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	menuModel, _ := finalM.(MenuList)
	return menuModel.choice
}
