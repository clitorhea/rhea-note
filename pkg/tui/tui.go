package tui

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/clitorhea/rhea-note/pkg/config"
	"github.com/clitorhea/rhea-note/pkg/crypto"
	"github.com/clitorhea/rhea-note/pkg/storage"
	"github.com/muesli/reflow/wrap"
)

var (
	docStyle  = lipgloss.NewStyle().Margin(1, 2)
	linkRegex = regexp.MustCompile(`\[\[(.*?)\]\]`)
)

type item struct {
	id      string
	modTime time.Time
}

func (i item) Title() string       { return i.id }
func (i item) Description() string { return i.modTime.Format(time.RFC1123) }
func (i item) FilterValue() string { return i.id }

type linkItem struct {
	id string
}

func (i linkItem) Title() string       { return i.id }
func (i linkItem) Description() string { return "Jump to this note" }
func (i linkItem) FilterValue() string { return i.id }

type model struct {
	list        list.Model
	password    textinput.Model
	viewport    viewport.Model
	linkList    list.Model
	editor               textarea.Model
	newPagePrompt        textinput.Model
	state                int // 0: list, 1: password prompt, 2: view note, 3: links menu, 4: delete confirm, 5: edit note, 6: new page
	selectedID           string
	noteContent          string
	cfg                  *config.Config
	err                  error
	width                int
	height               int
	history              []string
	pwCache              string
	creatingFromViewport bool
}

func InitialModel(cfg *config.Config, notes map[string]storage.LocalNoteInfo) model {
	m := list.New(buildItems(notes), list.NewDefaultDelegate(), 0, 0)
	m.Title = "SecNotes"

	ll := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	ll.Title = "Linked Notes"

	ti := textinput.New()
	ti.Placeholder = "Master Password"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()

	np := textinput.New()
	np.Placeholder = "New Page Name"
	np.Focus()

	vp := viewport.New(0, 0)

	ta := textarea.New()
	ta.Placeholder = "Write your note here..."
	ta.Focus()

	return model{
		list:          m,
		linkList:      ll,
		password:      ti,
		newPagePrompt: np,
		viewport:      vp,
		editor:        ta,
		cfg:           cfg,
		state:         0,
	}
}

func buildItems(notes map[string]storage.LocalNoteInfo) []list.Item {
	var items []list.Item
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
	return items
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
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
					m.history = []string{}
					return m.openNote()
				}
			}
			if msg.String() == "d" || msg.String() == "delete" {
				i, ok := m.list.SelectedItem().(item)
				if ok {
					m.selectedID = i.id
					m.state = 4 // confirm delete
					return m, nil
				}
			}
			if msg.String() == "n" {
				m.creatingFromViewport = false
				m.newPagePrompt.SetValue("")
				m.state = 6
				return m, textinput.Blink
			}

		case 1: // Password prompt
			if msg.String() == "enter" {
				m.pwCache = m.password.Value()
				m.decryptNote()
				m.state = 2
				return m, nil
			} else if msg.String() == "esc" {
				m.state = 0
				return m, nil
			}

		case 2: // Viewport
			if msg.String() == "esc" || msg.String() == "q" {
				if len(m.history) > 0 {
					m.selectedID = m.history[len(m.history)-1]
					m.history = m.history[:len(m.history)-1]
					return m.openNote()
				}
				m.state = 0
				m.err = nil
				cmd := m.refreshList()
				return m, cmd
			}
			if msg.String() == "tab" || msg.String() == "l" {
				matches := linkRegex.FindAllStringSubmatch(m.noteContent, -1)
				var items []list.Item
				seen := make(map[string]bool)
				for _, match := range matches {
					if len(match) > 1 {
						id := match[1]
						if !seen[id] {
							items = append(items, linkItem{id: id})
							seen[id] = true
						}
					}
				}
				if len(items) > 0 {
					m.linkList.SetItems(items)
					m.state = 3
				}
				return m, nil
			}
			if msg.String() == "e" {
				// Switch to edit mode
				m.editor.SetValue(m.noteContent)
				m.state = 5
				return m, nil
			}
			if msg.String() == "d" || msg.String() == "delete" {
				m.state = 4 // Confirm delete
				return m, nil
			}
			if msg.String() == "n" {
				m.creatingFromViewport = true
				m.newPagePrompt.SetValue("")
				m.state = 6
				return m, textinput.Blink
			}

		case 3: // Links menu
			if msg.String() == "esc" {
				m.state = 2
				return m, nil
			}
			if msg.String() == "enter" {
				i, ok := m.linkList.SelectedItem().(linkItem)
				if ok {
					m.history = append(m.history, m.selectedID)
					m.selectedID = i.id
					return m.openNote()
				}
			}

		case 4: // Delete confirm
			if strings.ToLower(msg.String()) == "y" {
				storage.DeleteNoteLocal(m.cfg.StoreDir, m.selectedID)
				m.state = 0
				cmd := m.refreshList()
				return m, cmd
			} else if msg.String() == "n" || msg.String() == "esc" || msg.String() == "enter" {
				m.state = 0
				return m, nil
			}

		case 5: // Edit mode
			if msg.String() == "esc" || msg.String() == "ctrl+s" {
				m.noteContent = m.editor.Value()
				m.encryptAndSaveNote()
				m.viewport.SetContent(wrap.String(m.noteContent, m.viewport.Width))
				m.state = 2
				return m, nil
			}

		case 6: // New page prompt
			if msg.String() == "esc" {
				if m.creatingFromViewport {
					m.state = 2
				} else {
					m.state = 0
				}
				return m, nil
			}
			if msg.String() == "enter" {
				newName := m.newPagePrompt.Value()
				if newName != "" {
					if m.creatingFromViewport {
						// Append to current note
						if m.noteContent != "" {
							m.noteContent += "\n"
						}
						m.noteContent += fmt.Sprintf("[[%s]]", newName)
						m.encryptAndSaveNote() // saves parent
						m.history = append(m.history, m.selectedID)
					} else {
						m.history = []string{}
					}
					
					// Navigate to new note
					m.selectedID = newName
					return m.openNote()
				}
			}
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		m.linkList.SetSize(msg.Width-h, msg.Height-v)
		m.viewport.Width = msg.Width - h
		m.viewport.Height = msg.Height - v
		m.editor.SetWidth(msg.Width - h)
		m.editor.SetHeight(msg.Height - v - 2)
		if m.noteContent != "" {
			m.viewport.SetContent(wrap.String(m.noteContent, m.viewport.Width))
		}
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
	case 3:
		m.linkList, cmd = m.linkList.Update(msg)
	case 5:
		m.editor, cmd = m.editor.Update(msg)
	case 6:
		m.newPagePrompt, cmd = m.newPagePrompt.Update(msg)
	}

	return m, cmd
}

