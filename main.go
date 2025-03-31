package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const AppName =`
	██╗███████╗ ██████╗ ███╗   ██╗██╗   ██╗
     ██║██╔════╝██╔═══██╗████╗  ██║██║   ██║
     ██║███████╗██║   ██║██╔██╗ ██║██║   ██║
██   ██║╚════██║██║   ██║██║╚██╗██║╚██╗ ██╔╝
╚█████╔╝███████║╚██████╔╝██║ ╚████║ ╚████╔╝ 
 ╚════╝ ╚══════╝ ╚═════╝ ╚═╝  ╚═══╝  ╚═══╝
 `

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func main() {
	// Define command line flags
	dirPtr := flag.String("dir", ".", "Directory to scan for JSON files")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	noUnicode := flag.Bool("no-unicode", false, "Fail if Unicode characters are found in values")
	flag.Parse()

	rootDir := *dirPtr
	
	// Get absolute path for display purposes
	absPath, err := filepath.Abs(rootDir)
	if err == nil {
		rootDir = absPath
	}
	
	fmt.Printf("Scanning for JSON errors in: %s\n\n", rootDir)

	// Counts for summary
	validCount := 0
	invalidCount := 0
	unicodeCount := 0
	totalFiles := 0

	// Walk through the directory tree
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil // Continue walking despite the error
		}

		// Debug directories when verbose
		if *verbose && info.IsDir() {
			fmt.Printf("Entering directory: %s\n", path)
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		totalFiles++

		// Only process JSON files
		if !strings.HasSuffix(strings.ToLower(path), ".json") {
			if *verbose {
				fmt.Printf("Skipping non-JSON file: %s\n", path)
			}
			return nil
		}

		// Read the file
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", path, err)
			invalidCount++
			return nil
		}

		// Try to parse the JSON
		var result any
		err = json.Unmarshal(data, &result)
		if err != nil {
			relPath, _ := filepath.Rel(rootDir, path)
			fmt.Printf("Invalid JSON in file: %s\n", relPath)
			fmt.Printf("  Error: %v\n", err)
			invalidCount++
			return nil
		}

		// Check for Unicode characters if requested
		if *noUnicode {
			hasUnicode, unicodeSamples := checkForUnicode(result)
			if hasUnicode {
				relPath, _ := filepath.Rel(rootDir, path)
				fmt.Printf("Unicode characters found in file: %s\n", relPath)
				fmt.Printf("  Examples: %s\n", strings.Join(unicodeSamples[:min(len(unicodeSamples), 3)], ", "))
				unicodeCount++
			}
		}

		if *verbose {
			relPath, _ := filepath.Rel(rootDir, path)
			fmt.Printf("Valid JSON: %s\n", relPath)
		}
		validCount++

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the directory tree: %v\n", err)
		return
	}

	// Print summary
	fmt.Printf("\n%s\n", AppName)
	fmt.Printf("Version → %s \n", Version )
	fmt.Printf("Commit → %s \n", Commit)
	fmt.Printf("Date → %s \n", Date)
	
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total files scanned: %d\n", totalFiles)
	fmt.Printf("  JSON files found: %d\n", validCount + invalidCount)
	fmt.Printf("  Valid JSON files: %d\n", validCount)
	fmt.Printf("  Invalid JSON files: %d\n", invalidCount)
	fmt.Printf("\n\n")
	if *noUnicode {
		fmt.Printf("  Files with Unicode characters: %d\n", unicodeCount)
	}
	
	// Exit with error if any issues found
	if invalidCount > 0 || (*noUnicode && unicodeCount > 0) {
		os.Exit(1)
	}
}

// Recursively check for Unicode characters in JSON values
func checkForUnicode(v any) (bool, []string) {
	found := false
	samples := []string{}

	switch x := v.(type) {
	case string:
		// Check if the string contains non-ASCII characters
		for _, r := range x {
			if r > 127 || !unicode.IsPrint(r) {
				found = true
				samples = append(samples, fmt.Sprintf("%q (codepoint U+%04X)", string(r), r))
				break
			}
		}
	case map[string]any:
		// Recursively check all values in the map
		for _, val := range x {
			subFound, subSamples := checkForUnicode(val)
			if subFound {
				found = true
				samples = append(samples, subSamples...)
			}
		}
	case []any:
		// Recursively check all items in the array
		for _, item := range x {
			subFound, subSamples := checkForUnicode(item)
			if subFound {
				found = true
				samples = append(samples, subSamples...)
			}
		}
	}

	return found, samples
}

// Helper function for min of two ints (Go < 1.21)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
