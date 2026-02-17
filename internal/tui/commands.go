package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samhep0803/ports/internal/app"
)

type startedMsg struct {
	name string
	pid  int
	err  error
}

type stoppedMsg struct {
	name string
	err  error
}

func startProfileCmd(mgr *app.Manager, p app.Profile) tea.Cmd {
	return func() tea.Msg {
		pid, err := mgr.Start(p)
		return startedMsg{name: p.Name, pid: pid, err: err}
	}
}

func stopProfileCmd(mgr *app.Manager, name string) tea.Cmd {
	return func() tea.Msg {
		err := mgr.Stop(name)
		return stoppedMsg{name: name, err: err}
	}
}
