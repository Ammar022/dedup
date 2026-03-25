package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
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
		regexp.MustCompile(`\s*\(\d+\)$`), // file (1), file (2) at end
		regexp.MustCompile(`\s*-\s*\d+$`), // file-1, file-2 at end
		regexp.MustCompile(`\s*_\d+$`),    // file_1, file_2 at end
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

	osName := runtime.GOOS
	fmt.Printf("\n" + strings.Repeat("=", 60))
	fmt.Printf("\nDELETION COMMANDS FOR %s", strings.ToUpper(osName))
	fmt.Printf("\n" + strings.Repeat("=", 60))

	switch osName {
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

	switch osName {
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
	fmt.Printf("\n• Consider creating a backup of important files first\n")
}

// deleteFile deletes a file and returns an error if it fails
func deleteFile(filePath string) error {
	return os.Remove(filePath)
}

// confirmDeletion asks the user to confirm deletion
func confirmDeletion(filesToDelete []*FileInfo) bool {
	fmt.Printf("\n⚠️  About to delete %d file(s). Continue? [y/N]: ", len(filesToDelete))
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// findDuplicates finds all duplicate files in the specified folder
func findDuplicates(folderPath string, deleteMode bool, forceDelete bool, verbose bool) error {
	// Map to store checksum -> list of files with that checksum
	checksumMap := make(map[string][]*FileInfo)

	// Read directory contents
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}

	if verbose {
		fmt.Printf("Scanning files in: %s\n", folderPath)
		fmt.Println("Calculating checksums...")
	}

	// Process each file (skip directories)
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}

		filePath := filepath.Join(folderPath, entry.Name())

		if verbose {
			fmt.Printf("Processing: %s\n", entry.Name())
		}

		fileInfo, err := getFileInfo(filePath)
		if err != nil {
			fmt.Printf("Warning: Could not process %s: %v\n", filePath, err)
			continue
		}

		// Add to checksum map
		checksumMap[fileInfo.Checksum] = append(checksumMap[fileInfo.Checksum], fileInfo)
	}

	// Find and display duplicates
	if !verbose {
		fmt.Printf("\n📁 Scanned %d files in %s\n", len(checksumMap), folderPath)
	} else {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("DUPLICATE FILES REPORT")
		fmt.Println(strings.Repeat("=", 60))
	}

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

			if verbose {
				fmt.Printf("\nDuplicate Group #%d (Checksum: %s)\n", duplicateGroups, checksum[:16]+"...")
				fmt.Printf("File Size: %d bytes\n", files[0].Size)
				fmt.Println("Files with priorities:")

				// Show all files with their priorities for debugging
				for _, file := range files {
					fmt.Printf("  Priority %d: %s\n", file.Priority, filepath.Base(file.Path))
				}

				fmt.Println("Decision:")
			}

			// First file (lowest priority number) should be kept
			keepFile := files[0]
			if verbose {
				fmt.Printf("  ✓ KEEP:   %s (Priority: %d)\n", keepFile.Path, keepFile.Priority)
			}

			// Rest should be deleted
			for i := 1; i < len(files); i++ {
				file := files[i]
				if verbose {
					fmt.Printf("  ✗ DELETE: %s (Priority: %d)\n", file.Path, file.Priority)
				}
				filesToDelete = append(filesToDelete, file)
				totalSizeToSave += file.Size
			}
		}
	}

	if duplicateGroups == 0 {
		fmt.Println("\n✅ No duplicate files found!")
		return nil
	}

	// Summary output
	if verbose {
		fmt.Printf("\n" + strings.Repeat("=", 60))
		fmt.Printf("\nDELETION RECOMMENDATIONS")
		fmt.Printf("\n" + strings.Repeat("=", 60))
	}

	if len(filesToDelete) > 0 {
		if !verbose {
			fmt.Printf("\n🔍 Found %d duplicate group(s)\n", duplicateGroups)
			fmt.Printf("💾 Can free %.2f MB by deleting %d file(s)\n\n",
				float64(totalSizeToSave)/(1024*1024), len(filesToDelete))
		} else {
			fmt.Printf("\nFiles recommended for deletion:\n")
			for i, file := range filesToDelete {
				fmt.Printf("%d. %s\n", i+1, file.Path)
			}

			fmt.Printf("\nSummary:\n")
			fmt.Printf("- Total duplicate groups: %d\n", duplicateGroups)
			fmt.Printf("- Files recommended for deletion: %d\n", len(filesToDelete))
			fmt.Printf("- Disk space that can be freed: %.2f MB (%.0f bytes)\n",
				float64(totalSizeToSave)/(1024*1024), float64(totalSizeToSave))
		}

		// Delete mode
		if deleteMode {
			// Ask for confirmation unless force flag is set
			if !forceDelete {
				if !verbose {
					fmt.Println("Files to delete:")
					for i, file := range filesToDelete {
						fmt.Printf("  %d. %s\n", i+1, filepath.Base(file.Path))
					}
				}
				if !confirmDeletion(filesToDelete) {
					fmt.Println("\n❌ Deletion cancelled.")
					return nil
				}
			}

			// Perform deletion
			fmt.Println("\n🗑️  Deleting files...")
			deleted := 0
			var deletedSize int64
			for _, file := range filesToDelete {
				if err := deleteFile(file.Path); err != nil {
					fmt.Printf("  ❌ Failed to delete %s: %v\n", filepath.Base(file.Path), err)
				} else {
					deleted++
					deletedSize += file.Size
					if verbose {
						fmt.Printf("  ✓ Deleted: %s\n", file.Path)
					}
				}
			}
			fmt.Printf("\n✅ Deleted %d file(s), freed %.2f MB\n", deleted, float64(deletedSize)/(1024*1024))
		} else {
			// Show deletion commands only in verbose mode
			if verbose {
				generateDeletionCommands(filesToDelete)
			} else {
				fmt.Println("💡 Use -d to delete files (with confirmation)")
				fmt.Println("💡 Use -d -f to delete without confirmation")
				fmt.Println("💡 Use -v for detailed output")
			}
		}
	}

	return nil
}

func main() {
	// Define flags
	deleteFlag := flag.Bool("d", false, "Delete duplicate files")
	forceFlag := flag.Bool("f", false, "Force deletion without confirmation (use with -d)")
	verboseFlag := flag.Bool("v", false, "Verbose output with detailed information")
	flag.Parse()

	// Get folder path from remaining arguments or use current directory
	folderPath := "."
	if flag.NArg() > 0 {
		folderPath = flag.Arg(0)
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
	if err := findDuplicates(folderPath, *deleteFlag, *forceFlag, *verboseFlag); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
