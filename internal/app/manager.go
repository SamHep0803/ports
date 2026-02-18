package app

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

type Starter interface {
	Start(profile Profile, onLog func(string)) (*exec.Cmd, error)
	Stop(cmd *exec.Cmd) error
}

type Manager struct {
	mu      sync.Mutex
	running map[string]*exec.Cmd
	logs    map[string][]string
	ssh     Starter
}

func NewManager(starter Starter) *Manager {
	return &Manager{
		running: make(map[string]*exec.Cmd),
		logs:    make(map[string][]string),
		ssh:     starter,
	}
}

func (m *Manager) Start(p Profile) (int, error) {
	m.mu.Lock()
	if _, ok := m.running[p.Name]; ok {
		m.mu.Unlock()
		return 0, fmt.Errorf("already running")
	}
	m.logs[p.Name] = nil
	m.mu.Unlock()

	cmd, err := m.ssh.Start(p, func(line string) {
		m.appendLog(p.Name, line)
	})
	if err != nil {
		return 0, err
	}

	m.mu.Lock()
	m.running[p.Name] = cmd
	m.mu.Unlock()

	go func(name string, started *exec.Cmd) {
		_ = started.Wait()

		m.mu.Lock()
		current, ok := m.running[name]
		if ok && current == started {
			delete(m.running, name)
		}
		m.mu.Unlock()
	}(p.Name, cmd)

	return cmd.Process.Pid, nil
}

func (m *Manager) Stop(name string) error {
	m.mu.Lock()
	cmd, ok := m.running[name]
	if ok {
		delete(m.running, name)
	}
	m.mu.Unlock()

	if !ok {
		return errors.New("not running")
	}

	return m.ssh.Stop(cmd)
}

func (m *Manager) StopAll() {
	m.mu.Lock()
	cmds := make([]*exec.Cmd, 0, len(m.running))
	for _, cmd := range m.running {
		cmds = append(cmds, cmd)
	}
	m.running = make(map[string]*exec.Cmd)
	m.mu.Unlock()

	for _, cmd := range cmds {
		_ = m.ssh.Stop(cmd)
	}
}

func (m *Manager) IsRunning(name string) (pid int, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cmd, ok := m.running[name]
	if !ok || cmd == nil || cmd.Process == nil {
		return 0, false
	}

	return cmd.Process.Pid, true
}

func (m *Manager) Logs(name string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing := m.logs[name]
	out := make([]string, len(existing))
	copy(out, existing)
	return out
}

func (m *Manager) appendLog(name, line string) {
	const maxLines = 300

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	buf := append(m.logs[name], trimmed)
	if len(buf) > maxLines {
		buf = buf[len(buf)-maxLines:]
	}
	m.logs[name] = buf
}
