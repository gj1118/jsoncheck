package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Define directory flag with current directory as default
	dirPtr := flag.String("dir", ".", "Directory to scan for JSON files")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	rootDir := *dirPtr
	
	// Get absolute path for display purposes
	absPath, err := filepath.Abs(rootDir)
	if err == nil {
		rootDir = absPath
	}
	
	fmt.Printf("Scanning for JSON errors in: %s\n\n", rootDir)

	// Count of valid and invalid files
	validCount := 0
	invalidCount := 0
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
		} else {
			if *verbose {
				relPath, _ := filepath.Rel(rootDir, path)
				fmt.Printf("Valid JSON: %s\n", relPath)
			}
			validCount++
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the directory tree: %v\n", err)
		return
	}

	// Print summary
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total files scanned: %d\n", totalFiles)
	fmt.Printf("  JSON files found: %d\n", validCount + invalidCount)
	fmt.Printf("  Valid JSON files: %d\n", validCount)
	fmt.Printf("  Invalid JSON files: %d\n", invalidCount)
	
	if invalidCount > 0 {
		os.Exit(1) // Exit with error code if invalid files were found
	}
}
