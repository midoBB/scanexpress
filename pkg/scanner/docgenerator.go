package scanner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

// PDFGenerationResult holds the result of PDF generation
type PDFGenerationResult struct {
	Success   bool
	Error     error
	OutputPDF string // Path to the generated PDF
}

// GeneratePDFMsg is sent when a PDF has been generated
type GeneratePDFMsg struct {
	Result PDFGenerationResult
}

// GeneratePDF converts scanned images to a PDF document
// It uses img2pdf to create the PDF from the images in the given directory
// After successful generation, it moves the PDF up one directory and removes the image directory
func GeneratePDF(imageDir string) PDFGenerationResult {
	// Get the parent directory name for the PDF filename
	parentDir := filepath.Dir(imageDir)
	dirName := filepath.Base(imageDir)
	pdfFile := dirName + ".pdf"
	pdfPath := filepath.Join(imageDir, pdfFile)

	// Get list of all PNG files in the directory
	pngFiles, err := filepath.Glob(filepath.Join(imageDir, "page_*.png"))
	if err != nil {
		return PDFGenerationResult{
			Success:   false,
			Error:     fmt.Errorf("failed to list PNG files: %v", err),
			OutputPDF: "",
		}
	}

	// Sort files in proper order to maintain page sequence
	sort.Strings(pngFiles)

	// Check if we have any files
	if len(pngFiles) == 0 {
		return PDFGenerationResult{
			Success:   false,
			Error:     fmt.Errorf("no PNG files found in directory"),
			OutputPDF: "",
		}
	}

	// Prepare command with img2pdf and all files as individual arguments
	args := append(pngFiles, "-o", pdfFile)
	cmd := exec.Command("img2pdf", args...)
	cmd.Dir = imageDir // Set working directory to the image directory
	output, err := cmd.CombinedOutput()

	// Check if command succeeded
	if err != nil {
		return PDFGenerationResult{
			Success:   false,
			Error:     fmt.Errorf("PDF generation failed: %v - %s", err, string(output)),
			OutputPDF: "",
		}
	}

	// Move the PDF one directory up
	destPDFPath := filepath.Join(parentDir, pdfFile)
	err = os.Rename(pdfPath, destPDFPath)
	if err != nil {
		return PDFGenerationResult{
			Success:   false,
			Error:     fmt.Errorf("failed to move PDF: %v", err),
			OutputPDF: pdfPath,
		}
	}

	// Clean up the image directory (best effort, don't fail if this doesn't work)
	os.RemoveAll(imageDir)

	return PDFGenerationResult{
		Success:   true,
		OutputPDF: destPDFPath,
	}
}
