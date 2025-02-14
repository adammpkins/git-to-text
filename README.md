# git-to-text

git-to-text is a Go-based tool inspired by the Python project [gpt-repository-loader](https://github.com/mpoon/gpt-repository-loader). It converts the contents of a Git repository into a single text file—ideal for loading into an LLM for repository analysis or chat-based interactions with your codebase.

## Acknowledgment

This project is a Go port of the original [gpt-repository-loader](https://github.com/mpoon/gpt-repository-loader) by mpoon. We appreciate their work and encourage you to check out the original Python implementation.

## Features

- Converts an entire Git repository into a single text file with clear file boundaries.
- Uses a detailed default ignore list to automatically skip build artifacts, caches, and dependency folders from nearly every ecosystem.
- Supports custom ignore patterns via a `.gptignore` file placed in the repository root.
- If a `.gptignore` pattern ends with a slash, it automatically appends `**` to match all files within that directory.
- Offers a `--unignore` flag so you can override default ignores and include specific directories if needed.
- Accepts a local repository path **or** a GitHub URL; if a URL is provided, the tool clones the repository (using a shallow clone) into a temporary directory and cleans it up afterward.
- Supports custom preamble files for contextual output.
- Ensures deterministic file ordering and skips binary files using a simple heuristic.

## Installation

### Prerequisites

- Go 1.16 or higher

### Steps

1. Clone the repository:
   ```
   git clone https://github.com/adammpkins/git-to-text.git
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

- `<repository_path_or_github_url>`: Either the path to the Git repository or a GitHub URL.
- `-p /path/to/preamble.txt`: Path to a custom preamble file (optional). If not provided, a default preamble is used.
- `-o /path/to/output_file.txt`: Path for the output file (optional, defaults to `output.txt`).
- `--unignore dir1,dir2,...`: (Optional) Comma-separated list of default ignored directories to include in the output.

### Examples:

- **Local Repository:**

```./git-to-text /home/user/projects/my-repo -p /home/user/preamble.txt -o /home/user/my-repo-output.txt```


- **GitHub URL:**
```./git-to-text https://github.com/adammpkins/my-repo --unignore node_modules,vendor```

The tool will clone the repository into a temporary directory, process it, and then clean up the clone.

## Default Ignores

By default, `git-to-text` automatically skips certain directories and files that are typically irrelevant to code analysis (e.g., build artifacts, caches, dependencies). Below is the exhaustive list:
   
- `.git`
- `.idea`
- `.vscode`
- `.vs`
- `node_modules`
- `vendor`
- `bower_components`
- `dist`
- `build`
- `coverage`
- `tmp`
- `cache`
- `.sass-cache`
- `.next`
- `target`
- `.bundle`
- `log`
- `bin`
- `pkg`
- `zig-out`
- `.gradle`
- `out`
- `_build`
- `deps`
- `pycache`
- `.venv`
- `env`
- `obj`
- `.dart_tool`
- `DerivedData`
- `CMakeFiles`
- `cmake-build-debug`
- `cmake-build-release`
- `Pods`
- `Library`
- `Temp`
- `Logs`
- `Binaries`
- `Intermediate`
- `Saved`
- `xcuserdata`
- `Rproj.user`
- `bazel-out`
- `bazel-bin`
- `bazel-testlogs`
- `bazel-genfiles`
- `nimcache`
- `TestResults`
- `elm-stuff`
- `export`
- `.eggs`
- `blib`
- `ebin`

Note: If any of these directories are important for your use case, you can include them via the --unignore flag (see above).



## .gptignore File

Place a `.gptignore` file in the **root** of your Git repository to specify files or patterns to ignore. The syntax is similar to `.gitignore`. Note that if a pattern ends with a slash (e.g., `logs/`), the tool will automatically append `**` so that all files within that directory are excluded.

Example `.gptignore`:
```logs/
bootstrap/
storage/
.env
```

## Preamble

By default, the tool uses a standard preamble explaining the output file's structure. You can override this by providing your own preamble file using the `-p` option.

# Additional Notes
 ### Binary Detection
git-to-text uses a simple heuristic to detect binary files: it scans each file for any NUL bytes (0x00). If a NUL byte is found, the file is considered binary and is automatically skipped. This helps ensure that non-text content or minified code isn't included in the output.

### GitHub URL Cloning
When you provide a GitHub URL (or any HTTP/HTTPS Git repository URL) instead of a local path, git-to-text performs a shallow clone using `git clone --depth 1` into a temporary directory. This minimizes both download size and processing time. After processing the repository, the temporary clone is automatically cleaned up. 

## Contributing

Contributions are welcome! Please submit a Pull Request. We encourage leveraging AI assistance in development while maintaining the spirit of the original project.



## License

This project is licensed under the MIT License – see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to [mpoon](https://github.com/mpoon) for the original [gpt-repository-loader](https://github.com/mpoon/gpt-repository-loader) project.
- Thanks to the creators of the `doublestar` package for providing powerful file pattern matching capabilities.

