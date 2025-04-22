package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"scanexpress/pkg/config"
	"scanexpress/pkg/scanner"
)

// UI states
const (
	StateListingScanners = iota
	StateSelectingScanner
	StateEnteringSaveFolder
	StateEnteringPageCount
	StateSelectingDuplexMode
)

// Model represents the UI state
type Model struct {
	Devices        []string
	Titles         []string
	SelectedDevice string
	SelectedTitle  string
	SaveFolder     string
	PageCount      int
	IsDuplex       bool
	State          int
	List           list.Model
	Spinner        spinner.Model
	FolderInput    textinput.Model
	PageCountInput textinput.Model

	// Configuration manager
	ConfigManager *config.ConfigManager
}

// ScanItem represents an item in the scanner list
type ScanItem struct {
	Device string
	Title  string
}

// FilterValue defines how scan items are filtered
func (i ScanItem) FilterValue() string { return i.Device }

// ItemStyle for list items
var ItemStyle = lipgloss.NewStyle().PaddingLeft(4)

// SelectedItemStyle for selected list items
var SelectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))

// ItemDelegate defines how to display list items
type ItemDelegate struct{}

// Height of list items
func (d ItemDelegate) Height() int { return 1 }

// Spacing between list items
func (d ItemDelegate) Spacing() int { return 0 }

// Update list item
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// Render list item
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(ScanItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title)

	fn := ItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return SelectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// Messages
type ScannersListedMsg struct {
	Scanners []scanner.Scanner
	Error    error
}

// NewModel creates a new UI model
func NewModel(cm *config.ConfigManager) Model {
	// Setup spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	// Setup text input for save folder
	ti := textinput.New()
	ti.Placeholder = "Enter folder path to save scans"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	// Setup text input for page count
	pci := textinput.New()
	pci.Placeholder = "Enter number of pages to scan"
	pci.Focus()
	pci.CharLimit = 3
	pci.Width = 5

	// Initialize model
	m := Model{
		List:           list.New(make([]list.Item, 0), ItemDelegate{}, 0, 0),
		State:          StateListingScanners,
		Spinner:        s,
		FolderInput:    ti,
		PageCountInput: pci,
		ConfigManager:  cm,
		PageCount:      1,     // Default to 1 page
		IsDuplex:       false, // Default to single-sided
	}
	m.List.Title = "Select a Scanner"

	// If we have a saved config, use it for the folder
	config := cm.GetConfig()
	if config.SaveFolder != "" {
		m.FolderInput.SetValue(config.SaveFolder)
	} else {
		// Default to home directory
		homeDir, err := os.UserHomeDir()
		if err == nil {
			m.FolderInput.SetValue(homeDir)
		}
	}

	// Set default page count to "1"
	m.PageCountInput.SetValue("1")

	return m
}

// ToListItems converts scanners to list items
func ToListItems(scanners []scanner.Scanner) []list.Item {
	items := make([]list.Item, len(scanners))
	for i, s := range scanners {
		items[i] = ScanItem{
			Device: s.Device,
			Title:  s.Title,
		}
	}
	return items
}

// Init is called when the model is initialized
func (m Model) Init() tea.Cmd {
	switch m.State {
	case StateListingScanners:
		return tea.Batch(
			m.Spinner.Tick,
			ListScannersCmd(),
		)

	case StateEnteringPageCount:
		return textinput.Blink

	default:
		return nil
	}
}
