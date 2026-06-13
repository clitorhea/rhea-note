package cmd

import (
	"fmt"

	"github.com/clitorhea/rhea-note/pkg/config"
	"github.com/clitorhea/rhea-note/pkg/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Aliases: []string{"s"},
	Short: "Synchronize local notes with the remote server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("could not load config, run 'secnotes init' first: %v", err)
		}

		client := sync.NewClient(cfg.ServerURL, cfg.Token)
		fmt.Printf("Synchronizing with %s...\n", cfg.ServerURL)
		
		if err := client.Synchronize(cfg.StoreDir); err != nil {
			return fmt.Errorf("sync failed: %v", err)
		}

		fmt.Println("Synchronization complete.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
