package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/clitorhea/rhea-note/pkg/auth"
	"github.com/clitorhea/rhea-note/pkg/config"
	"github.com/clitorhea/rhea-note/pkg/crypto"
	"github.com/clitorhea/rhea-note/pkg/storage"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var writeCmd = &cobra.Command{
	Use:   "write [note-id]",
	Short: "Write and encrypt a note from STDIN",
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

		if term.IsTerminal(int(os.Stdin.Fd())) {
			fmt.Println("Type your note below. Press Ctrl+D on a new line to save and exit:")
		}

		plaintext, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		ciphertext, err := crypto.Encrypt(plaintext, key)
		if err != nil {
			return err
		}

		if err := storage.SaveNoteLocal(cfg.StoreDir, noteID, ciphertext); err != nil {
			return err
		}

		fmt.Printf("Successfully encrypted and saved note: %s\n", noteID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(writeCmd)
}
