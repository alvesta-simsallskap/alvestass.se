package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpModel struct{ done bool }

func (m helpModel) Init() tea.Cmd { return nil }

func (m helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "q", "enter", "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m helpModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true)
	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true)
	faintStyle := lipgloss.NewStyle().Faint(true)

	b.WriteString(titleStyle.Render("Hjälp — Alvesta Simsällskap Admin CLI") + "\n\n")

	b.WriteString(headerStyle.Render("[1] Uppdatera kontaktuppgifter") + "\n")
	b.WriteString("    Hämtar den aktuella klubbinformationen och visar\n")
	b.WriteString("    alla fält med nuvarande värden. Använd piltangenterna\n")
	b.WriteString("    för att välja ett fält och tryck Enter för att\n")
	b.WriteString("    redigera det. När du är klar väljer du 'Klar' för att\n")
	b.WriteString("    se dina ändringar och bekräfta. Tryck j för att spara\n")
	b.WriteString("    eller n för att avbryta.\n\n")

	b.WriteString(headerStyle.Render("[2] Kontrollera data") + "\n")
	b.WriteString("    Hämtar den aktuella posten och kör alla\n")
	b.WriteString("    valideringsregler. Eventuella problem listas med\n")
	b.WriteString("    fältnamn, nuvarande värde och en beskrivning av\n")
	b.WriteString("    vad som är fel. Om inga problem finns visas\n")
	b.WriteString("    'Inga problem hittades.'. Tryck Enter för att gå\n")
	b.WriteString("    tillbaka till menyn.\n\n")

	b.WriteString(headerStyle.Render("[4] Importera tidrapportpass") + "\n")
	b.WriteString("    Importerar träningstillfällen från en CSV-fil.\n")
	b.WriteString("    Ange sökvägen till filen när du uppmanas.\n\n")
	b.WriteString("    CSV-filen måste ha en rubrikrad med kolumnerna:\n")
	b.WriteString("    month_key, training_group, date, title, hours\n")
	b.WriteString("    Kolumnen minutes är valfri (standard: 0).\n\n")
	b.WriteString("    Tillåtna värden för training_group:\n")
	b.WriteString("    simskola, tavlingA, tavlingB, teknik, masters, vuxencrawl\n\n")
	b.WriteString("    Sammansatt nyckel (month_key + training_group + date + title)\n")
	b.WriteString("    avgör om en rad ska infogas eller uppdateras. Identiska rader\n")
	b.WriteString("    hoppas över. Du ser alltid en sammanfattning och bekräftar\n")
	b.WriteString("    med j innan något skrivs till databasen.\n\n")

	b.WriteString(headerStyle.Render("Tangentbordskortkommandon") + "\n")
	b.WriteString("    ↑ / k    Flytta markören uppåt\n")
	b.WriteString("    ↓ / j    Flytta markören nedåt\n")
	b.WriteString("    Enter    Välj / bekräfta\n")
	b.WriteString("    Esc      Avbryt aktuellt steg\n")
	b.WriteString("    q        Gå tillbaka / avsluta\n")
	b.WriteString("    Ctrl+C   Avbryt och gå till menyn\n\n")

	b.WriteString(faintStyle.Render("Tryck Enter eller q för att återgå till menyn"))
	return b.String()
}

// RunHelp displays the help screen and blocks until the user dismisses it.
func RunHelp() {
	p := tea.NewProgram(helpModel{})
	_, _ = p.Run()
}
