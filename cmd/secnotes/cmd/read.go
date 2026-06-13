package cmd

import (
	"fmt"
	"os"

	"github.com/clitorhea/rhea-note/pkg/auth"
	"github.com/clitorhea/rhea-note/pkg/config"
	"github.com/clitorhea/rhea-note/pkg/crypto"
	"github.com/clitorhea/rhea-note/pkg/storage"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var readCmd = &cobra.Command{
	Use:   "read [note-id]",
	Short: "Decrypt and read a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		noteID := args[0]
		
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("could not load config, run 'secnotes init' first: %v", err)
		}

		password, _ := auth.GetPassword()
		if password == "" {
			fmt.Print("Enter Master Password: ")
			pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return err
			}
			password = string(pwBytes)
			_ = auth.SavePassword(password)
		}

		key := crypto.DeriveKey(password, cfg.Salt)

		ciphertext, err := storage.LoadNoteLocal(cfg.StoreDir, noteID)
		if err != nil {
			return err
		}

		plaintext, err := crypto.Decrypt(ciphertext, key)
		if err != nil {
			return fmt.Errorf("failed to decrypt (wrong password?): %v", err)
		}

		fmt.Print(string(plaintext))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
