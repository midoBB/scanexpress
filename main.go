package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// UI states
const (
	stateListingScanners = iota
	stateSelectingScanner
	stateEnteringSaveFolder
	stateDirectScan // New state for direct scanning
)

type model struct {
	devices        []string
	titles         []string
	selectedDevice string
	selectedTitle  string
	saveFolder     string
	state          int
	list           list.Model
	spinner        spinner.Model
	folderInput    textinput.Model
}

type scanItem struct {
	device string
	title  string
}

var itemStyle = lipgloss.NewStyle().PaddingLeft(4)
var selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))

func (i scanItem) FilterValue() string { return i.device }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(scanItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type ScannersListedMsg struct {
	devices []string
	titles  []string
}

func main() {
	// Initialize Viper configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	configPath := path.Join(xdg.ConfigHome, "scanexpress")
	viper.AddConfigPath(configPath)

	// Create config directory if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		os.MkdirAll(configPath, 0755)
	}

	// Try to read config, ignore error if file doesn't exist
	_ = viper.ReadInConfig()

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

	// Initialize model
	m := model{
		list:        list.New(make([]list.Item, 0), itemDelegate{}, 0, 0),
		spinner:     s,
		folderInput: ti,
	}
	m.list.Title = "Select a Scanner"

	// Define Cobra command
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "List and select scanners using scanimage",
	}

	// Add flags
	var forceSelection bool
	cmd.Flags().BoolVarP(&forceSelection, "select", "s", false, "Force scanner selection even if one is already configured")

	// Run command
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Check if we have saved config values and should use them
		savedDevice := viper.GetString("scanner.device")
		savedTitle := viper.GetString("scanner.title")
		savedFolder := viper.GetString("save.folder")

		// If we have saved values, the folder exists, and we're not forcing selection
		if !forceSelection && savedDevice != "" && savedTitle != "" && savedFolder != "" {
			if _, err := os.Stat(savedFolder); !os.IsNotExist(err) {
				// We have everything we need, skip UI and perform scan directly
				fmt.Printf("Using saved scanner: %s\nSave folder: %s\n", savedTitle, savedFolder)
				// Here you would directly perform the scan operation
				return nil
			} else {
				// Folder doesn't exist, start from scanner selection
				m.state = stateListingScanners
				// But still use the saved folder as default
				if savedFolder != "" {
					m.folderInput.SetValue(savedFolder)
				} else {
					// Default to home directory if no saved folder
					homeDir, err := os.UserHomeDir()
					if err == nil {
						m.folderInput.SetValue(homeDir)
					}
				}
			}
		} else {
			// No saved config or forced selection, start from scanner listing
			m.state = stateListingScanners
			// Default to home directory for folder input
			if savedFolder != "" {
				m.folderInput.SetValue(savedFolder)
			} else {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					m.folderInput.SetValue(homeDir)
				}
			}
		}

		p := tea.NewProgram(m)
		if err := p.Start(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return nil
	}

	cmd.Execute()
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		listScannersCmd,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateListingScanners:
		switch msg := msg.(type) {
		case ScannersListedMsg:
			m.devices = msg.devices
			m.titles = msg.titles

			// Check if no scanners were found and exit immediately
			if len(m.devices) == 1 && m.devices[0] == "No scanners found" {
				fmt.Println("Error: No scanners found. Please connect a scanner and try again.")
				os.Exit(1)
				return m, tea.Quit
			}

			m.list.SetItems(toListItems(m.devices, m.titles))
			m.state = stateSelectingScanner
			return m, nil
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case stateSelectingScanner:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyEnter {
				if len(m.devices) > 0 {
					selected, ok := m.list.SelectedItem().(scanItem)
					if ok {
						m.selectedDevice = selected.device
						m.selectedTitle = selected.title
						// Move to save folder input state
						m.state = stateEnteringSaveFolder
						return m, textinput.Blink
					}
				}
			}
		}

		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case stateEnteringSaveFolder:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				m.saveFolder = m.folderInput.Value()
				// Validate folder path
				if _, err := os.Stat(m.saveFolder); os.IsNotExist(err) {
					// Create the directory if it doesn't exist
					err := os.MkdirAll(m.saveFolder, 0755)
					if err != nil {
						fmt.Printf("Error creating directory: %v\n", err)
						return m, tea.Quit
					}
				}

				// Save to config
				viper.Set("scanner.device", m.selectedDevice)
				viper.Set("scanner.title", m.selectedTitle)
				viper.Set("save.folder", m.saveFolder)

				// Save config to file
				configPath := path.Join(xdg.ConfigHome, "scanexpress", "config.yaml")
				err := viper.WriteConfigAs(configPath)
				if err != nil {
					fmt.Printf("Error saving config: %v\n", err)
				}

				// Print final selection and save folder
				fmt.Printf("Selected scanner: %s\nSave folder: %s\n", m.selectedTitle, m.saveFolder)
				return m, tea.Quit
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}

			var cmd tea.Cmd
			m.folderInput, cmd = m.folderInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateListingScanners:
		return fmt.Sprintf("%s Looking for scanners...", m.spinner.View())

	case stateSelectingScanner:
		return m.list.View()

	case stateEnteringSaveFolder:
		return fmt.Sprintf(
			"Selected Scanner: %s\n\nSave scans to:\n\n%s\n\n(Press Enter to confirm, edit path to change)",
			m.selectedTitle,
			m.folderInput.View(),
		)
	}

	return ""
}

func listScannersCmd() tea.Msg {
	cmd := exec.Command("scanimage", "-L")
	output, err := cmd.Output()
	if err != nil {
		// Signal that no scanners were found
		return ScannersListedMsg{devices: []string{"No scanners found"}, titles: []string{"Error"}}
	}
	// Extract device name and title
	// Example: "device `brother5:bus1;dev4' is a Brother DS-740D USB scanner"
	deviceRegex := regexp.MustCompile("`([^']+)'")
	titleRegex := regexp.MustCompile("is a (.+)$")

	lines := strings.Split(string(output), "\n")
	devices := make([]string, 0)
	titles := make([]string, 0)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		deviceMatch := deviceRegex.FindStringSubmatch(line)
		titleMatch := titleRegex.FindStringSubmatch(line)

		if len(deviceMatch) > 1 {
			devices = append(devices, deviceMatch[1])

			// Extract title if available
			title := "Unknown Scanner"
			if len(titleMatch) > 1 {
				title = strings.TrimSpace(titleMatch[1])
			}
			titles = append(titles, title)
		}
	}

	if len(devices) == 0 {
		// Signal that no scanners were found
		devices = append(devices, "No scanners found")
		titles = append(titles, "No scanners available")
	}

	return ScannersListedMsg{devices: devices, titles: titles}
}

func toListItems(devices []string, titles []string) []list.Item {
	listItems := make([]list.Item, len(devices))
	for i, device := range devices {
		title := "Unknown Scanner"
		if i < len(titles) {
			title = titles[i]
		}
		listItems[i] = scanItem{
			device: device,
			title:  title,
		}
	}
	return listItems
}
