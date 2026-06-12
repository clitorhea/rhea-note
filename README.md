# SecNotes

SecNotes is a secure, lightweight, cross-platform terminal note-taking application designed with a **Zero-Knowledge Architecture**. It ensures your data is encrypted locally before it ever touches a network, meaning your sync server only stores unreadable ciphertext. It provides a standalone Terminal UI and transparent Neovim integration.

## Features

- **Zero-Knowledge Architecture:** Your master password is never stored and the server never sees your plaintext.
- **Argon2id Key Derivation:** Industry-standard password hashing derives your 256-bit Data Encryption Key (DEK).
- **AES-256-GCM Encryption:** Authenticated encryption ensures confidentiality and integrity.
- **Content-Addressable Sync:** A built-in CAS sync engine efficiently handles pushes/pulls.
- **Conflict Guard:** Safe Last-Write-Wins logic automatically spawns fork files on conflict.
- **Terminal User Interface (TUI):** Built with Bubbletea for a beautifully styled interactive note browser.
- **Neovim Integration:** Read and write notes seamlessly within Neovim. Files are loaded straight to memory and background synced upon save.

## Installation

### Build from source

Ensure you have Go 1.20+ installed.

```bash
# Clone the repository
git clone git@github.com:clitorhea/rhea-note.git
cd rhea-note

# Build the CLI
go build -o secnotes ./cmd/secnotes

# Build the Sync Server
go build -o secnotes-server ./cmd/server
```

## Usage

### 1. Initialize the CLI
Before you can take notes, initialize the application. This will generate your cryptographic salt and store your local config.
```bash
./secnotes init
```
*You will be prompted to enter a Master Password. Do not lose this!*

### 2. Terminal UI (TUI)
Browse your notes using the interactive interface.
```bash
./secnotes browse
```

### 3. Read & Write
To manually write a note, pipe text to standard input:
```bash
echo "My top secret note" | ./secnotes write my-first-note
```

To read a note:
```bash
./secnotes read my-first-note
```

### 4. Background Sync Server
Deploy the server on any machine.
```bash
./secnotes-server --port 8080 --store ./server-data --token "secret-token"
```

To manually push your local notes to the server:
```bash
./secnotes sync
```

## Neovim Integration

SecNotes comes with a seamless Neovim wrapper. It automatically prompts for your password via the Neovim command line and handles the memory decryption/encryption process for any file ending in `.secnote`.

To install, simply add `nvim/plugin/secnotes.lua` to your Neovim configuration's plugin folder.

## Architecture & Security

- Plaintext notes **never** touch the disk.
- Encryption uses AES-256-GCM.
- When you sync, the client pulls an index mapping, compares local and remote modification times against a baseline `.sync_state.json`, and intelligently pushes or pulls. If a conflict occurs, a safe fork is created.

---
*Built with ❤️ in Go.*
