package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// defaultIgnores contains directories that are typically not relevant to code analysis.
// This exhaustive list attempts to cover nearly every language and ecosystem:
//   - Version control & IDEs: .git, .idea, .vscode, .vs
//   - Node.js & front-end: node_modules, bower_components, dist, build, .next
//   - PHP & Ruby: vendor, .bundle, log, tmp, cache
//   - Go: bin, pkg, vendor
//   - Rust: target
//   - Zig: zig-out
//   - Java/Gradle/Maven: .gradle, out, target
//   - Haskell: .stack-work
//   - Elixir/Erlang: _build, deps, ebin
//   - Python: __pycache__, .venv, env, .eggs
//   - .NET: bin, obj, TestResults
//   - Flutter/Dart: .dart_tool, build
//   - Swift/Xcode: DerivedData, xcuserdata
//   - C/C++ (CMake): CMakeFiles, cmake-build-debug, cmake-build-release
//   - iOS: Pods
//   - Unity: Library, Temp, Logs
//   - Unreal Engine: Binaries, Intermediate, Saved
//   - R: Rproj.user
//   - Bazel: bazel-out, bazel-bin, bazel-testlogs, bazel-genfiles
//   - Nim: nimcache
//   - Elm: elm-stuff
//   - Haxe: export
//   - Perl: blib
var defaultIgnores = []string{
	".git",
	".idea",
	".vscode",
	".vs",
	"node_modules",
	"vendor",
	"bower_components",
	"dist",
	"build",
	"coverage",
	"tmp",
	"cache",
	".sass-cache",
	".next",
	"target",
	".bundle",
	"log",
	"bin",
	"pkg",
	"zig-out",
	".gradle",
	"out",
	".stack-work",
	"_build",
	"deps",
	"__pycache__",
	".venv",
	"env",
	"obj",
	".dart_tool",
	"DerivedData",
	"CMakeFiles",
	"cmake-build-debug",
	"cmake-build-release",
	"Pods",
	"Library",
	"Temp",
	"Logs",
	"Binaries",
	"Intermediate",
	"Saved",
	"xcuserdata",
	"Rproj.user",
	"bazel-out",
	"bazel-bin",
	"bazel-testlogs",
	"bazel-genfiles",
	"nimcache",
	"TestResults",
	"elm-stuff",
	"export",
	".eggs",
	"blib",
	"ebin",
}

