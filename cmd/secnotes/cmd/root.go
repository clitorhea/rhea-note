package cmd

import (
	"fmt"
	"os"

	"github.com/clitorhea/rhea-note/pkg/logger"
	"github.com/spf13/cobra"
)

var debugMode bool

var rootCmd = &cobra.Command{
	Use:   "secnotes",
	Short: "A secure terminal note-taking app",
	Long:  "SecNotes is a Zero-Knowledge note taking app that encrypts locally and syncs to a CAS backend.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return logger.Init(debugMode)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "Enable debug logging to ~/.config/secnotes/secnotes.log")
}
