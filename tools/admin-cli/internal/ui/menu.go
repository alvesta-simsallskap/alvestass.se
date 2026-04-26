package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuChoice identifies a user's main-menu selection.
type MenuChoice int

const (
	MenuUpdate MenuChoice = iota + 1
	MenuCheck
	MenuHelp
	MenuImport
	MenuQuit
)

type menuModel struct {
	cursor int // 0-based
	choice MenuChoice
}

var menuItems = []string{
	"Uppdatera kontaktuppgifter",
	"Kontrollera data",
	"Hjälp",
	"Importera tidrapportpass",
	"Avsluta",
}

func (m menuModel) Init() tea.Cmd { return nil }

func (m menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.choice = MenuQuit
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(menuItems)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.choice = MenuChoice(m.cursor + 1)
			return m, tea.Quit
		case "1":
			m.choice = MenuUpdate
			return m, tea.Quit
		case "2":
			m.choice = MenuCheck
			return m, tea.Quit
		case "3":
			m.choice = MenuHelp
			return m, tea.Quit
		case "4":
			m.choice = MenuImport
			return m, tea.Quit
		case "5":
			m.choice = MenuQuit
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m menuModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 2)
	b.WriteString(titleStyle.Render("Alvesta Simsällskap — Admin CLI") + "\n\n")
	b.WriteString("Välj en åtgärd:\n\n")

	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	activeStyle := lipgloss.NewStyle().Bold(true)
	normalStyle := lipgloss.NewStyle()

	for i, item := range menuItems {
		cursor := "  "
		style := normalStyle
		if i == m.cursor {
			cursor = cursorStyle.Render("> ")
			style = activeStyle
		}
		b.WriteString(cursor + style.Render(item) + "\n")
	}

	b.WriteString(lipgloss.NewStyle().Faint(true).Render("\n↑/↓ navigera  •  Enter välj  •  q avsluta"))
	return b.String()
}

// RunMenu presents the main menu and returns the user's choice.
func RunMenu() (MenuChoice, error) {
	p := tea.NewProgram(menuModel{})
	result, err := p.Run()
	if err != nil {
		return MenuQuit, err
	}
	return result.(menuModel).choice, nil
}
