package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/alvestass/admin-cli/internal/validate"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type checkPhase int

const (
	checkPhaseRunning checkPhase = iota
	checkPhaseDone
)

type checkResultsMsg struct {
	groups []checkGroup
}

type checkGroup struct {
	name   string
	issues []validate.CheckIssue
	err    error
}

type checkModel struct {
	phase    checkPhase
	spinner  spinner.Model
	checkers []validate.Checker
	groups   []checkGroup
}

func newCheckModel(checkers []validate.Checker) checkModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return checkModel{
		phase:    checkPhaseRunning,
		spinner:  sp,
		checkers: checkers,
	}
}

func (m checkModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, runCheckers(m.checkers))
}

func (m checkModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" || msg.Type == tea.KeyEnter {
			return m, tea.Quit
		}
	case spinner.TickMsg:
		if m.phase == checkPhaseRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case checkResultsMsg:
		m.groups = msg.groups
		m.phase = checkPhaseDone
	}
	return m, nil
}

func (m checkModel) View() string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Render("Kontrollera data")
	b.WriteString(title + "\n\n")

	switch m.phase {
	case checkPhaseRunning:
		b.WriteString(m.spinner.View() + " Kör kontroller...\n")

	case checkPhaseDone:
		totalIssues := 0
		for _, g := range m.groups {
			totalIssues += len(g.issues)
			if g.err != nil {
				errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
				b.WriteString(errStyle.Render(fmt.Sprintf("[%s] kunde inte hämta data — %v", g.name, g.err)) + "\n")
				continue
			}
			if len(g.issues) == 0 {
				continue
			}
			groupStyle := lipgloss.NewStyle().Bold(true).Underline(true)
			b.WriteString(groupStyle.Render(g.name) + "\n")
			for _, iss := range g.issues {
				b.WriteString(fmt.Sprintf("  [%s] %q — %s\n", iss.Field, iss.Value, iss.Rule))
			}
			b.WriteString("\n")
		}
		if totalIssues == 0 {
			ok := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓ Inga problem hittades.")
			b.WriteString(ok + "\n")
		} else {
			summary := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).
				Render(fmt.Sprintf("%d problem hittades.", totalIssues))
			b.WriteString(summary + "\n")
		}
		b.WriteString(lipgloss.NewStyle().Faint(true).Render("\nTryck Enter för att återgå till menyn"))
	}
	return b.String()
}

// RunCheck runs all registered checkers and displays grouped results.
func RunCheck(checkers []validate.Checker) error {
	m := newCheckModel(checkers)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

func runCheckers(checkers []validate.Checker) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		groups := make([]checkGroup, 0, len(checkers))
		for _, c := range checkers {
			issues, err := c.Run(ctx)
			groups = append(groups, checkGroup{
				name:   c.Name(),
				issues: issues,
				err:    err,
			})
		}
		return checkResultsMsg{groups: groups}
	}
}
