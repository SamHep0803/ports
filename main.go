package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Profiles []Profile `yaml:"profiles"`
}

type Profile struct {
	Name     string    `yaml:"name"`
	Host     string    `yaml:"host"`
	User     string    `yaml:"user"`
	KeyPath  string    `yaml:"keyPath"`
	Forwards []Forward `yaml:"forwards"`
}

type Forward struct {
	Bind       string `yaml:"bind"`
	LocalPort  int    `yaml:"localPort"`
	RemoteHost string `yaml:"remoteHost"`
	RemotePort int    `yaml:"remotePort"`
}

type Manager struct {
	mu      sync.Mutex
	running map[string]*exec.Cmd
}

func NewManager() *Manager {
	return &Manager{running: make(map[string]*exec.Cmd)}
}

func (m *Manager) Start(p Profile) (int, error) {
	m.mu.Lock()
	if _, ok := m.running[p.Name]; ok {
		m.mu.Unlock()
		return 0, fmt.Errorf("already running")
	}
	m.mu.Unlock()

	cmd, err := startSSH(p)
	if err != nil {
		fatal("start SSH", err)
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

	return stopCmd(cmd)
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
		_ = stopCmd(cmd)
	}
}

func stopCmd(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	pid := cmd.Process.Pid

	if err := syscall.Kill(-pid, syscall.SIGTERM); err == nil {
		return nil
	}
	return cmd.Process.Signal(syscall.SIGTERM)
}

func main() {
	configPath := "config.yaml"
	cfg, err := loadConfig(configPath)
	if err != nil {
		fatal("load config", err)
	}

	mgr := NewManager()
	defer mgr.StopAll()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		fmt.Println("shutting down - stopping tunnels...")
		mgr.StopAll()
		os.Exit(0)
	}()

	if len(os.Args) < 3 {
		fmt.Println("check args")
		os.Exit(2)
	}

	p, ok := findProfile(cfg, os.Args[2])
	if !ok {
		fatal("start", fmt.Errorf("unknown profile"))
	}
	pid, err := mgr.Start(*p)
	if err != nil {
		fatal("start", err)
	}
	fmt.Printf("started %s (pid %d)", p.Name, pid)

	select {}
}

func fatal(op string, err error) {
	fmt.Fprintf(os.Stderr, "error (%s): %v\n", op, err)
	os.Exit(1)
}

func loadConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func findProfile(cfg Config, name string) (*Profile, bool) {
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == name {
			return &cfg.Profiles[i], true
		}
	}

	return nil, false
}

func startSSH(p Profile) (*exec.Cmd, error) {
	args := []string{
		"-N",
		"-T",
		"-i", expandHome(p.KeyPath),
	}

	for _, f := range p.Forwards {
		bind := f.Bind
		if bind == "" {
			bind = "127.0.0.1"
		}
		spec := fmt.Sprintf("%s:%d:%s:%d", bind, f.LocalPort, f.RemoteHost, f.RemotePort)
		args = append(args, "-L", spec)
	}

	target := fmt.Sprintf("%s@%s", p.User, p.Host)
	args = append(args, target)

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	time.Sleep(150 * time.Millisecond)
	if !isRunning(cmd.Process.Pid) {
		return nil, fmt.Errorf("ssh exited immediately (check output above)")
	}

	return cmd, nil
}

func expandHome(p string) string {
	if p == "" {
		return p
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~/"))
		}
	}
	return p
}

func isRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	err := syscall.Kill(pid, 0)
	return err == nil
}
