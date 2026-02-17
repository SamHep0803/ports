package tui

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

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

		case "s", "enter":
			if len(m.profiles) == 0 {
				return m, nil
			}
			p := m.profiles[m.cursor]

			if _, ok := m.mgr.IsRunning(p.Name); ok {
				return m, stopProfileCmd(m.mgr, p.Name)
			}

			return m, startProfileCmd(m.mgr, p)
		}

	case startedMsg:
		if msg.err != nil {
			m.err = "start " + msg.name + ": " + msg.err.Error()
			m.info = ""
		} else {
			m.info = "started " + msg.name + " (pid " + strconv.Itoa(msg.pid) + ")"
			m.err = ""
		}

		return m, nil

	case stoppedMsg:
		if msg.err != nil {
			m.err = "stop " + msg.name + ": " + msg.err.Error()
			m.info = ""
		} else {
			m.info = "stopped " + msg.name
			m.err = ""
		}

		return m, nil
	}

	return m, nil
}
