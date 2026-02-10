package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

type State struct {
	Running map[string]int `json:"running"`
}

func main() {
	cmd := os.Args[1]
	configPath := "config.yaml"
	statePath := "state.json"

	cfg, err := loadConfig(configPath)
	if err != nil {
		fatal("load config", err)
	}

	st, err := loadState(statePath)
	if err != nil {
		fatal("load state", err)
	}

	switch cmd {
	case "start":
		name := os.Args[2]
		p, ok := findProfile(cfg, name)
		if !ok {
			fatal("start", fmt.Errorf("unknown profile"))
		}
		if pid, ok := st.Running[name]; ok && isRunning(pid) {
			fatal("start", fmt.Errorf("already running"))
		}

		pid, err := startSSH(*p)
		if err != nil {
			fatal("start ssh", err)
		}

		st.Running[name] = pid
		if err := saveState(statePath, st); err != nil {
			_ = stopPID(pid)
			fatal("save state", err)
		}

		fmt.Printf("started %s (pid %d)\n", name, pid)
	}
}

func stopPID(pid int) error {
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		return nil
	}

	return syscall.Kill(pid, syscall.SIGTERM)
}

func saveState(path string, st State) error {
	tmp := path + ".tmp"
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}

	return os.Rename(tmp, path)
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

func loadState(path string) (State, error) {
	st := State{Running: map[string]int{}}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return st, nil
		}
		return State{}, err
	}

	if err := json.Unmarshal(b, &st); err != nil {
		return State{}, err
	}

	if st.Running == nil {
		st.Running = map[string]int{}
	}

	return st, nil
}

func startSSH(p Profile) (int, error) {
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
		return 0, err
	}

	time.Sleep(150 * time.Millisecond)
	if !isRunning(cmd.Process.Pid) {
		return 0, fmt.Errorf("ssh exited immediately (check output above)")
	}

	return cmd.Process.Pid, nil
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
