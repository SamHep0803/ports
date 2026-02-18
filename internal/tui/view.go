package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m Model) View() tea.View {
	var v tea.View

	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion

	if len(m.profiles) == 0 {
		v.SetContent("No profiles found.\nPress q to quit.\n")
		return v
	}

	h := m.height
	w := m.width

	if m.width > 100 {
		profilesW := w / 3
		detailsW := w - profilesW

		profiles := paneStyle.Width(profilesW).Height(h).Render(m.profilesPane())
		details := paneStyle.Width(detailsW).Height(h).Render(m.detailsPane())

		v.SetContent(lipgloss.JoinHorizontal(lipgloss.Top, profiles, details))
	} else {
		profilesH := h / 2
		detailsH := h - profilesH

		profiles := paneStyle.Width(w).Height(profilesH).Render(m.profilesPane())
		details := paneStyle.Width(w).Height(detailsH).Render(m.detailsPane())

		v.SetContent(lipgloss.JoinVertical(lipgloss.Right, profiles, details))
	}

	return v

}

func (m Model) profilesPane() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Profiles") + "\n\n")

	fmt.Fprintf(&b, "height: %d, width %d\n\n", m.height, m.width)

	for i, p := range m.profiles {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		status := "stopped"
		if pid, ok := m.mgr.IsRunning(p.Name); ok {
			status = "run:" + strconv.Itoa(pid)
		}

		line := fmt.Sprintf("%s %-18s %s", cursor, p.Name, status)
		if i == m.cursor {
			line = titleStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n" + faintStyle.Render("j/k move • s start/stop • q quit"))
	return b.String()
}

func (m Model) detailsPane() string {
	p := m.profiles[m.cursor]

	var b strings.Builder
	b.WriteString(titleStyle.Render("Details") + "\n\n")

	fmt.Fprintf(&b, "Name:   %s\n", p.Name)
	fmt.Fprintf(&b, "Target: %s@%s\n", p.User, p.Host)

	if pid, ok := m.mgr.IsRunning(p.Name); ok {
		fmt.Fprintf(&b, "Status: running (pid %d)\n", pid)
	} else {
		b.WriteString("Status: stopped\n")
	}

	b.WriteString("\n" + titleStyle.Render("Forwards (Local -> Remote)") + "\n")

	if len(p.Forwards) == 0 {
		b.WriteString(faintStyle.Render("No forwards configured.") + "\n")
	} else {
		for _, f := range p.Forwards {
			bind := f.Bind
			if bind == "" {
				bind = "127.0.0.1"
			}
			fmt.Fprintf(&b, "%s:%d -> %s:%d\n",
				bind, f.LocalPort, f.RemoteHost, f.RemotePort)
		}
	}

	b.WriteString("\n" + titleStyle.Render("Logs") + "\n")
	logs := m.mgr.Logs(p.Name)
	if len(logs) == 0 {
		b.WriteString(faintStyle.Render("No logs yet.") + "\n")
	} else {
		start := 0
		const showLast = 12
		if len(logs) > showLast {
			start = len(logs) - showLast
		}
		for _, line := range logs[start:] {
			b.WriteString(line + "\n")
		}
	}

	if m.info != "" {
		b.WriteString("\n" + m.info + "\n")
	}
	if m.err != "" {
		b.WriteString("\nERROR: " + m.err + "\n")
	}

	return b.String()
}
