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

// PerformScan performs a scan using the given device and saves it to the specified folder
func PerformScan(device, saveFolder string) ScanResult {
	// Create a timestamp for the filename
	fmt.Printf("Here we would perform the scan\n")
	// timestamp := time.Now().Format("20060102_150405")
	// filename := fmt.Sprintf("scan_%s.png", timestamp)
	// outputPath := path.Join(saveFolder, filename)
	//
	// // Set up the scanimage command with basic options
	// cmd := exec.Command(
	// 	"scanimage",
	// 	"--device", device,
	// 	"--format=png",
	// 	"-o", outputPath,
	// )
	//
	// // Run the command
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	return ScanResult{
	// 		Success: false,
	// 		Error:   fmt.Errorf("scanning failed: %v - %s", err, string(output)),
	// 	}
	// }
	//
	// return ScanResult{
	// 	Success:  true,
	// 	FilePath: outputPath,
	// }
	return ScanResult{
		Success: false,
	}
}
