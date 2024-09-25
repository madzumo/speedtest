package bubbles

// import (
// 	"fmt"
// 	"io"
// 	"os"
// 	"strings"

// 	"github.com/atotto/clipboard"
// 	"github.com/charmbracelet/bubbles/key"
// 	"github.com/charmbracelet/bubbles/list"
// 	tea "github.com/charmbracelet/bubbletea"
// 	"github.com/charmbracelet/lipgloss"
// )

// const listHeight = 14

// var (
// 	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("111"))
// 	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
// 	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
// 	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
// 	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
// )

// type item string

// func (i item) FilterValue() string { return "" }

// type itemDelegate struct{}

// func (d itemDelegate) Height() int                             { return 1 }
// func (d itemDelegate) Spacing() int                            { return 0 }
// func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
// func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
// 	i, ok := listItem.(item)
// 	if !ok {
// 		return
// 	}

// 	str := fmt.Sprintf("%d. %s", index+1, i)

// 	fn := itemStyle.Render
// 	if index == m.Index() {
// 		fn = func(s ...string) string {
// 			return selectedItemStyle.Render("> " + strings.Join(s, " "))
// 		}
// 	}

// 	fmt.Fprint(w, fn(str))
// }

// type MenuList struct {
// 	list     list.Model
// 	choice   string
// 	quitting bool
// 	header   string
// 	headerIP string
// }

// func (m MenuList) Init() tea.Cmd {
// 	return nil
// }

// func (m MenuList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	// case tea.WindowSizeMsg:
// 	// 	m.list.SetWidth(msg.Width)
// 	// 	return m, nil

// 	// Handle mouse events using MouseAction and MouseButton
// 	case tea.MouseMsg:
// 		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
// 			// Copy headerIP to clipboard when a left mouse button click occurs
// 			err := clipboard.WriteAll(m.headerIP)
// 			if err != nil {
// 				fmt.Println("Failed to copy to clipboard:", err)
// 			}
// 		}
// 		return m, nil

// 	case tea.KeyMsg:
// 		switch keypress := msg.String(); keypress {
// 		case "q", "ctrl+c":
// 			m.quitting = true
// 			return m, tea.Quit

// 		case "enter":
// 			i, ok := m.list.SelectedItem().(item)
// 			if ok {
// 				m.choice = string(i)
// 			}
// 			m.quitting = true
// 			return m, tea.Quit
// 		}
// 	}

// 	var cmd tea.Cmd
// 	m.list, cmd = m.list.Update(msg)
// 	return m, cmd
// }

// func (m MenuList) View() string {
// 	return m.header + "\n" + m.list.View()
// }

// func ShowMenuList(menuTitle string, showtitle bool, menuItems []string, selectColor string, header string, headerIP string) string {
// 	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color(selectColor))
// 	titleStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color(selectColor))
// 	items := []list.Item{}
// 	for _, value := range menuItems {
// 		items = append(items, item(value))
// 	}

// 	const defaultWidth = 20

// 	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
// 	l.Title = menuTitle
// 	l.SetShowStatusBar(false)
// 	l.SetFilteringEnabled(false)
// 	l.SetShowTitle(showtitle)
// 	l.Styles.Title = titleStyle
// 	l.Styles.PaginationStyle = paginationStyle
// 	l.Styles.HelpStyle = helpStyle
// 	l.KeyMap.ShowFullHelp = key.NewBinding() //remove ? more

// 	m := MenuList{list: l, header: header, headerIP: headerIP}
// 	m.list.KeyMap.Quit = key.NewBinding(
// 		key.WithKeys("esc", "ctrl+c"), //you can add q to escape here
// 		key.WithHelp("esc", "quit"),
// 	)

// 	// finalM, err := tea.NewProgram(m).Run() //6:36pm
// 	finalM, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
// 	if err != nil {
// 		fmt.Println("Error running program:", err)
// 		os.Exit(1)
// 	}
// 	menuModel, _ := finalM.(MenuList)
// 	return menuModel.choice
// }