func (m *model) refreshList() tea.Cmd {
	notes, _ := storage.ListNotesLocalWithTime(m.cfg.StoreDir)
	return m.list.SetItems(buildItems(notes))
}

func (m model) openNote() (tea.Model, tea.Cmd) {
	envPw := os.Getenv("SECNOTES_PASSWORD")
	if envPw != "" {
		m.pwCache = envPw
	}
	if m.pwCache != "" {
		m.decryptNote()
		m.state = 2
		return m, nil
	}
	m.state = 1
	m.password.SetValue("")
	return m, textinput.Blink
}

func (m *model) decryptNote() {
	key := crypto.DeriveKey(m.pwCache, m.cfg.Salt)
	
	ciphertext, err := storage.LoadNoteLocal(m.cfg.StoreDir, m.selectedID)
	if err != nil {
		// Note doesn't exist yet, we will start with empty content
		m.err = nil
		m.noteContent = ""
		m.viewport.SetContent("")
		return
	}
	
	plaintext, err := crypto.Decrypt(ciphertext, key)
	if err != nil {
		m.err = fmt.Errorf("failed to decrypt (wrong password?)")
		m.pwCache = "" // clear wrong password
		m.noteContent = ""
		m.viewport.SetContent("")
		return
	}
	
	m.noteContent = string(plaintext)
	m.viewport.SetContent(wrap.String(m.noteContent, m.viewport.Width))
	m.viewport.GotoTop()
	m.err = nil
}

func (m *model) encryptAndSaveNote() {
	key := crypto.DeriveKey(m.pwCache, m.cfg.Salt)
	ciphertext, err := crypto.Encrypt([]byte(m.noteContent), key)
	if err != nil {
		m.err = fmt.Errorf("failed to encrypt: %v", err)
		return
	}
	err = storage.SaveNoteLocal(m.cfg.StoreDir, m.selectedID, ciphertext)
	if err != nil {
		m.err = fmt.Errorf("failed to save: %v", err)
	}
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
			errText := fmt.Sprintf("Error:\n%v\n\n", m.err)
			if len(m.history) > 0 {
				errText += "(esc or q to go back)"
			} else {
				errText += "(esc or q to go to menu)"
			}
			return docStyle.Render(errText)
		}
		
		header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(fmt.Sprintf("Viewing Note: %s", m.selectedID))
		footerText := "(e: edit) • (d: delete) • (n: new page) • "
		if len(m.history) > 0 {
			footerText += fmt.Sprintf("(esc/q: back to %s)", m.history[len(m.history)-1])
		} else {
			footerText += "(esc/q: back)"
		}
		
		matches := linkRegex.FindAllStringSubmatch(m.noteContent, -1)
		if len(matches) > 0 {
			footerText += " • (tab/l: jump links)"
		}

		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(footerText)
		content := fmt.Sprintf("%s\n\n%s\n\n%s", header, m.viewport.View(), footer)
		return docStyle.Render(content)

	case 3:
		return docStyle.Render(m.linkList.View())
	
	case 4:
		return docStyle.Render(fmt.Sprintf("Are you sure you want to delete '%s'? (y/N)", m.selectedID))
	
	case 5:
		header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render(fmt.Sprintf("Editing Note: %s", m.selectedID))
		footer := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("(esc or ctrl+s to save and exit)")
		content := fmt.Sprintf("%s\n\n%s\n\n%s", header, m.editor.View(), footer)
		return docStyle.Render(content)
	
	case 6:
		return docStyle.Render(fmt.Sprintf(
			"Enter name for new page:\n\n%s\n\n(esc to cancel)",
			m.newPagePrompt.View(),
		))
	}

	return ""
}
