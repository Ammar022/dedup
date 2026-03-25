# Dedup - Intelligent Duplicate File Finder

A smart Go utility that finds duplicate files by checksum and intelligently recommends which files to delete based on common naming patterns.

## Installation

```bash
go install github.com/datumbrain/dedup@latest
```

## Usage

```bash
# Scan current directory (shows what would be deleted)
dedup

# Scan specific directory
dedup /path/to/folder
dedup "C:\Users\Username\Downloads"

# Delete duplicates with confirmation
dedup -d /path/to/folder

# Delete duplicates without confirmation (force)
dedup -d -f /path/to/folder

# Verbose output with detailed information
dedup -v /path/to/folder

# Combine flags
dedup -d -f -v ~/Downloads
```

### Flags

- **`-d`**: Delete duplicate files (asks for confirmation before deleting)
- **`-f`**: Force deletion without confirmation (**must be combined with `-d`**)
- **`-v`**: Verbose output showing detailed priority information and deletion commands

**Note**: The `-f` flag only works when combined with `-d`. Using `-f` alone will not delete files.

## Features

- **Smart Detection**: Uses SHA-256 checksums for accurate duplicate detection
- **Intelligent Recommendations**: Automatically identifies which files to keep vs delete
- **Simple & Clean UI**: Minimal output by default, verbose mode available with `-v` flag
- **Safe Deletion**: Delete duplicates with `-d` flag, confirmation prompt by default
- **Force Mode**: Skip confirmation with `-f` flag for automated workflows
- **Platform-Specific Commands**: Generates deletion commands in verbose mode for your OS
- **Safe by Default**: Scan-only mode unless `-d` flag is specified
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

### Default Mode (Simple & Clean)

```bash
$ dedup ~/Downloads

📁 Scanned 15 files in /Users/john/Downloads

🔍 Found 3 duplicate group(s)
💾 Can free 5.42 MB by deleting 8 file(s)

💡 Use -d to delete files (with confirmation)
💡 Use -d -f to delete without confirmation
💡 Use -v for detailed output
```

### Delete Mode with Confirmation

```bash
$ dedup -d ~/Downloads

📁 Scanned 15 files in /Users/john/Downloads

🔍 Found 3 duplicate group(s)
💾 Can free 5.42 MB by deleting 8 file(s)

Files to delete:
  1. invoice (1).pdf
  2. report-2.xlsx
  3. image_copy.jpg
  ...

⚠️  About to delete 8 file(s). Continue? [y/N]: y

🗑️  Deleting files...

✅ Deleted 8 file(s), freed 5.42 MB
```

### Verbose Mode

```bash
$ dedup -v ~/Downloads

Scanning files in: /Users/john/Downloads
Calculating checksums...
Processing: invoice.pdf
Processing: invoice (1).pdf
...

============================================================
DUPLICATE FILES REPORT
============================================================

Duplicate Group #1 (Checksum: a665a45920422f9d...)
File Size: 1024 bytes
Files with priorities:
  Priority 0: invoice.pdf
  Priority 1001: invoice (1).pdf
Decision:
  ✓ KEEP:   /Users/john/Downloads/invoice.pdf (Priority: 0)
  ✗ DELETE: /Users/john/Downloads/invoice (1).pdf (Priority: 1001)

...
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

## Testing

Run the unit tests:

```bash
go test -v
```

Run tests with coverage:

```bash
go test -v -cover
```

## License

MIT License - Feel free to use, modify, and distribute.

## Contributing

Issues and pull requests welcome! Please ensure any changes maintain the safety-first approach of the tool.
