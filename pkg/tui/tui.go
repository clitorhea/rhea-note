package tui

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clitorhea/rhea-note/pkg/config"
	"github.com/clitorhea/rhea-note/pkg/crypto"
	"github.com/clitorhea/rhea-note/pkg/storage"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

type item struct {
	id      string
	modTime time.Time
}

func (i item) Title() string       { return i.id }
func (i item) Description() string { return i.modTime.Format(time.RFC1123) }
func (i item) FilterValue() string { return i.id }

type model struct {
	list        list.Model
	password    textinput.Model
	viewport    viewport.Model
	state       int // 0: list, 1: password prompt, 2: viewing note
	selectedID  string
	noteContent string
	cfg         *config.Config
	err         error
	width       int
	height      int
}

func InitialModel(cfg *config.Config, notes map[string]storage.LocalNoteInfo) model {
	var items []list.Item
	
	// Sort by mod time descending
	var sortedNotes []storage.LocalNoteInfo
	for _, n := range notes {
		sortedNotes = append(sortedNotes, n)
	}
	sort.Slice(sortedNotes, func(i, j int) bool {
		return sortedNotes[i].UpdatedAt.After(sortedNotes[j].UpdatedAt)
	})

	for _, n := range sortedNotes {
		items = append(items, item{id: n.ID, modTime: n.UpdatedAt})
	}

	m := list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.Title = "SecNotes"

	ti := textinput.New()
	ti.Placeholder = "Master Password"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()

	vp := viewport.New(0, 0)

	return model{
		list:     m,
		password: ti,
		viewport: vp,
		cfg:      cfg,
		state:    0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.state {
		case 0: // List
			if msg.String() == "enter" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.selectedID = i.id
					
					// Check if password is in env
					envPw := os.Getenv("SECNOTES_PASSWORD")
					if envPw != "" {
						m.decryptNote(envPw)
						m.state = 2
						return m, nil
					}
					
					m.state = 1
					m.password.SetValue("")
					return m, textinput.Blink
				}
			}
			
		case 1: // Password prompt
			if msg.String() == "enter" {
				m.decryptNote(m.password.Value())
				m.state = 2
				return m, nil
			} else if msg.String() == "esc" {
				m.state = 0
				return m, nil
			}
			
		case 2: // Viewport
			if msg.String() == "esc" || msg.String() == "q" {
				m.state = 0
				m.err = nil
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.viewport.Width = msg.Width - h
		m.viewport.Height = msg.Height - v
		m.width = msg.Width
		m.height = msg.Height
	}

	var cmd tea.Cmd
	switch m.state {
	case 0:
		m.list, cmd = m.list.Update(msg)
	case 1:
		m.password, cmd = m.password.Update(msg)
	case 2:
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

func (m *model) decryptNote(password string) {
	key := crypto.DeriveKey(password, m.cfg.Salt)
	
	ciphertext, err := storage.LoadNoteLocal(m.cfg.StoreDir, m.selectedID)
	if err != nil {
		m.err = err
		return
	}
	
	plaintext, err := crypto.Decrypt(ciphertext, key)
	if err != nil {
		m.err = fmt.Errorf("failed to decrypt (wrong password?)")
		return
	}
	
	m.noteContent = string(plaintext)
	m.viewport.SetContent(m.noteContent)
	m.err = nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	switch m.state {
	case 0:
		return docStyle.Render(m.list.View())
	case 1:
		return docStyle.Render(fmt.Sprintf(
			"Enter Master Password for %s:\n\n%s\n\n(esc to cancel)",
			m.selectedID,
			m.password.View(),
		))
	case 2:
		if m.err != nil {
			return docStyle.Render(fmt.Sprintf("Error:\n%v\n\n(esc to go back)", m.err))
		}
		
		header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(fmt.Sprintf("Viewing Note: %s", m.selectedID))
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("(esc or q to go back)")
		
		content := fmt.Sprintf("%s\n\n%s\n\n%s", header, m.viewport.View(), footer)
		return docStyle.Render(content)
	}

	return ""
}
