/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samhep0803/ports/internal/app"
	"github.com/samhep0803/ports/internal/ssh"
	"github.com/samhep0803/ports/internal/tui"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start <profile_name>",
	Short: "Start SSH port forwarding tunnel.",
	Long: `This command starts the SSH port forwarding
						using the specified profile name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := "config.yaml"
		cfg, err := app.LoadConfig(configPath)
		if err != nil {
			fatal("load config", err)
		}

		sshRunner := ssh.NewSSHRunner()
		mgr := app.NewManager(sshRunner)

		model := tui.New(cfg.Profiles, mgr)
		t := tea.NewProgram(model)

		if _, err := t.Run(); err != nil {
			fatal("tui", err)
		}

		defer mgr.StopAll()

		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigc
			fmt.Println("shutting down - stopping tunnels...")
			mgr.StopAll()
			os.Exit(0)
		}()

		if len(args) < 1 {
			return errors.New("You must specify a profile name.")
		}
		profileName := args[0]

		p, ok := app.FindProfile(cfg, profileName)
		if !ok {
			fatal("start", fmt.Errorf("unknown profile"))
		}
		pid, err := mgr.Start(*p)
		if err != nil {
			fatal("start", err)
		}
		fmt.Printf("started %s (pid %d)", p.Name, pid)

		select {}
	},
}

func fatal(op string, err error) {
	fmt.Fprintf(os.Stderr, "error (%s): %v\n", op, err)
	os.Exit(1)
}

func init() {
	rootCmd.AddCommand(startCmd)
}
