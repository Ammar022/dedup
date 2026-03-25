package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateFilePriority(t *testing.T) {
	tests := []struct {
		filename string
		expected int
	}{
		// Original files (priority 0)
		{"document.pdf", 0},
		{"image.jpg", 0},
		{"report.xlsx", 0},
		{"file.txt", 0},

		// Numbered copies (priority 1000+)
		{"document (1).pdf", 1001},
		{"document (2).pdf", 1002},
		{"image-1.jpg", 1001},
		{"image-2.jpg", 1002},
		{"file_1.txt", 1001},
		{"file_10.txt", 1010},

		// Copy indicators (priority 500)
		{"document_copy.pdf", 500},
		{"image copy.jpg", 500},
		{"file_backup.txt", 500},
		{"report_duplicate.xlsx", 500},
		{"temp_file.txt", 500},

		// Download suffixes (priority 300)
		{"document_final.pdf", 300},
		{"image_latest.jpg", 300},
		{"report_v2.xlsx", 300},
		{"file_updated.txt", 300},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := calculateFilePriority(tt.filename)
			if result != tt.expected {
				t.Errorf("calculateFilePriority(%q) = %d, want %d", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestCalculateChecksum(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("Hello, World!")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Calculate checksum
	checksum1, err := calculateChecksum(testFile)
	if err != nil {
		t.Fatalf("calculateChecksum failed: %v", err)
	}

	// Verify checksum is not empty
	if checksum1 == "" {
		t.Error("Checksum should not be empty")
	}

	// Verify same content produces same checksum
	checksum2, err := calculateChecksum(testFile)
	if err != nil {
		t.Fatalf("calculateChecksum failed on second call: %v", err)
	}

	if checksum1 != checksum2 {
		t.Errorf("Same file should produce same checksum: %s != %s", checksum1, checksum2)
	}

	// Create another file with different content
	testFile2 := filepath.Join(tmpDir, "test2.txt")
	content2 := []byte("Different content")
	if err := os.WriteFile(testFile2, content2, 0644); err != nil {
		t.Fatalf("Failed to create second test file: %v", err)
	}

	checksum3, err := calculateChecksum(testFile2)
	if err != nil {
		t.Fatalf("calculateChecksum failed for second file: %v", err)
	}

	if checksum1 == checksum3 {
		t.Error("Different files should produce different checksums")
	}
}

func TestGetFileInfo(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "document (1).pdf")
	content := []byte("Test content")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileInfo, err := getFileInfo(testFile)
	if err != nil {
		t.Fatalf("getFileInfo failed: %v", err)
	}

	// Verify path
	if fileInfo.Path != testFile {
		t.Errorf("Path = %s, want %s", fileInfo.Path, testFile)
	}

	// Verify size
	if fileInfo.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", fileInfo.Size, len(content))
	}

	// Verify checksum is not empty
	if fileInfo.Checksum == "" {
		t.Error("Checksum should not be empty")
	}

	// Verify priority (document (1).pdf should have priority 1001)
	if fileInfo.Priority != 1001 {
		t.Errorf("Priority = %d, want 1001", fileInfo.Priority)
	}
}

func TestDeleteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create a test file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Test file should exist")
	}

	// Delete the file
	if err := deleteFile(testFile); err != nil {
		t.Fatalf("deleteFile failed: %v", err)
	}

	// Verify file no longer exists
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File should not exist after deletion")
	}
}

func TestFindDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	content1 := []byte("Same content")
	content2 := []byte("Different content")

	// Create duplicates with same content
	file1 := filepath.Join(tmpDir, "original.txt")
	file2 := filepath.Join(tmpDir, "original (1).txt")
	file3 := filepath.Join(tmpDir, "original (2).txt")

	// Create a unique file
	file4 := filepath.Join(tmpDir, "unique.txt")

	if err := os.WriteFile(file1, content1, 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, content1, 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}
	if err := os.WriteFile(file3, content1, 0644); err != nil {
		t.Fatalf("Failed to create file3: %v", err)
	}
	if err := os.WriteFile(file4, content2, 0644); err != nil {
		t.Fatalf("Failed to create file4: %v", err)
	}

	// Run findDuplicates in non-delete mode
	err := findDuplicates(tmpDir, false, false, false)
	if err != nil {
		t.Fatalf("findDuplicates failed: %v", err)
	}

	// Verify all files still exist (no deletion in scan mode)
	for _, file := range []string{file1, file2, file3, file4} {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("File %s should still exist in scan mode", file)
		}
	}
}

func TestFindDuplicatesWithDeletion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	content := []byte("Duplicate content")

	file1 := filepath.Join(tmpDir, "keep.txt")
	file2 := filepath.Join(tmpDir, "delete (1).txt")

	if err := os.WriteFile(file1, content, 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, content, 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Run findDuplicates with delete and force flags
	err := findDuplicates(tmpDir, true, true, false)
	if err != nil {
		t.Fatalf("findDuplicates with deletion failed: %v", err)
	}

	// Verify original file still exists
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Error("Original file should still exist")
	}

	// Verify duplicate file was deleted
	if _, err := os.Stat(file2); !os.IsNotExist(err) {
		t.Error("Duplicate file should have been deleted")
	}
}

func TestNoDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create unique files
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(file1, []byte("Content 1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("Content 2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Run findDuplicates
	err := findDuplicates(tmpDir, false, false, false)
	if err != nil {
		t.Fatalf("findDuplicates failed: %v", err)
	}

	// Both files should still exist
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Error("file1 should exist")
	}
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Error("file2 should exist")
	}
}
