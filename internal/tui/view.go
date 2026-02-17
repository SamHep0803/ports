package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if len(m.profiles) == 0 {
		return "No profiles found.\nPress q to quit.\n"
	}

	if m.width == 0 || m.height == 0 {
		return m.basicView()
	}

	leftW := max(26, m.width/3)
	rightW := max(26, m.width-leftW-4)

	h := max(10, m.height-2)

	left := paneStyle.Width(leftW).Height(h).Render(m.viewLeft())
	right := paneStyle.Width(rightW).Height(h).Render(m.viewRight())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)

}

func (m Model) basicView() string {
	var out strings.Builder
	out.WriteString("ports - SSH Port Forward Manager\n\n")

	for i, p := range m.profiles {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		out.WriteString(cursor + " " + p.Name + "\n")
	}

	out.WriteString("\n j/k move - ret/s toggle tunnel - q quit\n")
	return out.String()
}

func (m Model) viewLeft() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Profiles") + "\n\n")

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

func (m Model) viewRight() string {
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

	b.WriteString("\n" + titleStyle.Render("Forwards") + "\n")

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

	// Reuse your existing message fields (if you have them)
	if m.info != "" {
		b.WriteString("\n" + m.info + "\n")
	}
	if m.err != "" {
		b.WriteString("\nERROR: " + m.err + "\n")
	}

	return b.String()
}
