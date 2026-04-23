package ui

import (
	"fmt"
	"strings"

	"github.com/alvestass/admin-cli/internal/config"
	"github.com/alvestass/admin-cli/internal/trailbase"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type setupStep int

const (
	setupStepURL      setupStep = 0
	setupStepEmail    setupStep = 1
	setupStepPassword setupStep = 2
	setupStepAuthing  setupStep = 3
	setupStepDone     setupStep = 4
)

type authDoneMsg struct {
	tokens *trailbase.Tokens
	err    error
}

type setupModel struct {
	step    setupStep
	inputs  [3]textinput.Model
	spinner spinner.Model
	tokens  *trailbase.Tokens
	errMsg  string
}

func newSetupModel(existingURL string) setupModel {
	urlIn := textinput.New()
	urlIn.Placeholder = "https://alvestass-trailbase.fly.dev"
	if existingURL != "" {
		urlIn.SetValue(existingURL)
	}
	urlIn.Focus()

	emailIn := textinput.New()
	emailIn.Placeholder = "admin@localhost"

	passIn := textinput.New()
	passIn.Placeholder = "lösenord"
	passIn.EchoMode = textinput.EchoPassword
	passIn.EchoCharacter = '•'

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return setupModel{
		inputs:  [3]textinput.Model{urlIn, emailIn, passIn},
		spinner: sp,
	}
}

func (m setupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if msg.Type == tea.KeyEnter && m.step < setupStepAuthing {
			val := strings.TrimSpace(m.inputs[m.step].Value())
			if val == "" {
				if placeholder := m.inputs[m.step].Placeholder; placeholder != "" {
					m.inputs[m.step].SetValue(placeholder)
					val = placeholder
				} else {
					return m, nil
				}
			}
			m.inputs[m.step].Blur()
			next := m.step + 1
			if next < setupStepAuthing {
				m.inputs[next].Focus()
				m.step = next
				return m, textinput.Blink
			}
			// All inputs filled — authenticate.
			m.step = setupStepAuthing
			url := strings.TrimSpace(m.inputs[0].Value())
			email := strings.TrimSpace(m.inputs[1].Value())
			pass := m.inputs[2].Value()
			return m, tea.Batch(m.spinner.Tick, authenticate(url, email, pass))
		}

	case spinner.TickMsg:
		if m.step == setupStepAuthing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case authDoneMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("Autentisering misslyckades: %v\nFörsök igen.", msg.err)
			m.step = setupStepPassword
			m.inputs[2].SetValue("")
			m.inputs[2].Focus()
			return m, textinput.Blink
		}
		m.tokens = msg.tokens
		m.step = setupStepDone
		return m, tea.Quit
	}

	if m.step < setupStepAuthing {
		var cmd tea.Cmd
		m.inputs[m.step], cmd = m.inputs[m.step].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m setupModel) View() string {
	var b strings.Builder

	header := lipgloss.NewStyle().Bold(true).Render("Alvesta Simsällskap — Admin CLI")
	b.WriteString(header + "\n")
	b.WriteString("Ange anslutningsuppgifter för att fortsätta.\n\n")

	labels := []string{"Backend-URL:", "E-postadress:", "Lösenord:"}
	for i, label := range labels {
		active := m.step == setupStep(i)
		lStyle := lipgloss.NewStyle()
		if active {
			lStyle = lStyle.Bold(true)
		} else {
			lStyle = lStyle.Faint(true)
		}
		b.WriteString(lStyle.Render(label) + "\n")
		b.WriteString(m.inputs[i].View() + "\n\n")
	}

	switch m.step {
	case setupStepAuthing:
		b.WriteString(m.spinner.View() + " Ansluter...\n")
	case setupStepDone:
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓ Inloggning lyckades. Konfiguration sparad.") + "\n")
	}

	if m.errMsg != "" {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.errMsg) + "\n")
	}

	b.WriteString(lipgloss.NewStyle().Faint(true).Render("\nTryck Ctrl+C för att avbryta"))
	return b.String()
}

func authenticate(url, email, password string) tea.Cmd {
	return func() tea.Msg {
		client, err := trailbase.NewClient(url)
		if err != nil {
			return authDoneMsg{err: fmt.Errorf("ogiltig URL: %w", err)}
		}
		if err := client.Login(email, password); err != nil {
			return authDoneMsg{err: err}
		}
		return authDoneMsg{tokens: client.Tokens()}
	}
}

// RunSetup presents the first-run or re-authentication wizard and returns an
// updated Config populated with the backend URL and fresh session tokens.
func RunSetup(cfg *config.Config) (*config.Config, error) {
	m := newSetupModel(cfg.BackendURL)
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return cfg, fmt.Errorf("inloggningsguide misslyckades: %w", err)
	}
	final := result.(setupModel)
	if final.step != setupStepDone {
		return cfg, fmt.Errorf("inloggning avbruten")
	}
	tokens := final.tokens
	cfg.BackendURL = strings.TrimSpace(final.inputs[0].Value())
	cfg.AuthToken = tokens.AuthToken
	cfg.RefreshToken = tokens.RefreshToken
	cfg.CsrfToken = tokens.CsrfToken
	return cfg, nil
}
