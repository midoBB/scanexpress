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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type model struct {
	devices         []string
	titles          []string
	selectedDevice  string
	selectedTitle   string
	list            list.Model
	listingScanners bool
	spinner         spinner.Model
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
	// Initialize Viper (though not used fully in this step)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path.Join(xdg.ConfigHome, "scanexpress"))
	_ = viper.ReadInConfig() // Ignore error if config doesn't exist

	// Setup spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	// Initialize model
	m := model{
		list:            list.New(make([]list.Item, 0), itemDelegate{}, 0, 0),
		listingScanners: true,
		spinner:         s,
	}
	m.list.Title = "Select a Scanner"

	// Define Cobra command
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "List and select scanners using scanimage",
		Run: func(cmd *cobra.Command, args []string) {
			p := tea.NewProgram(m)
			if err := p.Start(); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		},
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
	if m.listingScanners {
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
			m.listingScanners = false
			return m, nil
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			if len(m.devices) > 0 {
				selected, ok := m.list.SelectedItem().(scanItem)
				if ok {
					m.selectedDevice = selected.device
					m.selectedTitle = selected.title
					fmt.Printf("Selected scanner: %s\n", m.selectedTitle)
				}
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.listingScanners {
		return fmt.Sprintf("%s Looking for scanners...", m.spinner.View())
	}

	// Display an error message if no scanners were found
	if len(m.devices) == 1 && m.devices[0] == "No scanners found" {
		return "Error: No scanners found. Please connect a scanner and try again."
	}

	return m.list.View()
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
