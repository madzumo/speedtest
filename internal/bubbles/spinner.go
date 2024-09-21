package bubbles

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error
type quitMsg struct{} // Define a struct for our quit message

type SpinModel struct {
	spinner  spinner.Model
	quitting bool
	err      error
	// Message to show
	spinMsg string
	// ANSI 256 colors (8-bit) for the spinner
	lipColor string
}

func newSpinModel(spinMsg string, lipColor string) SpinModel {
	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(lipColor))

	return SpinModel{spinner: s, spinMsg: spinMsg, lipColor: lipColor}
}

func (m SpinModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m SpinModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}
	case quitMsg:
		m.quitting = true // Handle our custom quit message
		return m, tea.Quit
	case errMsg:
		m.err = msg
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m SpinModel) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	str := fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.spinMsg)
	if m.quitting {
		return str + "\n"
	}
	return str
}

func ShowSpinner(quit chan struct{}, spinMessage string, lipColor string) {
	p := tea.NewProgram(newSpinModel(spinMessage, lipColor))
	go func() {
		<-quit            // Wait for a message on the quit channel
		p.Send(quitMsg{}) // Send a quitMsg to the program
	}()
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
