package scanner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

// PageScanResult holds the result of scanning a single page or duplex pages
type PageScanResult struct {
	Success   bool
	Error     error
	FilePaths []string // List of file paths for the scanned pages
	PageNums  []int    // List of page numbers for the scanned pages
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

	// Get the directory of the output file
	outputDir := filepath.Dir(outputFile)

	// Set up the scanimage command with required options
	var cmd *exec.Cmd

	if isDuplex {
		// For duplex scanning, we change to the output directory and use -b
		// scanimage will generate out1.png and out2.png in the output directory
		cmd = exec.Command(
			"scanimage",
			"--device-name="+device,
			"--format=png",
			"-b",
			"--resolution=300",
			"--source="+source,
			"--AutoDeskew=yes",
			"--AutoDocumentSize=yes",
		)
		// Set the command's working directory to the output directory
		cmd.Dir = outputDir
	} else {
		// For normal scanning, use --output-file option
		cmd = exec.Command(
			"scanimage",
			"--device-name="+device,
			"--format=png",
			"--output-file="+outputFile,
			"--resolution=300",
			"--source="+source,
			"--AutoDeskew=yes",
			"--AutoDocumentSize=yes",
		)
	}

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return PageScanResult{
			Success:   false,
			Error:     fmt.Errorf("scanning page %d failed: %v - %s", pageNum, err, string(output)),
			FilePaths: nil,
			PageNums:  nil,
		}
	}

	// For duplex scanning, verify output files were created
	if isDuplex {
		// Check for generated out1.png and out2.png files
		// Note: For blank pages, some files might be missing, so we only need to find at least one
		files, err := filepath.Glob(filepath.Join(outputDir, "out*.png"))
		scannedFiles := make([]string, 0)
		if err != nil || len(files) == 0 {
			return PageScanResult{
				Success:   false,
				Error:     fmt.Errorf("duplex scan completed but no output files were found: %v", err),
				FilePaths: nil,
				PageNums:  nil,
			}
		}

		for i, file := range files {
			newFilename := filepath.Join(outputDir, fmt.Sprintf("page_%03d_%s.png",
				pageNum,
				getSideLabel(i)))

			// Rename the file
			err := os.Rename(file, newFilename)
			if err != nil {
				return PageScanResult{
					Success:   false,
					Error:     fmt.Errorf("failed to rename duplex scan file %s: %v", file, err),
					FilePaths: nil,
					PageNums:  nil,
				}
			}

			// Add the renamed file to our list
			scannedFiles = append(scannedFiles, newFilename)

		}

		// Success - we've verified output files exist
		return PageScanResult{
			Success:   true,
			FilePaths: files,
			PageNums:  []int{pageNum, pageNum + 1},
		}
	}

	// For non-duplex scanning, just return the specified output file
	return PageScanResult{
		Success:   true,
		FilePaths: []string{outputFile},
		PageNums:  []int{pageNum},
	}
}

// getSideLabel returns a label for the side of a duplex scan (A for front, B for back)
func getSideLabel(index int) string {
	if index == 0 {
		return "A" // Front side
	}
	return "B" // Back side
}
