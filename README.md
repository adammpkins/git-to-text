# git-to-text

git-to-text is a Go-based implementation inspired by the Python project [gpt-repository-loader](https://github.com/mpoon/gpt-repository-loader). It converts the contents of a Git repository into a single text file, designed to help developers easily share or analyze their codebase in a linear format, especially for loading into an LLM.

## Acknowledgment

This project is a Go port of the original [gpt-repository-loader](https://github.com/mpoon/gpt-repository-loader) by mpoon. We appreciate their work and encourage you to check out the original Python implementation.

## Features

- Converts an entire Git repository into a single text file
- Respects `.gptignore` file for excluding specific files or patterns
- Supports custom preambles to provide context for the output
- Cross-platform compatibility (Windows, macOS, Linux)

## Installation

### Prerequisites

- Go 1.16 or higher

### Steps

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/git-to-text.git
   cd git-to-text
   ```

2. Install dependencies:
   ```
   go get github.com/bmatcuk/doublestar/v4
   ```

3. Build the project:
   ```
   go build
   ```

This will create an executable named `git-to-text` (or `git-to-text.exe` on Windows) in your project directory.

## Usage

Run the program with the following syntax:

```
./git-to-text /path/to/git/repository [-p /path/to/preamble.txt] [-o /path/to/output_file.txt]
```

### Arguments:

- `/path/to/git/repository`: The path to the Git repository you want to convert (required)
- `-p /path/to/preamble.txt`: Path to a custom preamble file (optional)
- `-o /path/to/output_file.txt`: Path for the output file (optional, defaults to `output.txt`)

### Example:

```
./git-to-text /home/user/projects/my-repo -p /home/user/preamble.txt -o /home/user/my-repo-output.txt
```

## .gptignore File

You can create a `.gptignore` file in the root of your Git repository to specify files or patterns to ignore. The syntax is similar to `.gitignore`. If no `.gptignore` file is found in the repository, the program will look for one in the same directory as the executable.

Example `.gptignore`:

```
*.log
node_modules/
build/
```

## Preamble

By default, the program uses a standard preamble to explain the structure of the output file. You can provide a custom preamble file using the `-p` option. If no custom preamble is specified, the program will attempt to use the repository's README.md file as the preamble.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. We encourage continuing the spirit of the original project by using AI assistance in development where possible.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to [mpoon](https://github.com/mpoon) for the original [gpt-repository-loader](https://github.com/mpoon/gpt-repository-loader) project that inspired this Go implementation.
- Thanks to the creators of the `doublestar` package for providing powerful file pattern matching capabilities.
