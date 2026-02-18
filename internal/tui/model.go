package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/samhep0803/ports/internal/app"
)

type Model struct {
	profiles []app.Profile
	cursor   int
	mgr      *app.Manager

	info string
	err  string

	width  int
	height int
}

func New(profiles []app.Profile, mgr *app.Manager) Model {
	return Model{
		profiles: profiles,
		mgr:      mgr,
	}
}

func (m Model) Init() tea.Cmd {
	return refreshCmd()
}
