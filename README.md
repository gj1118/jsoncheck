# JSONCheck

A simple command-line tool to recursively scan directories and identify JSON files with syntax errors.

## Features

- Recursively traverses directories to find all JSON files
- Validates JSON syntax and reports errors
- Provides summary statistics of valid and invalid files
- Cross-platform support (Windows, macOS, Linux)

## Installation

### Download Binary

Download the latest release for your platform from the [Releases page](https://github.com/yourusername/jsoncheck/releases).

### Build from Source

```bash
git clone https://github.com/yourusername/jsoncheck.git
cd jsoncheck
go build -o jsoncheck
```

## Usage

Basic usage:

```bash
# Check JSON files in current directory and all subdirectories
jsoncheck

# Check JSON files in a specific directory
jsoncheck -dir=/path/to/your/directory
```

### Command Line Options

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-dir` | Directory to scan for JSON files | Current directory (`.`) | `-dir=/path/to/scan` |
| `-verbose` | Enable verbose output | `false` | `-verbose` |
| `-no-unicode` | Fail if Unicode characters are found in values | `false` | `-no-unicode` |
| `-pam` | Select a PAM file to extract and validate JSON from | Empty | `-pam=/path/to/file.pam` |
| `-l10nstrict` | Check existence of l10n folders (must be used with `-pam` flag) | `false` | `-l10nstrict` |

### Exit Codes

- `0`: All JSON files are valid (or no JSON files found)
- `1`: One or more JSON files have syntax errors

## Examples

```bash
# Check current directory with minimal output
jsoncheck

# Check a specific directory with verbose output
jsoncheck -dir=/path/to/project -verbose

# Check for Unicode characters in JSON values
jsoncheck -dir=./configs -no-unicode

# Extract and validate JSON from a PAM file
jsoncheck -pam=/path/to/file.pam

# Check l10n folders in a PAM file
jsoncheck -pam=/path/to/file.pam -l10nstrict

# Use in a script or CI pipeline
jsoncheck -dir=./configs && echo "All JSON configs are valid!"
```

## Output Example

```
Scanning for JSON errors in: /home/user/project

Invalid JSON in file: configs/settings.json
  Error: invalid character '}' looking for beginning of object key string

Summary:
  Total files scanned: 42
  JSON files found: 15
  Valid JSON files: 14
  Invalid JSON files: 1
  L10n validation: PASS
```

## License

[MIT License](LICENSE)
