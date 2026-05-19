package ui

import (
	"fmt"
	"strings"

	"github.com/alvestass/admin-cli/internal/memberimporter"
	"github.com/alvestass/admin-cli/internal/trailbase"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type importMembersPhase int

const (
	imPhasePathIO      importMembersPhase = iota // prompt for IdrottOnline path
	imPhasePathWU                                // prompt for WeUnite path
	imPhaseParsing                               // spinner while parsing
	imPhaseParseError                            // parse/validation errors
	imPhasePreview                               // preview + j/n confirmation
	imPhaseImporting                             // spinner while writing to Trailbase
	imPhaseDone                                  // success summary
	imPhaseCancelled                             // user cancelled or fatal error
)

type importMembersModel struct {
	phase       importMembersPhase
	spinner     spinner.Model
	input       textinput.Model
	client      *trailbase.Client
	ioPath      string
	wuPath      string
	preview     memberimporter.ImportPreview
	result      memberimporter.MemberImportResult
	errMsg      string
	parseErrMsg string
}

type membersParsedMsg struct {
	preview memberimporter.ImportPreview
	data    memberimporter.ImportData
	err     error
}

type membersImportedMsg struct {
	result memberimporter.MemberImportResult
	err    error
}

func newImportMembersModel(client *trailbase.Client) importMembersModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	in := textinput.New()
	in.Placeholder = "/sökväg/till/fil.csv"
	in.CharLimit = 1024
	in.Focus()

	return importMembersModel{
		phase:   imPhasePathIO,
		spinner: sp,
		input:   in,
		client:  client,
	}
}

func (m importMembersModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m importMembersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.phase = imPhaseCancelled
			return m, tea.Quit
		}
		switch m.phase {
		case imPhasePathIO:
			return m.handlePathInput(msg, true)
		case imPhasePathWU:
			return m.handlePathInput(msg, false)
		case imPhasePreview:
			return m.handlePreview(msg)
		case imPhaseParseError, imPhaseDone, imPhaseCancelled:
			if msg.Type == tea.KeyEnter || msg.String() == "q" {
				return m, tea.Quit
			}
		}

	case spinner.TickMsg:
		if m.phase == imPhaseParsing || m.phase == imPhaseImporting {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case membersParsedMsg:
		if msg.err != nil {
			m.parseErrMsg = msg.err.Error()
			m.phase = imPhaseParseError
			return m, nil
		}
		m.preview = msg.preview
		m.phase = imPhasePreview
		return m, nil

	case membersImportedMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("Importfel: %v", msg.err)
			m.phase = imPhaseCancelled
			return m, tea.Quit
		}
		m.result = msg.result
		m.phase = imPhaseDone
		return m, tea.Quit
	}

	return m, nil
}

func (m importMembersModel) handlePathInput(msg tea.KeyMsg, isIO bool) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		path := strings.TrimSpace(m.input.Value())
		if path == "" {
			return m, nil
		}
		m.input.SetValue("")
		m.input.Focus()
		if isIO {
			m.ioPath = path
			m.phase = imPhasePathWU
			return m, nil
		}
		m.wuPath = path
		m.phase = imPhaseParsing
		return m, tea.Batch(m.spinner.Tick, parseMembers(m.client, m.ioPath, m.wuPath))
	case tea.KeyEsc:
		m.phase = imPhaseCancelled
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m importMembersModel) handlePreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch strings.ToLower(msg.String()) {
	case "j", "y":
		m.phase = imPhaseImporting
		return m, tea.Batch(m.spinner.Tick, runMembersImport(m.client, m.ioPath, m.wuPath))
	case "n", "q":
		m.phase = imPhaseCancelled
		return m, tea.Quit
	}
	return m, nil
}

