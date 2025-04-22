package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"scanexpress/pkg/config"
	"scanexpress/pkg/scanner"
)

// ListScannersCmd returns a command that lists available scanners
func ListScannersCmd() tea.Cmd {
	return func() tea.Msg {
		result := scanner.ListScanners()
		return ScannersListedMsg{
			Scanners: result.Scanners,
			Error:    result.Error,
		}
	}
}

// Update handles state changes for the UI model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.State {
	case StateListingScanners:
		switch msg := msg.(type) {
		case ScannersListedMsg:
			if msg.Error != nil || len(msg.Scanners) == 0 {
				fmt.Printf("Error: No scanners found. Please connect a scanner and try again.\n")
				if msg.Error != nil {
					fmt.Printf("Details: %v\n", msg.Error)
				}
				return m, tea.Quit
			}

			// Store the scanner list
			m.Devices = make([]string, len(msg.Scanners))
			m.Titles = make([]string, len(msg.Scanners))
			for i, s := range msg.Scanners {
				m.Devices[i] = s.Device
				m.Titles[i] = s.Title
			}

			// Set the list items
			m.List.SetItems(ToListItems(msg.Scanners))
			m.State = StateSelectingScanner
			return m, nil

		case spinner.TickMsg:
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		}

	case StateSelectingScanner:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyEnter {
				if m.List.SelectedItem() != nil {
					selected, ok := m.List.SelectedItem().(ScanItem)
					if ok {
						m.SelectedDevice = selected.Device
						m.SelectedTitle = selected.Title
						// Move to save folder input state
						m.State = StateEnteringSaveFolder
						return m, textinput.Blink
					}
				}
			}
		}

		var cmd tea.Cmd
		m.List, cmd = m.List.Update(msg)
		return m, cmd

	case StateEnteringSaveFolder:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				m.SaveFolder = m.FolderInput.Value()

				// Validate folder path
				if _, err := os.Stat(m.SaveFolder); os.IsNotExist(err) {
					// Create the directory if it doesn't exist
					err := os.MkdirAll(m.SaveFolder, 0755)
					if err != nil {
						fmt.Printf("Error creating directory: %v\n", err)
						return m, tea.Quit
					}
				}

				// Save to config
				err := m.ConfigManager.SaveConfig(config.Config{
					ScannerDevice: m.SelectedDevice,
					ScannerTitle:  m.SelectedTitle,
					SaveFolder:    m.SaveFolder,
				})
				if err != nil {
					fmt.Printf("Error saving config: %v\n", err)
				}

				// Perform scan
				result := scanner.PerformScan(m.SelectedDevice, m.SaveFolder)
				if result.Success {
					fmt.Printf("Scan completed successfully!\nSaved to: %s\n", result.FilePath)
				} else {
					fmt.Printf("Scanning failed: %v\n", result.Error)
				}

				return m, tea.Quit

			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}

			var cmd tea.Cmd
			m.FolderInput, cmd = m.FolderInput.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}
