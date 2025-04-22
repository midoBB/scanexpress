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
	}

	return ""
}
