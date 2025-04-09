package main

import (
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

const AppName = `
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

const ResourcesFolder = "resources"

//go:embed pam.exe
var PAM embed.FS

func runPamFile(pamfilelocation string) (string, error) {
	// Extract the embedded executable to a temporary location
	tempDir, err := os.MkdirTemp("", "pam-app")
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to create temporary directory: %v", err)
		return "", errors.New(errorMessage)
	}
	defer os.RemoveAll(tempDir) // Clean up the executable directory when done
	// Note: We don't defer cleanup of the output directory as we need to return it

	exePath := filepath.Join(tempDir, "pam.exe")

	// Read the embedded executable
	exeData, err := PAM.ReadFile("pam.exe")
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to read embedded executable: %v\n", err)
		return "", errors.New(errorMessage)
	}

	// Write it to the temporary location
	// 0700 permissions for an executable file (Windows will ignore these, but it's more appropriate)
	err = os.WriteFile(exePath, exeData, 0700)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to write executable to disk: %v\n", err)
		return "", errors.New(errorMessage)
	}

	// Create a separate directory for output files
	outputDir, err := os.MkdirTemp("", "pam-output")
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to create output directory: %v", err)
		return "", errors.New(errorMessage)
	}

	// Get absolute paths for both directories
	outputDirEscaped, err := filepath.Abs(outputDir)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to get absolute path for output dir: %v\n", err)
		return "", errors.New(errorMessage)
	}

	// Properly handle Windows paths - use the output directory path in the /fo: parameter
	outputFlag := fmt.Sprintf("/fo:%s", outputDirEscaped)

	// Arguments for PAM.exe - don't modify backslashes in actual arguments
	args := []string{"extract", pamfilelocation, outputFlag}

	// Execute the extracted application
	cmd := exec.Command(exePath, args...)

	// Set working directory to the temp directory
	cmd.Dir = tempDir

	// Build command string for display
	fullCommand := fmt.Sprintf(`%s extract %s %s`, exePath, pamfilelocation, outputFlag)

	// Print the command being executed
	fmt.Printf("Executing command: %s\n", fullCommand)
	// Create pipes for stdout and stderr to capture and display output
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	// Provide stdin from the console
	cmd.Stdin = os.Stdin

	// Start the command but don't wait for it to complete
	if err = cmd.Start(); err != nil {
		fmt.Printf("Failed to start executable: %v\n", err)
		return "", errors.New(fmt.Sprintf("Failed to start executable: %v", err))
	}

	// Read output asynchronously
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stdoutPipe.Read(buffer)
			if n > 0 {
				fmt.Print(string(buffer[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stderrPipe.Read(buffer)
			if n > 0 {
				fmt.Print(string(buffer[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	// Wait for the command to complete
	err = cmd.Wait()
	if err != nil {
		// Report error but continue execution
		fmt.Printf("Command execution error: %v\n", err)
	} else {
		fmt.Println("Command executed successfully")
	}

	// Check for output files
	files, _ := filepath.Glob(filepath.Join(outputDir, "*"))
	if len(files) > 0 {
		fmt.Printf("Found %d files in output directory\n", len(files))
	}

	// Try to open the output folder for the user
	_ = openFolder(outputDirEscaped)

	// Return the absolute path of the output directory
	return outputDirEscaped, nil
}

func main() {
	fmt.Printf("\n%s\n", AppName)
	fmt.Printf("Version → %s \n", Version)
	fmt.Printf("Commit → %s \n", Commit)
	fmt.Printf("Date → %s \n", Date)

	// Define command line flags
	dirPtr := flag.String("dir", ".", "Directory to scan for JSON files")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	noUnicode := flag.Bool("no-unicode", false, "Fail if Unicode characters are found in values")
	runPam := flag.String("pam", "", "select the pam file to check")
	flag.Parse()

	rootDir := *dirPtr

	pamLocation := *runPam
	
	// Check if -pam flag was provided on command line
	pamFlagProvided := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "pam" {
			pamFlagProvided = true
		}
	})
	
	if pamFlagProvided && len(pamLocation) < 1 {
		fmt.Println("You must enter the pam file location if you are using the -pam flag!")
		os.Exit(0)
	}

	outputFileLocation, err := runPamFile(pamLocation)
	if len(outputFileLocation) < 1 {
		fmt.Println("output file is not defined")
	} else {
		fmt.Println("Output file is well defined")
	}

	completePath := path.Join(outputFileLocation, ResourcesFolder)

	var newRootDir string
	if len(completePath) < 1 {
		newRootDir = rootDir
	} else {
		newRootDir = completePath
	}
	fmt.Printf("\nWill validate this directory : %s\n", newRootDir)

	// Get absolute path for display purposes
	absPath, err := filepath.Abs(newRootDir)
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

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total files scanned: %d\n", totalFiles)
	fmt.Printf("  JSON files found: %d\n", validCount+invalidCount)
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

// openFolder opens the system file explorer to the specified folder
func openFolder(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cmd.Start()
}
