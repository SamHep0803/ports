package ssh

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/samhep0803/ports/internal/app"
)

type SSHRunner struct{}

func NewSSHRunner() *SSHRunner { return &SSHRunner{} }

func (r *SSHRunner) Start(p app.Profile, onLog func(string)) (*exec.Cmd, error) {
	args := []string{
		"-N",
		"-T",
		"-o", "ExitOnForwardFailure=yes",
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
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	go streamLogs(stdout, "OUT", onLog)
	go streamLogs(stderr, "ERR", onLog)

	if err := waitForStartup(p, cmd.Process.Pid, 5*time.Second); err != nil {
		return nil, err
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

func streamLogs(r io.Reader, prefix string, onLog func(string)) {
	if onLog == nil {
		return
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		onLog(prefix + ": " + scanner.Text())
	}
}

func waitForStartup(p app.Profile, pid int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !isRunning(pid) {
			return fmt.Errorf("ssh exited during startup")
		}

		if forwardsListening(p) {
			return nil
		}

		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("ssh did not become ready within %s", timeout)
}

func forwardsListening(p app.Profile) bool {
	if len(p.Forwards) == 0 {
		return true
	}
	for _, f := range p.Forwards {
		bind := f.Bind
		if bind == "" || bind == "0.0.0.0" || bind == "::" {
			bind = "127.0.0.1"
		}

		addr := net.JoinHostPort(bind, strconv.Itoa(f.LocalPort))
		conn, err := net.DialTimeout("tcp", addr, 120*time.Millisecond)
		if err != nil {
			return false
		}
		_ = conn.Close()
	}

	return true
}
