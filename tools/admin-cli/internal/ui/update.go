package ui

import (
	"fmt"
	"strings"

	"github.com/alvestass/admin-cli/internal/trailbase"
	"github.com/alvestass/admin-cli/internal/validate"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type updatePhase int

const (
	updatePhaseLoading  updatePhase = iota
	updatePhaseSelect              // choose which field to edit
	updatePhaseEdit                // editing a single field
	updatePhaseConfirm             // show diff, ask y/n
	updatePhaseSaving
	updatePhaseDone
	updatePhaseCancelled
)

// fieldDef describes one editable field.
type fieldDef struct {
	key   string // JSON/API key
	label string // Swedish display label
}

var clubInfoFields = []fieldDef{
	{"name", "Namn"},
	{"tagline", "Slogan"},
	{"founding_year", "Grundår"},
	{"short_description", "Kort beskrivning"},
	{"address", "Adress"},
	{"city", "Stad"},
	{"postal_code", "Postnummer"},
	{"phone", "Telefon"},
	{"email", "E-post"},
}

type fetchedMsg struct {
	info *trailbase.ClubInfo
	err  error
}

type savedMsg struct{ err error }

type updateModel struct {
	phase      updatePhase
	spinner    spinner.Model
	client     *trailbase.Client
	info       *trailbase.ClubInfo
	cursor     int
	input      textinput.Model
	editField  int
	changes    map[string]string // key → new value
	issues     []validate.CheckIssue
	errMsg     string
}

func newUpdateModel(client *trailbase.Client) updateModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	in := textinput.New()
	in.CharLimit = 400

	return updateModel{
		phase:   updatePhaseLoading,
		spinner: sp,
		client:  client,
		input:   in,
		changes: make(map[string]string),
	}
}

func (m updateModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchClubInfo(m.client))
}

func (m updateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.phase = updatePhaseCancelled
			return m, tea.Quit
		}
		switch m.phase {
		case updatePhaseSelect:
			return m.handleSelect(msg)
		case updatePhaseEdit:
			return m.handleEdit(msg)
		case updatePhaseConfirm:
			return m.handleConfirm(msg)
		}

	case spinner.TickMsg:
		if m.phase == updatePhaseLoading || m.phase == updatePhaseSaving {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case fetchedMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("Kunde inte hämta kontaktuppgifter: %v", msg.err)
			m.phase = updatePhaseCancelled
			return m, tea.Quit
		}
		m.info = msg.info
		m.phase = updatePhaseSelect

	case savedMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("Kunde inte spara: %v", msg.err)
			m.phase = updatePhaseSelect
			return m, nil
		}
		m.phase = updatePhaseDone
		return m, tea.Quit
	}
	return m, nil
}

func (m updateModel) handleSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(clubInfoFields) {
			m.cursor++
		}
	case "enter", " ":
		if m.cursor == len(clubInfoFields) {
			// "Done" row selected.
			if len(m.changes) == 0 {
				m.phase = updatePhaseCancelled
				return m, tea.Quit
			}
			m.phase = updatePhaseConfirm
			return m, nil
		}
		m.editField = m.cursor
		current := currentValue(m.info, clubInfoFields[m.cursor].key)
		m.input.SetValue(current)
		m.input.Focus()
		m.phase = updatePhaseEdit
		return m, textinput.Blink
	case "q":
		m.phase = updatePhaseCancelled
		return m, tea.Quit
	}
	return m, nil
}

func (m updateModel) handleEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		newVal := strings.TrimSpace(m.input.Value())
		m.input.Blur()
		original := currentValue(m.info, clubInfoFields[m.editField].key)
		if newVal != original {
			m.changes[clubInfoFields[m.editField].key] = newVal
		} else {
			delete(m.changes, clubInfoFields[m.editField].key)
		}
		m.phase = updatePhaseSelect
		return m, nil
	case tea.KeyEsc:
		m.input.Blur()
		m.phase = updatePhaseSelect
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m updateModel) handleConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch strings.ToLower(msg.String()) {
	case "j", "y":
		merged := buildPatch(m.info, m.changes)
		issues := validate.ValidateClubInfo(merged)
		if len(issues) > 0 {
			m.issues = issues
			m.phase = updatePhaseSelect
			return m, nil
		}
		m.phase = updatePhaseSaving
		return m, tea.Batch(m.spinner.Tick, saveClubInfo(m.client, *merged))
	case "n", "q":
		m.phase = updatePhaseCancelled
		return m, tea.Quit
	}
	return m, nil
}

