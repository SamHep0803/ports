/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samhep0803/ports/internal/app"
	"github.com/samhep0803/ports/internal/ssh"
	"github.com/samhep0803/ports/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ports",
	Short: "Ports is an SSH port forward manager.",
	Long: `An SSH port forwarding manager built by
					SamHep0803 in go.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := "config.yaml"
		cfg, err := app.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		sshRunner := ssh.NewSSHRunner()
		mgr := app.NewManager(sshRunner)
		defer mgr.StopAll()

		model := tui.New(cfg.Profiles, mgr)
		t := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
		if _, err := t.Run(); err != nil {
			return fmt.Errorf("starting tui: %w", err)
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ports.yaml)")
}
