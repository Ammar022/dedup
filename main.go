package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// FileInfo holds information about a file
type FileInfo struct {
	Path     string
	Checksum string
	Size     int64
	Priority int // Lower number = higher priority (keep this file)
}

// calculateFilePriority determines deletion priority based on filename patterns
// Lower number = higher priority (should be kept)
func calculateFilePriority(filename string) int {
	name := strings.ToLower(filepath.Base(filename))
	nameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))

	// Patterns for numbered copies (highest priority to delete)
	numberedPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\s*\(\d+\)$`),   // file (1), file (2) at end
		regexp.MustCompile(`\s*\(\d+\)\s*`), // file (1) anywhere in name
		regexp.MustCompile(`\s*-\s*\d+$`),   // file-1, file-2 at end
		regexp.MustCompile(`\s*_\d+$`),      // file_1, file_2 at end
		regexp.MustCompile(`\d+$`),          // file1, file2 at end (numbers only)
	}

	// Check for numbered patterns and extract number for sub-priority
	for _, pattern := range numberedPatterns {
		if matches := pattern.FindStringSubmatch(nameWithoutExt); matches != nil {
			// Extract number from the match
			numStr := regexp.MustCompile(`\d+`).FindString(matches[0])
			if num, err := strconv.Atoi(numStr); err == nil {
				return 1000 + num // Higher numbers get deleted first
			}
			return 1000 // Default for numbered files
		}
	}

	// Check for copy/duplicate indicators
	copyIndicators := []string{
		"copy", "copy of", "duplicate", "dup",
		"backup", "bak", "temp", "tmp",
	}

	for _, indicator := range copyIndicators {
		if strings.Contains(nameWithoutExt, indicator) {
			return 500 // Medium priority for deletion
		}
	}

	// Check for download suffixes
	downloadSuffixes := []string{
		"download", "downloaded", "new", "latest", "final", "v2", "v3", "updated",
	}

	for _, suffix := range downloadSuffixes {
		if strings.HasSuffix(nameWithoutExt, suffix) {
			return 300 // Lower priority for deletion
		}
	}

	// Original files get highest priority (lowest number = keep)
	return 0
}

// calculateChecksum computes SHA-256 checksum of a file
func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// getFileInfo returns file information including checksum and priority
func getFileInfo(filePath string) (*FileInfo, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	checksum, err := calculateChecksum(filePath)
	if err != nil {
		return nil, err
	}

	priority := calculateFilePriority(filePath)

	return &FileInfo{
		Path:     filePath,
		Checksum: checksum,
		Size:     stat.Size(),
		Priority: priority,
	}, nil
}

// generateDeletionCommands creates platform-specific deletion commands
func generateDeletionCommands(filesToDelete []*FileInfo) {
	if len(filesToDelete) == 0 {
		return
	}

	fmt.Printf("\n" + strings.Repeat("=", 60))
	fmt.Printf("\nDELETION COMMANDS FOR %s", strings.ToUpper(runtime.GOOS))
	fmt.Printf("\n" + strings.Repeat("=", 60))

	switch runtime.GOOS {
	case "windows":
		fmt.Printf("\n# Command Prompt (recommended):\n")
		for _, file := range filesToDelete {
			fmt.Printf("del \"%s\"\n", file.Path)
		}

		fmt.Printf("\n# PowerShell (alternative):\n")
		for _, file := range filesToDelete {
			fmt.Printf("Remove-Item \"%s\"\n", file.Path)
		}

		fmt.Printf("\n# Batch file (delete_duplicates.bat):\n")
		fmt.Printf("@echo off\n")
		fmt.Printf("echo Deleting %d duplicate files...\n", len(filesToDelete))
		for _, file := range filesToDelete {
			fmt.Printf("if exist \"%s\" (\n", file.Path)
			fmt.Printf("    echo Deleting: %s\n", filepath.Base(file.Path))
			fmt.Printf("    del \"%s\"\n", file.Path)
			fmt.Printf(")\n")
		}
		fmt.Printf("echo Done!\npause\n")

	case "darwin": // macOS
		fmt.Printf("\n# Terminal (recommended):\n")
		for _, file := range filesToDelete {
			fmt.Printf("rm \"%s\"\n", file.Path)
		}

		fmt.Printf("\n# Move to Trash (safer option):\n")
		for _, file := range filesToDelete {
			fmt.Printf("osascript -e \"tell application \\\"Finder\\\" to delete POSIX file \\\"%s\\\"\"\n", file.Path)
		}

		fmt.Printf("\n# Shell script (delete_duplicates.sh):\n")
		fmt.Printf("#!/bin/bash\n")
		fmt.Printf("echo \"Deleting %d duplicate files...\"\n", len(filesToDelete))
		for _, file := range filesToDelete {
			fmt.Printf("if [ -f \"%s\" ]; then\n", file.Path)
			fmt.Printf("    echo \"Deleting: %s\"\n", filepath.Base(file.Path))
			fmt.Printf("    rm \"%s\"\n", file.Path)
			fmt.Printf("fi\n")
		}
		fmt.Printf("echo \"Done!\"\n")

	case "linux":
		fmt.Printf("\n# Terminal (recommended):\n")
		for _, file := range filesToDelete {
			fmt.Printf("rm \"%s\"\n", file.Path)
		}

		fmt.Printf("\n# Move to Trash (if trash-cli is installed):\n")
		for _, file := range filesToDelete {
			fmt.Printf("trash \"%s\"\n", file.Path)
		}

		fmt.Printf("\n# Interactive deletion (asks for confirmation):\n")
		for _, file := range filesToDelete {
			fmt.Printf("rm -i \"%s\"\n", file.Path)
		}

		fmt.Printf("\n# Shell script (delete_duplicates.sh):\n")
		fmt.Printf("#!/bin/bash\n")
		fmt.Printf("echo \"Deleting %d duplicate files...\"\n", len(filesToDelete))
		for _, file := range filesToDelete {
			fmt.Printf("if [ -f \"%s\" ]; then\n", file.Path)
			fmt.Printf("    echo \"Deleting: %s\"\n", filepath.Base(file.Path))
			fmt.Printf("    rm \"%s\"\n", file.Path)
			fmt.Printf("fi\n")
		}
		fmt.Printf("echo \"Done!\"\n")

	default:
		fmt.Printf("\n# Generic Unix commands:\n")
		for _, file := range filesToDelete {
			fmt.Printf("rm \"%s\"\n", file.Path)
		}
	}

	// Platform-specific safety notes
	fmt.Printf("\n" + strings.Repeat("-", 60))
	fmt.Printf("\nSAFETY RECOMMENDATIONS:")
	fmt.Printf("\n" + strings.Repeat("-", 60))

	switch runtime.GOOS {
	case "windows":
		fmt.Printf("\n• Test with a few files first before running all commands")
		fmt.Printf("\n• Consider using PowerShell's -WhatIf parameter to preview actions")
		fmt.Printf("\n• Files deleted with 'del' go to Recycle Bin on most systems")
		fmt.Printf("\n• Create the batch file and review it before running")
	case "darwin":
		fmt.Printf("\n• Test with a few files first before running all commands")
		fmt.Printf("\n• Use the 'Move to Trash' option for safer deletion")
		fmt.Printf("\n• Make the shell script executable: chmod +x delete_duplicates.sh")
		fmt.Printf("\n• Files moved to trash can be recovered from Trash")
	case "linux":
		fmt.Printf("\n• Test with a few files first before running all commands")
		fmt.Printf("\n• Install trash-cli for safer deletion: sudo apt install trash-cli")
		fmt.Printf("\n• Use 'rm -i' for interactive confirmation")
		fmt.Printf("\n• Make the shell script executable: chmod +x delete_duplicates.sh")
	default:
		fmt.Printf("\n• Test with a few files first before running all commands")
		fmt.Printf("\n• Consider backing up important files before deletion")
	}

	fmt.Printf("\n• Always review the file list before executing any commands!")
	fmt.Printf("\n• Consider creating a backup of important files first")
}

// findDuplicates finds all duplicate files in the specified folder
func findDuplicates(folderPath string) error {
	// Map to store checksum -> list of files with that checksum
	checksumMap := make(map[string][]*FileInfo)

	// Read directory contents
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}

	fmt.Printf("Scanning files in: %s\n", folderPath)
	fmt.Println("Calculating checksums...")

	// Process each file (skip directories)
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		filePath := filepath.Join(folderPath, entry.Name())

		fmt.Printf("Processing: %s\n", entry.Name())

		fileInfo, err := getFileInfo(filePath)
		if err != nil {
			fmt.Printf("Warning: Could not process %s: %v\n", filePath, err)
			continue
		}

		// Add to checksum map
		checksumMap[fileInfo.Checksum] = append(checksumMap[fileInfo.Checksum], fileInfo)
	}

	// Find and display duplicates
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("DUPLICATE FILES REPORT")
	fmt.Println(strings.Repeat("=", 60))

	duplicateGroups := 0
	var filesToDelete []*FileInfo
	var totalSizeToSave int64

	for checksum, files := range checksumMap {
		if len(files) > 1 {
			duplicateGroups++

			// Sort files by priority (lower number = higher priority to keep)
			sort.Slice(files, func(i, j int) bool {
				return files[i].Priority < files[j].Priority
			})

			fmt.Printf("\nDuplicate Group #%d (Checksum: %s)\n", duplicateGroups, checksum[:16]+"...")
			fmt.Printf("File Size: %d bytes\n", files[0].Size)
			fmt.Println("Files:")

			// First file (lowest priority number) should be kept
			keepFile := files[0]
			fmt.Printf("  ✓ KEEP:   %s (Priority: %d)\n", keepFile.Path, keepFile.Priority)

			// Rest should be deleted
			for i := 1; i < len(files); i++ {
				file := files[i]
				fmt.Printf("  ✗ DELETE: %s (Priority: %d)\n", file.Path, file.Priority)
				filesToDelete = append(filesToDelete, file)
				totalSizeToSave += file.Size
			}
		}
	}

	if duplicateGroups == 0 {
		fmt.Println("\nNo duplicate files found!")
	} else {
		fmt.Printf("\n" + strings.Repeat("=", 60))
		fmt.Printf("\nDELETION RECOMMENDATIONS")
		fmt.Printf("\n" + strings.Repeat("=", 60))

		if len(filesToDelete) > 0 {
			fmt.Printf("\nFiles recommended for deletion:\n")
			for i, file := range filesToDelete {
				fmt.Printf("%d. %s\n", i+1, file.Path)
			}

			fmt.Printf("\nSummary:\n")
			fmt.Printf("- Total duplicate groups: %d\n", duplicateGroups)
			fmt.Printf("- Files recommended for deletion: %d\n", len(filesToDelete))
			fmt.Printf("- Disk space that can be freed: %.2f MB (%.0f bytes)\n",
				float64(totalSizeToSave)/(1024*1024), float64(totalSizeToSave))

			// Generate platform-specific deletion commands
			generateDeletionCommands(filesToDelete)
		}
	}

	return nil
}

func main() {
	// Get folder path from command line argument or use current directory
	folderPath := "."
	if len(os.Args) > 1 {
		folderPath = os.Args[1]
	}

	// Verify the path exists and is a directory
	stat, err := os.Stat(folderPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !stat.IsDir() {
		fmt.Printf("Error: %s is not a directory\n", folderPath)
		os.Exit(1)
	}

	// Find duplicates
	if err := findDuplicates(folderPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
