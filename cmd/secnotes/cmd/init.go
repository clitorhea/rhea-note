package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/clitorhea/rhea-note/pkg/auth"
	"github.com/clitorhea/rhea-note/pkg/config"
	"github.com/clitorhea/rhea-note/pkg/crypto"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Aliases: []string{"i"},
	Short: "Initialize secnotes configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("Enter Master Password: ")
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return err
		}

		fmt.Print("Confirm Master Password: ")
		confirm, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return err
		}

		if string(password) != string(confirm) {
			return fmt.Errorf("passwords do not match")
		}

		salt, err := crypto.GenerateSalt()
		if err != nil {
			return err
		}

		cfg := &config.Config{
			Salt:      salt,
			StoreDir:  filepath.Join(config.ConfigDir(), "store"),
			ServerURL: "http://localhost:8080",
			Token:     "secret-token",
		}

		if err := config.SaveConfig(cfg); err != nil {
			return err
		}
		
		// Attempt to save to OS keyring
		if err := auth.SavePassword(string(password)); err != nil {
			fmt.Printf("Warning: Could not save password to OS Keyring: %v\n", err)
		} else {
			fmt.Println("Master password securely saved to OS Keyring.")
		}

		fmt.Println("Configuration successfully initialized.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
