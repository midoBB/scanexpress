package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"scanexpress/pkg/config"
	"scanexpress/pkg/scanner"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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

// ScanPageCmd returns a command that scans a single page
func ScanPageCmd(device string, outputFile string, isDuplex bool, pageNum int) tea.Cmd {
	return func() tea.Msg {
		result := scanner.ScanPage(device, outputFile, isDuplex, pageNum)
		return PageScannedMsg{
			Result: result,
		}
	}
}

// GeneratePDFCmd returns a command that generates a PDF from scanned images
func GeneratePDFCmd(imageDir string) tea.Cmd {
	return func() tea.Msg {
		result := scanner.GeneratePDF(imageDir)
		return PDFGeneratedMsg{
			Result: result,
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

				// Move to page count input
				m.State = StateEnteringPageCount
				return m, textinput.Blink

			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}

			var cmd tea.Cmd
			m.FolderInput, cmd = m.FolderInput.Update(msg)
			return m, cmd
		}

	case StateEnteringPageCount:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				// Convert input to integer
				pageCount, err := strconv.Atoi(m.PageCountInput.Value())
				if err != nil || pageCount < 1 {
					// Default to 1 if invalid
					pageCount = 1
				}
				m.PageCount = pageCount

				// Move to duplex selection
				m.State = StateSelectingDuplexMode
				return m, nil

			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit

			case tea.KeyUp, tea.KeyLeft:
				// Increase page count
				currentValue := m.PageCountInput.Value()
				pageCount, err := strconv.Atoi(currentValue)
				if err != nil || pageCount < 1 {
					pageCount = 0
				}
				m.PageCountInput.SetValue(strconv.Itoa(pageCount + 1))
				return m, nil

			case tea.KeyDown, tea.KeyRight:
				// Decrease page count
				currentValue := m.PageCountInput.Value()
				pageCount, err := strconv.Atoi(currentValue)
				if err != nil || pageCount <= 1 {
					pageCount = 2
				}
				m.PageCountInput.SetValue(strconv.Itoa(pageCount - 1))
				return m, nil
			}

			// Handle additional keyboard shortcuts
			switch msg.String() {
			case "k", "p":
				// Increase page count
				currentValue := m.PageCountInput.Value()
				pageCount, err := strconv.Atoi(currentValue)
				if err != nil || pageCount < 1 {
					pageCount = 0
				}
				m.PageCountInput.SetValue(strconv.Itoa(pageCount + 1))
				return m, nil

			case "j", "n":
				// Decrease page count
				currentValue := m.PageCountInput.Value()
				pageCount, err := strconv.Atoi(currentValue)
				if err != nil || pageCount <= 1 {
					pageCount = 2
				}
				m.PageCountInput.SetValue(strconv.Itoa(pageCount - 1))
				return m, nil
			}

			var cmd tea.Cmd
			m.PageCountInput, cmd = m.PageCountInput.Update(msg)
			return m, cmd
		}

	case StateSelectingDuplexMode:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y":
				m.IsDuplex = true
				return m, nil

			case "n", "N":
				m.IsDuplex = false
				return m, nil

			case "enter":
				// Create timestamp for scan directory
				timestamp := time.Now().Format("20060102_150405")
				m.ScanOutputDir = filepath.Join(m.SaveFolder, "scan_"+timestamp)

				// Create scan directory
				err := os.MkdirAll(m.ScanOutputDir, 0755)
				if err != nil {
					fmt.Printf("Error creating directory: %v\n", err)
					return m, tea.Quit
				}

				// Initialize scanning state
				m.CurrentPage = 1
				m.ScannedFiles = []string{}

				// Move to waiting for first page
				m.State = StateWaitingForPageScan
				return m, nil

			case "ctrl+c", "esc":
				return m, tea.Quit
			}
		}

	case StateWaitingForPageScan:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				// Move to scanning state
				m.State = StateScanningPage

				// Create output file path
				outputFile := filepath.Join(m.ScanOutputDir, fmt.Sprintf("page_%03d.png", m.CurrentPage))

				// Start scan
				return m, tea.Batch(
					m.Spinner.Tick,
					ScanPageCmd(m.SelectedDevice, outputFile, m.IsDuplex, m.CurrentPage),
				)

			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			}
		}

	case StateScanningPage:
		switch msg := msg.(type) {
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd

		case PageScannedMsg:
			if msg.Result.Success {
				// Add to scanned files
				m.ScannedFiles = append(m.ScannedFiles, msg.Result.FilePaths...)

				// Check if we've scanned all pages
				if m.CurrentPage >= m.PageCount {
					// Move to PDF generation
					m.State = StateGeneratingPDF
					return m, tea.Batch(
						m.Spinner.Tick,
						GeneratePDFCmd(m.ScanOutputDir),
					)
				}

				// Move to next page
				m.CurrentPage++
				m.State = StateWaitingForPageScan
				return m, nil
			} else {
				// Scan failed
				m.ScanError = msg.Result.Error
				m.State = StateScanComplete
				return m, nil
			}

		case tea.KeyMsg:
			if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
				return m, tea.Quit
			}
		}

	case StateGeneratingPDF:
		switch msg := msg.(type) {
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd

		case PDFGeneratedMsg:
			if msg.Result.Success {
				m.GeneratedPDF = msg.Result.OutputPDF
			} else {
				m.ScanError = msg.Result.Error
			}
			// Move to completion state
			m.State = StateScanComplete
			return m, nil

		case tea.KeyMsg:
			if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
				return m, tea.Quit
			}
		}

	case StateScanComplete:
		return m, tea.Quit
	}

	return m, nil
}
