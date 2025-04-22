package scanner

import (
	"fmt"
	"os/exec"

	// "path"
	"regexp"
	"strings"
	// "time"
)

// Scanner represents a physical scanner device
type Scanner struct {
	Device string // Device identifier (e.g., "brother5:bus1;dev4")
	Title  string // Human-readable name (e.g., "Brother DS-740D USB scanner")
}

// ScanResult represents the result of a scan operation
type ScanResult struct {
	Success  bool
	Error    error
	FilePath string
}

// ScanConfig holds configuration options for scanning
type ScanConfig struct {
	Device     string // Scanner device identifier
	SaveFolder string // Folder to save scanned files
	PageCount  int    // Number of pages to scan
	IsDuplex   bool   // Whether to scan both sides (duplex/recto-verso)
}

// ListScannersResult holds the result of listing scanners
type ListScannersResult struct {
	Scanners []Scanner
	Error    error
}

// ListScanners detects available scanners using scanimage
func ListScanners() ListScannersResult {
	cmd := exec.Command("scanimage", "-L")
	output, err := cmd.Output()
	if err != nil {
		return ListScannersResult{
			Error: fmt.Errorf("failed to list scanners: %v", err),
		}
	}

	// Extract device name and title
	// Example: "device `brother5:bus1;dev4' is a Brother DS-740D USB scanner"
	deviceRegex := regexp.MustCompile("`([^']+)'")
	titleRegex := regexp.MustCompile("is a (.+)$")

	lines := strings.Split(string(output), "\n")
	scanners := make([]Scanner, 0)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		deviceMatch := deviceRegex.FindStringSubmatch(line)
		titleMatch := titleRegex.FindStringSubmatch(line)

		if len(deviceMatch) > 1 {
			scanner := Scanner{
				Device: deviceMatch[1],
			}

			// Extract title if available
			if len(titleMatch) > 1 {
				scanner.Title = strings.TrimSpace(titleMatch[1])
			} else {
				scanner.Title = "Unknown Scanner"
			}

			scanners = append(scanners, scanner)
		}
	}

	return ListScannersResult{
		Scanners: scanners,
	}
}

// PerformScan performs a scan using the given configuration
func PerformScan(config ScanConfig) ScanResult {
	// For demonstration purposes, just print what we're going to do
	fmt.Printf("Would scan %d pages", config.PageCount)
	if config.IsDuplex {
		fmt.Printf(" in duplex mode (recto-verso)\n")
	} else {
		fmt.Printf(" in single-sided mode\n")
	}

	// For actual scanning, we would:
	// 1. Create a timestamp for the filename
	// timestamp := time.Now().Format("20060102_150405")
	// 2. Create filenames based on page count - either multiple files or batch
	// 3. Use scanimage with appropriate options including duplex if needed

	// For now, return a failure result as we're just implementing the UI part
	return ScanResult{
		Success: false,
	}
}
