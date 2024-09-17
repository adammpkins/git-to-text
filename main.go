package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

func getIgnoreList(ignoreFilePath string) ([]string, error) {
	ignoreList := []string{}
	file, err := os.Open(ignoreFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if runtime.GOOS == "windows" {
			line = strings.ReplaceAll(line, "/", "\\")
		}
		ignoreList = append(ignoreList, strings.TrimSpace(line))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoreList, nil
}

func shouldIgnore(filePath string, ignoreList []string) bool {
	for _, pattern := range ignoreList {
		match, _ := doublestar.Match(pattern, filePath)
		if match {
			return true
		}
	}
	return false
}

func processRepository(repoPath string, ignoreList []string, outputFile *os.File) error {
	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, err := filepath.Rel(repoPath, path)
			if err != nil {
				return err
			}

			if !shouldIgnore(relPath, ignoreList) {
				content, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}

				_, err = fmt.Fprintf(outputFile, "----\n%s\n%s\n", relPath, string(content))
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run git_to_text.go /path/to/git/repository [-p /path/to/preamble.txt] [-o /path/to/output_file.txt]")
		os.Exit(1)
	}

	repoPath := os.Args[1]
	ignoreFilePath := filepath.Join(repoPath, ".gptignore")

	if _, err := os.Stat(ignoreFilePath); os.IsNotExist(err) {
		// Try to use the .gptignore file in the current directory as a fallback
		execPath, err := os.Executable()
		if err != nil {
			fmt.Println("Error getting executable path:", err)
			os.Exit(1)
		}
		ignoreFilePath = filepath.Join(filepath.Dir(execPath), ".gptignore")
	}

	var preambleFile string
	var outputFilePath string = "output.txt"

	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "-p" && i+1 < len(os.Args) {
			preambleFile = os.Args[i+1]
			i++
		} else if os.Args[i] == "-o" && i+1 < len(os.Args) {
			outputFilePath = os.Args[i+1]
			i++
		}
	}

	ignoreList := []string{}
	if _, err := os.Stat(ignoreFilePath); err == nil {
		var err error
		ignoreList, err = getIgnoreList(ignoreFilePath)
		if err != nil {
			fmt.Println("Error reading ignore file:", err)
			os.Exit(1)
		}
	}

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	if preambleFile != "" {
		preambleContent, err := ioutil.ReadFile(preambleFile)
		if err != nil {
			fmt.Println("Error reading preamble file:", err)
			os.Exit(1)
		}
		_, err = outputFile.Write(preambleContent)
		if err != nil {
			fmt.Println("Error writing preamble to output file:", err)
			os.Exit(1)
		}
		_, err = outputFile.WriteString("\n")
		if err != nil {
			fmt.Println("Error writing newline after preamble:", err)
			os.Exit(1)
		}
	} else {
		defaultPreamble := "The following text is a Git repository with code. The structure of the text are sections that begin with ----, followed by a single line containing the file path and file name, followed by a variable amount of lines containing the file contents. The text representing the Git repository ends when the symbols --END-- are encounted. Any further text beyond --END-- are meant to be interpreted as instructions using the aforementioned Git repository as context.\n"
		_, err := outputFile.WriteString(defaultPreamble)
		if err != nil {
			fmt.Println("Error writing default preamble:", err)
			os.Exit(1)
		}
	}

	err = processRepository(repoPath, ignoreList, outputFile)
	if err != nil {
		fmt.Println("Error processing repository:", err)
		os.Exit(1)
	}

	_, err = outputFile.WriteString("--END--")
	if err != nil {
		fmt.Println("Error writing end marker:", err)
		os.Exit(1)
	}

	fmt.Printf("Repository contents written to %s.\n", outputFilePath)
}