// getIgnoreList reads the .gptignore file in the repository and returns a slice of patterns,
// skipping empty lines and comments. If a pattern ends with a slash, it appends "**" to match all files.
func getIgnoreList(ignoreFilePath string) ([]string, error) {
	var ignoreList []string
	file, err := os.Open(ignoreFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines or comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// If pattern ends with a slash, append "**" to match all files inside the directory.
		if strings.HasSuffix(line, "/") {
			line = line + "**"
		}
		// On Windows, convert forward slashes to backslashes.
		if runtime.GOOS == "windows" {
			line = strings.ReplaceAll(line, "/", "\\")
		}
		ignoreList = append(ignoreList, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ignoreList, nil
}

// shouldIgnore checks if a given file path should be ignored by matching
// against the default directories and user-provided ignore patterns.
func shouldIgnore(filePath string, ignoreList []string) bool {
	// Check default ignore patterns first.
	for _, def := range defaultIgnores {
		if filePath == def || strings.HasPrefix(filePath, def+string(os.PathSeparator)) {
			return true
		}
	}
	// Then check patterns from the ignore file.
	for _, pattern := range ignoreList {
		match, _ := doublestar.Match(pattern, filePath)
		if match {
			return true
		}
	}
	return false
}

// isBinary uses a simple heuristic to determine if a file is binary.
// It returns true if any NUL byte is found.
func isBinary(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

// processRepository walks the repository, collects file paths (skipping ignored
// directories and files), sorts them, and then writes each text file's content
// to the output.
func processRepository(repoPath string, ignoreList []string, outputFile *os.File) error {
	var files []string

	// Walk through the repository using WalkDir for efficiency.
	err := filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}
		// If the directory should be ignored, skip it entirely.
		if d.IsDir() {
			if shouldIgnore(relPath, ignoreList) {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip the file if it matches an ignore pattern.
		if shouldIgnore(relPath, ignoreList) {
			return nil
		}
		files = append(files, relPath)
		return nil
	})
	if err != nil {
		return err
	}

	// Sort file paths to ensure deterministic ordering.
	sort.Strings(files)

	for _, relPath := range files {
		fullPath := filepath.Join(repoPath, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", fullPath, err)
		}
		// Skip binary files to avoid including non-text content.
		if isBinary(content) {
			continue
		}
		// Write a separator, the file's relative path, and its contents.
		_, err = fmt.Fprintf(outputFile, "----\n%s\n%s\n", relPath, string(content))
		if err != nil {
			return err
		}
	}
	return nil
}

// cloneRepository clones the repository from the given URL into a temporary directory.
func cloneRepository(repoURL string) (string, error) {
	tempDir, err := os.MkdirTemp("", "repo_clone_")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, tempDir)
	// Optionally, you can capture stdout/stderr for debugging:
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("git clone failed: %w", err)
	}
	return tempDir, nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <repository_path_or_github_url> [-p /path/to/preamble.txt] [-o /path/to/output_file.txt] [--unignore dir1,dir2,...]\n", os.Args[0])
		flag.PrintDefaults()
	}

	preambleFile := flag.String("p", "", "Path to preamble file")
	outputFilePath := flag.String("o", "output.txt", "Path to output file")
	// The unignore flag allows users to specify default ignored directories to include.
	unignoreFlag := flag.String("unignore", "", "Comma-separated list of default ignored directories to include")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}
	repoArg := args[0]

	cloned := false
	var repoPath string

	// Check if the provided repository argument is a GitHub URL.
	if strings.HasPrefix(repoArg, "http://") || strings.HasPrefix(repoArg, "https://") {
		fmt.Printf("Detected GitHub URL. Cloning repository...\n")
		tempDir, err := cloneRepository(repoArg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error cloning repository: %v\n", err)
			os.Exit(1)
		}
		repoPath = tempDir
		cloned = true
		// Ensure the temporary clone is cleaned up after processing.
		defer os.RemoveAll(tempDir)
	} else {
		repoPath = repoArg
	}

	// Process the unignore flag if provided.
	if *unignoreFlag != "" {
		unignoreList := strings.Split(*unignoreFlag, ",")
		// Trim spaces from each value.
		for i, v := range unignoreList {
			unignoreList[i] = strings.TrimSpace(v)
		}
		// Remove the unignored directories from the defaultIgnores.
		var newDefaults []string
		for _, def := range defaultIgnores {
			ignoreIt := false
			for _, un := range unignoreList {
				if def == un {
					ignoreIt = true
					break
				}
			}
			if !ignoreIt {
				newDefaults = append(newDefaults, def)
			}
		}
		defaultIgnores = newDefaults
	}

	// Look for a .gptignore file in the repository root.
	ignoreFilePath := filepath.Join(repoPath, ".gptignore")
	ignoreList := []string{}
	if _, err := os.Stat(ignoreFilePath); err == nil {
		ignoreList, err = getIgnoreList(ignoreFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading .gptignore file: %v\n", err)
			os.Exit(1)
		}
	}

	outFile, err := os.Create(*outputFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	// Write the preamble: either from the provided file or use the default.
	if *preambleFile != "" {
		preambleContent, err := os.ReadFile(*preambleFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading preamble file: %v\n", err)
			os.Exit(1)
		}
		_, err = outFile.Write(preambleContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing preamble to output file: %v\n", err)
			os.Exit(1)
		}
		_, err = outFile.WriteString("\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing newline after preamble: %v\n", err)
			os.Exit(1)
		}
	} else {
		defaultPreamble := "The following text is a Git repository with code. The structure of the text are sections that begin with ----, followed by a single line containing the file path and file name, followed by a variable amount of lines containing the file contents. The text representing the Git repository ends when the symbols --END-- are encountered. Any further text beyond --END-- are meant to be interpreted as instructions using the aforementioned Git repository as context.\n"
		_, err := outFile.WriteString(defaultPreamble)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing default preamble: %v\n", err)
			os.Exit(1)
		}
	}

	// Process the repository and output the file contents.
	err = processRepository(repoPath, ignoreList, outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing repository: %v\n", err)
		os.Exit(1)
	}

	_, err = outFile.WriteString("--END--")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing end marker: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Repository contents written to %s.\n", *outputFilePath)
	if cloned {
		fmt.Println("Temporary clone has been cleaned up.")
	}
}
