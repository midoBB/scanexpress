package scanner

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Scanner represents a physical scanner device
type Scanner struct {
	Device string // Device identifier (e.g., "brother5:bus1;dev4")
	Title  string // Human-readable name (e.g., "Brother DS-740D USB scanner")
}

// ScanResult represents the result of a scan operation
type ScanResult struct {
	Success      bool
	Error        error
	OutputDir    string   // Directory containing the scanned files
	ScannedFiles []string // List of scanned files
}

// ScanConfig holds configuration options for scanning
type ScanConfig struct {
	Device     string // Scanner device identifier
	SaveFolder string // Folder to save scanned files
	PageCount  int    // Number of pages to scan
	IsDuplex   bool   // Whether to scan both sides (duplex/recto-verso)
}

// PageScanResult holds the result of scanning a single page
type PageScanResult struct {
	Success  bool
	Error    error
	FilePath string
	PageNum  int
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

// ScanPage scans a single page and saves it to the specified folder
func ScanPage(device string, outputFile string, isDuplex bool, pageNum int) PageScanResult {
	// Set the source based on duplex mode
	source := "Automatic Document Feeder(left aligned)"
	if isDuplex {
		source = "Automatic Document Feeder(left aligned,Duplex)"
	}

	// Set up the scanimage command with required options
	cmd := exec.Command(
		"scanimage",
		"--device-name="+device,
		"--format=png",
		"--output-file="+outputFile,
		"--resolution=600",
		"--source="+source,
		"--AutoDeskew=yes",
		"--AutoDocumentSize=yes",
	)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return PageScanResult{
			Success:  false,
			Error:    fmt.Errorf("scanning page %d failed: %v - %s", pageNum, err, string(output)),
			PageNum:  pageNum,
			FilePath: "",
		}
	}

	return PageScanResult{
		Success:  true,
		PageNum:  pageNum,
		FilePath: outputFile,
	}
}

// PerformScan performs a scan using the given configuration
func PerformScan(config ScanConfig) ScanResult {
	// Create a timestamp for the folder name
	timestamp := time.Now().Format("20060102_150405")
	scanDir := path.Join(config.SaveFolder, "scan_"+timestamp)

	// Create scan directory
	err := os.MkdirAll(scanDir, 0755)
	if err != nil {
		return ScanResult{
			Success:      false,
			Error:        fmt.Errorf("failed to create scan directory: %v", err),
			OutputDir:    "",
			ScannedFiles: nil,
		}
	}

	// Track scanned files
	scannedFiles := make([]string, 0, config.PageCount)

	// Scan each page
	var scanErr error
	for i := 1; i <= config.PageCount; i++ {
		// Create output filename for this page
		outputFile := filepath.Join(scanDir, fmt.Sprintf("page_%03d.tiff", i))

		// Scan the page
		result := ScanPage(config.Device, outputFile, config.IsDuplex, i)

		if result.Success {
			scannedFiles = append(scannedFiles, result.FilePath)
		} else {
			scanErr = result.Error
			break
		}
	}

	// Check if all pages were scanned successfully
	if scanErr != nil {
		return ScanResult{
			Success:      false,
			Error:        scanErr,
			OutputDir:    scanDir,
			ScannedFiles: scannedFiles,
		}
	}

	return ScanResult{
		Success:      true,
		OutputDir:    scanDir,
		ScannedFiles: scannedFiles,
	}
}
