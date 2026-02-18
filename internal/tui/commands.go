package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
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

type refreshMsg struct{}

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

func refreshCmd() tea.Cmd {
	return tea.Tick(250*time.Millisecond, func(time.Time) tea.Msg {
		return refreshMsg{}
	})
}
