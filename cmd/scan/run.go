package scan

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"scanexpress/pkg/config"
	"scanexpress/pkg/ui"
)

func Run() {
	// Setup configuration manager
	cm, err := config.NewConfigManager()
	if err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		os.Exit(1)
	}

	// Define Cobra command
	rootCmd := &cobra.Command{
		Use:   "scan",
		Short: "List and select scanners using scanimage",
	}

	// Add flags
	var forceSelection bool
	rootCmd.Flags().BoolVarP(&forceSelection, "select", "s", false, "Force scanner selection even if one is already configured")

	// Run command
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Create and initialize the UI model
		model := ui.NewModel(cm)

		// If we have a saved config and not forcing selection, set initial state to page count
		if !forceSelection && cm.HasValidSavedConfig() {
			config := cm.GetConfig()

			// Pre-fill model with saved config
			model.SelectedDevice = config.ScannerDevice
			model.SelectedTitle = config.ScannerTitle
			model.SaveFolder = config.SaveFolder

			// Skip directly to page count state
			model.State = ui.StateEnteringPageCount

			fmt.Printf("Using saved scanner: %s\nSave folder: %s\n", config.ScannerTitle, config.SaveFolder)
		}

		// Start the UI program
		p := tea.NewProgram(model)
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}

		return nil
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
