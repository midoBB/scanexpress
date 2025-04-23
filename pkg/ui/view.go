package ui

import (
	"fmt"
)

// View renders the current UI state
func (m Model) View() string {
	switch m.State {
	case StateListingScanners:
		return fmt.Sprintf("%s Looking for scanners...", m.Spinner.View())

	case StateSelectingScanner:
		return m.List.View()

	case StateEnteringSaveFolder:
		return fmt.Sprintf(
			"Selected Scanner: %s\n\nSave scans to:\n\n%s\n\n(Press Enter to confirm, edit path to change)",
			m.SelectedTitle,
			m.FolderInput.View(),
		)

	case StateEnteringPageCount:
		return fmt.Sprintf(
			"Selected Scanner: %s\n\nHow many pages to scan?\n\n%s\n\n(Press Enter to confirm)",
			m.SelectedTitle,
			m.PageCountInput.View(),
		)

	case StateSelectingDuplexMode:
		duplex := "No"
		if m.IsDuplex {
			duplex = "Yes"
		}
		return fmt.Sprintf(
			"Selected Scanner: %s\n\nNumber of pages: %d\n\nIs this a double-sided (recto-verso) document? %s\n\n(Press y/n to select, Enter to confirm)",
			m.SelectedTitle,
			m.PageCount,
			duplex,
		)

	case StateWaitingForPageScan:
		return fmt.Sprintf(
			"Ready to scan page %d of %d\n\nPlace the document in the scanner and press Enter when ready.",
			m.CurrentPage,
			m.PageCount,
		)

	case StateScanningPage:
		return fmt.Sprintf(
			"%s Scanning page %d of %d...",
			m.Spinner.View(),
			m.CurrentPage,
			m.PageCount,
		)

	case StateGeneratingPDF:
		return fmt.Sprintf(
			"%s Creating PDF document from %d scanned pages...",
			m.Spinner.View(),
			len(m.ScannedFiles),
		)

	case StateScanComplete:
		if m.ScanError != nil {
			return fmt.Sprintf(
				"Scanning failed at page %d of %d\nError: %v\n\nScanned %d pages successfully.\nFiles are located at: %s\n\nPress Enter to exit.",
				m.CurrentPage,
				m.PageCount,
				m.ScanError,
				len(m.ScannedFiles),
				m.ScanOutputDir,
			)
		}

		pdfMessage := ""
		if m.GeneratedPDF != "" {
			pdfMessage = fmt.Sprintf("\n\nA PDF document was created at: %s", m.GeneratedPDF)
		}

		return fmt.Sprintf(
			"Scan completed successfully!\nScanned %d pages.%s\n\nPress Enter to exit.",
			m.PageCount,
			pdfMessage,
		)
	}

	return ""
}
