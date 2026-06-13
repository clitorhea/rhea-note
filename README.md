# SecNotes

SecNotes is a secure, lightweight, cross-platform terminal note-taking application designed with a **Zero-Knowledge Architecture**. It ensures your data is encrypted locally before it ever touches a network, meaning your sync server only stores unreadable ciphertext. It provides a highly interactive Terminal UI (like a terminal-based Notion) and transparent Neovim integration.

## Features

- **Zero-Knowledge Architecture:** Your master password is never stored and the server never sees your plaintext.
- **Argon2id & AES-256-GCM:** Industry-standard password hashing derives your 256-bit Data Encryption Key (DEK) for authenticated encryption.
- **Notion-style Nested Pages:** Use `[[Page Name]]` syntax to seamlessly link notes. Hyper-jump between links and spawn new nested pages instantly from within the TUI.
- **Interactive Terminal UI:** Built with Bubbletea for a beautifully styled interactive note browser. Edit notes inline, create sub-pages, and traverse your note graph with hotkeys.
- **Background Sync Daemon:** A built-in CAS sync engine efficiently handles pushes/pulls automatically in the background.
- **Conflict Guard:** Safe Last-Write-Wins logic automatically spawns fork files on conflict.
- **Neovim Integration:** Read and write notes seamlessly within Neovim (`gf` to jump to links). Files are loaded straight to memory and background synced upon save.
- **Cross-Platform:** First-class compilation support for Linux and Windows.

## Installation

### Build from source

SecNotes includes a Makefile to automatically cross-compile the CLI client and the Server binary for Linux and Windows. Ensure you have Go 1.20+ installed.

```bash
# Clone the repository
git clone git@github.com:clitorhea/rhea-note.git
cd rhea-note

# Cross-compile for Linux and Windows
make build
```
The compiled binaries will be output to `bin/linux-amd64/` and `bin/windows-amd64/`.
To install on Linux, simply copy the client to your path:
```bash
sudo cp bin/linux-amd64/secnotes /usr/local/bin/
```

## Usage

### 1. Initialize the CLI
Before you can take notes, initialize the application. This will generate your cryptographic salt, configure your remote server URL, and securely link your device footprint into your OS Keyring.
```bash
secnotes init (alias: i)
```
*You will be prompted to enter a Master Password. Do not lose this!*

### 2. Interactive Terminal UI
Browse, edit, format, delete, and link your notes using the interactive TUI.
```bash
secnotes browse (alias: b)
```
**Hotkeys inside the TUI:**
- `Enter`: Open a note
- `e`: Edit the current note inline
- `f`: Auto-format JSON/Code natively within your note
- `n`: Create a new nested note instantly
- `d` or `delete`: Delete a note (prompts to delete locally or instantly wipe from remote server)
- `tab` or `l`: Jump to `[[Links]]` within the note
- `esc` or `q`: Go back to the previous note or list

### 3. CLI Read & Write
To manually write a note, pipe text to standard input:
```bash
echo "My top secret note" | secnotes write my-first-note (alias: w)
```
To read a note to standard output:
```bash
secnotes read my-first-note (alias: r)
```

### 4. Background Syncing
SecNotes comes with a background daemon to automatically keep your notes in sync with the server. Run this in the background or configure it as a systemd service:
```bash
secnotes sync-daemon
```
Alternatively, trigger a one-off manual sync:
```bash
secnotes sync (alias: s)
```

## Server Deployment

The backend sync server (`secnotes-server`) is a minimal Go application.

### Local / Self-Hosted
```bash
bin/linux-amd64/secnotes-server --port 8080 --store ./server-data --token "secret-token"
```

### Fly.io Deployment
SecNotes includes native `fly.toml` and `Dockerfile` configurations for instant deployment to Fly.io. The server is highly optimized (preventing 10MB OOM crashes, handling precise millisecond timestamps, and persistent indexing).
```bash
# Deploy instantly
fly deploy
```

## Neovim Integration

SecNotes comes with a seamless Neovim wrapper. It automatically prompts for your password via the Neovim command line and handles the memory decryption/encryption process for any file ending in `.secnote`.

To install, simply copy `nvim/plugin/secnotes.lua` to your Neovim configuration's plugin folder (`~/.config/nvim/plugin/`).
Inside a `.secnote` file, you can press `gf` (goto file) on any `[[Link]]` to instantly jump into that nested note!

## Architecture & Security

- Plaintext notes **never** touch the disk.
- When you sync, the client pulls a remote index mapping, compares local and remote modification times against a baseline `.sync_state.json`, and intelligently pushes or pulls. If a conflict occurs, a safe fork is created.
- The server possesses zero cryptographic knowledge.

---
*Built with ❤️ in Go.*
