package ui

import (
	"fmt"
	"strings"

	"github.com/alvestass/admin-cli/internal/importer"
	"github.com/alvestass/admin-cli/internal/trailbase"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type importPhase int

const (
	importPhaseInput     importPhase = iota
	importPhaseFetching              // spinner while fetching existing sessions
	importPhaseError                 // CSV validation errors — no backend contact
	importPhaseSummary               // diff counts + j/n confirmation
	importPhaseApplying              // spinner while applying
	importPhaseDone                  // success (or empty-file) message
	importPhaseCancelled             // user cancelled or error
)

type sessionsFetchedMsg struct {
	existing []importer.ExistingSession
	err      error
}

type importAppliedMsg struct {
	result importer.ImportResult
	err    error
}

type importModel struct {
	phase     importPhase
	spinner   spinner.Model
	input     textinput.Model
	client    *trailbase.Client
	rows      []importer.SessionRow
	parseErrs []importer.ParseError
	diff      importer.ImportDiff
	result    importer.ImportResult
	errMsg    string
}

func newImportModel(client *trailbase.Client) importModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	in := textinput.New()
	in.Placeholder = "/sökväg/till/fil.csv"
	in.CharLimit = 1024
	in.Focus()

	return importModel{
		phase:   importPhaseInput,
		spinner: sp,
		input:   in,
		client:  client,
	}
}

func (m importModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m importModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.phase = importPhaseCancelled
			return m, tea.Quit
		}
		switch m.phase {
		case importPhaseInput:
			return m.handleInput(msg)
		case importPhaseSummary:
			return m.handleSummary(msg)
		case importPhaseError, importPhaseDone, importPhaseCancelled:
			if msg.Type == tea.KeyEnter || msg.String() == "q" {
				return m, tea.Quit
			}
		}

	case spinner.TickMsg:
		if m.phase == importPhaseFetching || m.phase == importPhaseApplying {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case sessionsFetchedMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("Anslutningsfel: %v", msg.err)
			m.phase = importPhaseCancelled
			return m, tea.Quit
		}
		diff := importer.ComputeDiff(m.rows, msg.existing)
		m.diff = diff
		// Empty file or fully-matching import: skip confirmation.
		if len(diff.Inserts) == 0 && len(diff.Updates) == 0 && len(diff.Skipped) == 0 {
			m.phase = importPhaseDone
			return m, tea.Quit
		}
		m.phase = importPhaseSummary
		return m, nil

	case importAppliedMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("Importfel: %v", msg.err)
			m.phase = importPhaseCancelled
			return m, tea.Quit
		}
		m.result = msg.result
		m.phase = importPhaseDone
		return m, tea.Quit
	}

	return m, nil
}

func (m importModel) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		path := strings.TrimSpace(m.input.Value())
		if path == "" {
			return m, nil
		}
		m.input.Blur()
		rows, parseErrs, err := importer.ParseCSV(path)
		if err != nil {
			m.errMsg = fmt.Sprintf("Kunde inte läsa filen: %v", err)
			m.phase = importPhaseCancelled
			return m, tea.Quit
		}
		if len(parseErrs) > 0 {
			m.parseErrs = parseErrs
			m.phase = importPhaseError
			return m, nil
		}
		m.rows = rows
		monthKeys := uniqueMonthKeys(rows)
		m.phase = importPhaseFetching
		return m, tea.Batch(m.spinner.Tick, fetchSessions(m.client, monthKeys))
	case tea.KeyEsc:
		m.phase = importPhaseCancelled
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m importModel) handleSummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch strings.ToLower(msg.String()) {
	case "j", "y":
		m.phase = importPhaseApplying
		return m, tea.Batch(m.spinner.Tick, applyImport(m.client, m.diff))
	case "n", "q":
		m.phase = importPhaseCancelled
		return m, tea.Quit
	}
	return m, nil
}

func (m importModel) View() string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Render("Importera tidrapportpass")
	b.WriteString(title + "\n\n")

	switch m.phase {
	case importPhaseInput:
		b.WriteString("Ange sökväg till CSV-filen (Enter bekräftar, Esc avbryter):\n")
		b.WriteString(m.input.View() + "\n")

	case importPhaseFetching:
		b.WriteString(m.spinner.View() + " Hämtar befintliga sessioner...\n")

	case importPhaseError:
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		b.WriteString(errStyle.Render("Fel i CSV-filen:") + "\n")
		for _, e := range m.parseErrs {
			col := ""
			if e.Column != "" {
				col = " (" + e.Column + ")"
			}
			b.WriteString(fmt.Sprintf("  Rad %d%s: %s\n", e.Line, col, e.Message))
		}
		b.WriteString("\n" + errStyle.Render("Ingen data importerades.") + "\n")
		b.WriteString(lipgloss.NewStyle().Faint(true).Render("\nEnter/q för att återgå till menyn"))

	case importPhaseSummary:
		b.WriteString("Förhandsgranskning:\n\n")
		b.WriteString(fmt.Sprintf("  Infogningar:   %d\n", len(m.diff.Inserts)))
		b.WriteString(fmt.Sprintf("  Uppdateringar: %d\n", len(m.diff.Updates)))
		b.WriteString(fmt.Sprintf("  Oförändrade:   %d\n", len(m.diff.Skipped)))
		b.WriteString("\nVill du genomföra importen? (j/n) ")

	case importPhaseApplying:
		b.WriteString(m.spinner.View() + " Importerar...\n")

	case importPhaseDone:
		greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		if m.result.Inserted == 0 && m.result.Updated == 0 && m.result.Skipped == 0 {
			b.WriteString(greenStyle.Render("0 rader hittades. Inget att importera.") + "\n")
		} else {
			b.WriteString(greenStyle.Render(fmt.Sprintf(
				"✓ Import klar: %d tillagda, %d uppdaterade, %d oförändrade.",
				m.result.Inserted, m.result.Updated, m.result.Skipped,
			)) + "\n")
		}

	case importPhaseCancelled:
		if m.errMsg != "" {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.errMsg) + "\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Faint(true).Render("Import avbruten.") + "\n")
		}
	}

	return b.String()
}

// RunImport is the entry point for the CSV import TUI flow.
func RunImport(client *trailbase.Client) error {
	m := newImportModel(client)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return err
	}
	final := result.(importModel)
	if final.errMsg != "" {
		return fmt.Errorf("%s", final.errMsg)
	}
	return nil
}

func fetchSessions(client *trailbase.Client, monthKeys []string) tea.Cmd {
	return func() tea.Msg {
		existing, err := client.ListAllSessions(monthKeys)
		return sessionsFetchedMsg{existing: existing, err: err}
	}
}

func applyImport(client *trailbase.Client, diff importer.ImportDiff) tea.Cmd {
	return func() tea.Msg {
		result, err := client.ApplyImport(diff)
		return importAppliedMsg{result: result, err: err}
	}
}

func uniqueMonthKeys(rows []importer.SessionRow) []string {
	seen := make(map[string]bool)
	var keys []string
	for _, r := range rows {
		if !seen[r.MonthKey] {
			seen[r.MonthKey] = true
			keys = append(keys, r.MonthKey)
		}
	}
	return keys
}
