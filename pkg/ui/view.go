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
	}

	return ""
}
