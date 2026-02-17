package app

import (
	"errors"
	"fmt"
	"os/exec"
	"sync"
)

type Starter interface {
	Start(profile Profile) (*exec.Cmd, error)
	Stop(cmd *exec.Cmd) error
}

type Manager struct {
	mu      sync.Mutex
	running map[string]*exec.Cmd
	ssh     Starter
}

func NewManager(starter Starter) *Manager {
	return &Manager{
		running: make(map[string]*exec.Cmd),
		ssh:     starter,
	}
}

func (m *Manager) Start(p Profile) (int, error) {
	m.mu.Lock()
	if _, ok := m.running[p.Name]; ok {
		m.mu.Unlock()
		return 0, fmt.Errorf("already running")
	}
	m.mu.Unlock()

	cmd, err := m.ssh.Start(p)
	if err != nil {
		return 0, err
	}

	m.mu.Lock()
	m.running[p.Name] = cmd
	m.mu.Unlock()

	go func() { _ = cmd.Wait() }()

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