func (m updateModel) View() string {
	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Render("Uppdatera kontaktuppgifter")
	b.WriteString(title + "\n\n")

	switch m.phase {
	case updatePhaseLoading:
		b.WriteString(m.spinner.View() + " Hämtar uppgifter...\n")

	case updatePhaseSelect:
		if len(m.issues) > 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Valideringsfel:") + "\n")
			for _, iss := range m.issues {
				b.WriteString(fmt.Sprintf("  [%s] %s — %s\n", iss.Field, iss.Value, iss.Rule))
			}
			b.WriteString("\n")
		}
		cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
		for i, f := range clubInfoFields {
			cursor := "  "
			if i == m.cursor {
				cursor = cursorStyle.Render("> ")
			}
			current := currentValue(m.info, f.key)
			if v, changed := m.changes[f.key]; changed {
				current = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(v + " (ändrad)")
			}
			b.WriteString(fmt.Sprintf("%s%-20s %s\n", cursor, f.label+":", current))
		}
		doneLabel := "  Klar"
		if m.cursor == len(clubInfoFields) {
			doneLabel = cursorStyle.Render("> ") + "Klar"
		}
		b.WriteString(doneLabel + "\n")
		b.WriteString(lipgloss.NewStyle().Faint(true).Render("\n↑/↓ välj fält  •  Enter redigera/klar  •  q avbryt"))

	case updatePhaseEdit:
		f := clubInfoFields[m.editField]
		b.WriteString(fmt.Sprintf("Redigera %s (Enter sparar, Esc avbryter):\n", f.label))
		b.WriteString(m.input.View() + "\n")

	case updatePhaseConfirm:
		b.WriteString("Ändringar att spara:\n\n")
		for k, newVal := range m.changes {
			old := currentValue(m.info, k)
			label := keyLabel(k)
			b.WriteString(fmt.Sprintf("  %s: %q → %q\n", label, old, newVal))
		}
		b.WriteString("\nSpara ändringar? (j/n) ")

	case updatePhaseSaving:
		b.WriteString(m.spinner.View() + " Sparar...\n")

	case updatePhaseDone:
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓ Kontaktuppgifter uppdaterade.") + "\n")

	case updatePhaseCancelled:
		if m.errMsg != "" {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.errMsg) + "\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Faint(true).Render("Uppdatering avbruten.") + "\n")
		}
	}
	return b.String()
}

// RunUpdate runs the interactive update flow for club_info.
func RunUpdate(client *trailbase.Client) error {
	m := newUpdateModel(client)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return err
	}
	final := result.(updateModel)
	if final.errMsg != "" {
		return fmt.Errorf("%s", final.errMsg)
	}
	return nil
}

// --- helpers ---

func fetchClubInfo(client *trailbase.Client) tea.Cmd {
	return func() tea.Msg {
		info, err := client.GetClubInfo()
		return fetchedMsg{info: info, err: err}
	}
}

func saveClubInfo(client *trailbase.Client, info trailbase.ClubInfo) tea.Cmd {
	return func() tea.Msg {
		return savedMsg{err: client.UpdateClubInfo(info)}
	}
}

// currentValue extracts the string representation of a ClubInfo field by key.
func currentValue(info *trailbase.ClubInfo, key string) string {
	if info == nil {
		return ""
	}
	switch key {
	case "name":
		return info.Name
	case "tagline":
		return info.Tagline
	case "founding_year":
		return fmt.Sprintf("%d", info.FoundingYear)
	case "short_description":
		return info.ShortDescription
	case "address":
		return info.Address
	case "city":
		return info.City
	case "postal_code":
		return info.PostalCode
	case "phone":
		return info.Phone
	case "email":
		return info.Email
	}
	return ""
}

// buildPatch merges the changes map into a copy of the ClubInfo record.
func buildPatch(info *trailbase.ClubInfo, changes map[string]string) *trailbase.ClubInfo {
	if info == nil {
		return &trailbase.ClubInfo{}
	}
	merged := *info
	for k, v := range changes {
		switch k {
		case "name":
			merged.Name = v
		case "tagline":
			merged.Tagline = v
		case "founding_year":
			var yr int
			fmt.Sscanf(v, "%d", &yr)
			merged.FoundingYear = yr
		case "short_description":
			merged.ShortDescription = v
		case "address":
			merged.Address = v
		case "city":
			merged.City = v
		case "postal_code":
			merged.PostalCode = v
		case "phone":
			merged.Phone = v
		case "email":
			merged.Email = v
		}
	}
	return &merged
}

func keyLabel(key string) string {
	for _, f := range clubInfoFields {
		if f.key == key {
			return f.label
		}
	}
	return key
}