func (m importMembersModel) View() string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Render("Importera memberregister")
	b.WriteString(title + "\n\n")

	switch m.phase {
	case imPhasePathIO:
		b.WriteString("Ange sökväg till IdrottOnline-exportfilen (Enter bekräftar, Esc avbryter):\n")
		b.WriteString(m.input.View() + "\n")

	case imPhasePathWU:
		b.WriteString(fmt.Sprintf("IdrottOnline: %s\n\n", m.ioPath))
		b.WriteString("Ange sökväg till WeUnite Grupplista-filen:\n")
		b.WriteString(m.input.View() + "\n")

	case imPhaseParsing:
		b.WriteString(m.spinner.View() + " Läser och analyserar filer...\n")

	case imPhaseParseError:
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		b.WriteString(errStyle.Render("Fel vid läsning av filer:") + "\n\n")
		b.WriteString(m.parseErrMsg + "\n\n")
		b.WriteString(lipgloss.NewStyle().Faint(true).Render("Ingen data importerades.\n\nEnter/q för att återgå till menyn"))

	case imPhasePreview:
		b.WriteString("Förhandsgranskning:\n\n")
		b.WriteString(fmt.Sprintf("  Simmare:             %d\n", m.preview.SwimmerCount))
		b.WriteString(fmt.Sprintf("  Styrelseledamöter:   %d\n", m.preview.BoardMemberCount))
		b.WriteString(fmt.Sprintf("  Vårdnadshavare:      %d\n", m.preview.GuardianCount))
		b.WriteString(fmt.Sprintf("  Träningsgrupper:     %d\n", m.preview.GroupCount))
		b.WriteString(fmt.Sprintf("  Familjekonst.:       %d\n", m.preview.FamilyCount))
		if m.preview.SkipCount > 0 {
			b.WriteString(fmt.Sprintf("  Hoppas över:         %d\n", m.preview.SkipCount))
		}
		b.WriteString("\nImportera till Trailbase? (j/n) ")

	case imPhaseImporting:
		b.WriteString(m.spinner.View() + " Importerar till Trailbase...\n")

	case imPhaseDone:
		green := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		b.WriteString(green.Render("✓ Import klar!") + "\n\n")
		b.WriteString(fmt.Sprintf("  Importerade poster:\n"))
		b.WriteString(fmt.Sprintf("    Medlemmar totalt:    %d\n", m.result.MembersImported))
		b.WriteString(fmt.Sprintf("      varav simmare:     %d\n", m.result.SwimmersImported))
		b.WriteString(fmt.Sprintf("      varav styrelsen:   %d\n", m.result.BoardMembersImported))
		b.WriteString(fmt.Sprintf("    Vårdnadshavare:      %d\n", m.result.GuardiansImported))
		b.WriteString(fmt.Sprintf("    Träningsgrupper:     %d\n", m.result.GroupsImported))
		b.WriteString(fmt.Sprintf("    Familjelänkar:       %d\n", m.result.FamilyLinksImported))
		if len(m.result.GroupsByCategory) > 0 {
			b.WriteString("\n  Grupper per kategori:\n")
			for cat, n := range m.result.GroupsByCategory {
				b.WriteString(fmt.Sprintf("    %-14s %d\n", cat+":", n))
			}
		}
		if len(m.result.Skipped) > 0 {
			yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
			b.WriteString(fmt.Sprintf("\n%s\n", yellow.Render(fmt.Sprintf("  Hoppade över: %d poster", len(m.result.Skipped)))))
			for _, s := range m.result.Skipped {
				b.WriteString(fmt.Sprintf("    [%s rad %d] %s\n", s.SourceFile, s.Line, s.Reason))
			}
		}
		b.WriteString(lipgloss.NewStyle().Faint(true).Render("\nEnter/q för att återgå till menyn"))

	case imPhaseCancelled:
		if m.errMsg != "" {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.errMsg) + "\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Faint(true).Render("Import avbruten.") + "\n")
		}
	}

	return b.String()
}

// RunImportMembers is the entry point for the member import TUI flow.
func RunImportMembers(client *trailbase.Client) error {
	m := newImportMembersModel(client)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return err
	}
	final := result.(importMembersModel)
	if final.errMsg != "" {
		return fmt.Errorf("%s", final.errMsg)
	}
	return nil
}

// parseMembers is a Tea command that parses both CSV files and returns preview data
// without writing anything to Trailbase.
func parseMembers(client *trailbase.Client, ioPath, wuPath string) tea.Cmd {
	return func() tea.Msg {
		rawMembers, _, err := memberimporter.ParseIdrottOnline(ioPath)
		if err != nil {
			return membersParsedMsg{err: err}
		}
		deltagare, instructorGroups, rawGuardians, _, err := memberimporter.ParseWeUnite(wuPath)
		if err != nil {
			return membersParsedMsg{err: err}
		}
		instructorEmails, err := client.ListInstructorEmails()
		if err != nil {
			return membersParsedMsg{err: fmt.Errorf("kunde inte hämta instruktörer: %w", err)}
		}
		data, err := memberimporter.BuildImportDataWithInstructors(rawMembers, deltagare, instructorGroups, rawGuardians, instructorEmails)
		if err != nil {
			return membersParsedMsg{err: err}
		}
		return membersParsedMsg{preview: data.Preview, data: data}
	}
}

// runMembersImport is a Tea command that performs the full import.
func runMembersImport(client *trailbase.Client, ioPath, wuPath string) tea.Cmd {
	return func() tea.Msg {
		result, err := memberimporter.RunImport(client, ioPath, wuPath)
		return membersImportedMsg{result: result, err: err}
	}
}
