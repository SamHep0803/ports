package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samhep0803/ports/internal/app"
)

type Model struct {
	profiles []app.Profile
	cursor   int
	mgr      *app.Manager

	info string
	err  string
}

func New(profiles []app.Profile, mgr *app.Manager) Model {
	return Model{
		profiles: profiles,
		mgr:      mgr,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.profiles)-1 {
				m.cursor++
			}
			return m, nil
		}

	}

	return m, nil
}

func (m Model) View() string {
	if len(m.profiles) == 0 {
		return "No profiles found.\nPress q to quit.\n"
	}

	var out strings.Builder
	out.WriteString("ports - SSH Port Forward Manager\n\n")

	for i, p := range m.profiles {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		out.WriteString(cursor + " " + p.Name + "\n")
	}

	out.WriteString("\n j/k move - q quit\n")

	return out.String()
}
