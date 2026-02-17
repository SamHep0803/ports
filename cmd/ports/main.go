package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/samhep0803/ports/internal/app"
	"github.com/samhep0803/ports/internal/ssh"
)

func main() {
	configPath := "config.yaml"
	cfg, err := app.LoadConfig(configPath)
	if err != nil {
		fatal("load config", err)
	}

	sshRunner := ssh.NewSSHRunner()
	mgr := app.NewManager(sshRunner)
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

	p, ok := app.FindProfile(cfg, os.Args[2])
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
