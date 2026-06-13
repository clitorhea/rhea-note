package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/clitorhea/rhea-note/pkg/config"
	"github.com/clitorhea/rhea-note/pkg/storage"
	"github.com/clitorhea/rhea-note/pkg/tui"
	"github.com/spf13/cobra"
)

var browseCmd = &cobra.Command{
	Use:   "browse",
	Aliases: []string{"b"},
	Short: "Launch the Terminal User Interface to browse notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("could not load config, run 'secnotes init' first: %v", err)
		}

		notes, err := storage.ListNotesLocalWithTime(cfg.StoreDir)
		if err != nil {
			return err
		}

		p := tea.NewProgram(tui.InitialModel(cfg, notes), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(browseCmd)
}
