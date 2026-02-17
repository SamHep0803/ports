package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/samhep0803/ports/internal/app"
)

type SSHRunner struct{}

func NewSSHRunner() *SSHRunner { return &SSHRunner{} }

func (r *SSHRunner) Start(p app.Profile) (*exec.Cmd, error) {
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

func (r *SSHRunner) Stop(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	pid := cmd.Process.Pid

	if err := syscall.Kill(-pid, syscall.SIGTERM); err == nil {
		return nil
	}
	return cmd.Process.Signal(syscall.SIGTERM)
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
