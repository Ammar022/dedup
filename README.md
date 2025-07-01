# Dedup - Intelligent Duplicate File Finder

A smart Go utility that finds duplicate files by checksum and intelligently recommends which files to delete based on common naming patterns.

## Installation

```bash
go install github.com/datumbrain/dedup@latest
```

## Usage

```bash
# Scan current directory
dedup

# Scan specific directory
dedup /path/to/folder
dedup "C:\Users\Username\Downloads"
```

## Features

- **Smart Detection**: Uses SHA-256 checksums for accurate duplicate detection
- **Intelligent Recommendations**: Automatically identifies which files to keep vs delete
- **Platform-Specific Commands**: Generates deletion commands for your OS (Windows/macOS/Linux)
- **Safe by Default**: Only scans files, never deletes anything automatically
- **Non-Recursive**: Only scans the specified folder (doesn't go into subdirectories)

## How It Works

The tool prioritizes files based on common naming patterns:

### ✅ **KEEP** (Priority 0 - Original files)

- `document.pdf`
- `image.jpg`
- `report.xlsx`

### ❌ **DELETE** (Higher priority numbers)

- `document (1).pdf` - Downloaded copies
- `image-2.jpg` - Numbered variants
- `report_copy.xlsx` - Copy indicators
- `file_backup.txt` - Backup files
- `document_final.pdf` - Version suffixes

## Example Output

```raw
============================================================
DUPLICATE FILES REPORT
============================================================

Duplicate Group #1 (Checksum: a665a45920422f9d...)
File Size: 1024 bytes
Files:
  ✓ KEEP:   /Users/john/Downloads/invoice.pdf (Priority: 0)
  ✗ DELETE: /Users/john/Downloads/invoice (1).pdf (Priority: 1001)

============================================================
DELETION RECOMMENDATIONS
============================================================

Files recommended for deletion:
1. /Users/john/Downloads/invoice (1).pdf

Summary:
- Total duplicate groups: 1
- Files recommended for deletion: 1
- Disk space that can be freed: 1.00 MB (1048576 bytes)

============================================================
DELETION COMMANDS FOR DARWIN
============================================================

# Terminal (recommended):
rm "/Users/john/Downloads/invoice (1).pdf"

# Move to Trash (safer option):
osascript -e "tell application \"Finder\" to delete POSIX file \"/Users/john/Downloads/invoice (1).pdf\""
```

## Safety Features

- **Read-Only**: Never modifies or deletes files automatically
- **Platform-Aware**: Provides appropriate commands for your operating system
- **Trash Options**: Includes safer "move to trash" alternatives when available
- **Validation**: Checks file existence before generating deletion commands
- **Clear Warnings**: Reminds you to review recommendations before executing

## Supported Platforms

- **Windows**: Command Prompt, PowerShell, and Batch file commands
- **macOS**: Terminal commands and Finder trash integration
- **Linux**: Terminal commands with trash-cli support

## Build from Source

```bash
git clone https://github.com/datumbrain/dedup.git
cd dedup
go build -o dedup
```

## License

MIT License - Feel free to use, modify, and distribute.

## Contributing

Issues and pull requests welcome! Please ensure any changes maintain the safety-first approach of the tool.
