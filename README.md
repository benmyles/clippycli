# ClippyCLI 🔧

A Go CLI utility that uses the Anthropic Claude API to generate shell commands based on natural language descriptions. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) for an interactive terminal user interface.

## Features

- 🤖 **AI-Powered Command Generation**: Uses Claude Sonnet 4 to generate safe, practical shell commands
- 🎨 **Beautiful TUI**: Interactive terminal interface with syntax highlighting and smooth animations
- ✏️ **Editable Prompts**: Modify your request and regenerate commands on the fly
- 🛡️ **Safety First**: Built-in safeguards to prevent dangerous commands
- ⚡ **Fast Execution**: Execute generated commands directly or cancel safely

## Installation

### Prerequisites

- Go 1.24.3 or later
- An Anthropic API key

### Build from Source

```bash
git clone https://github.com/benmyles/cliclippy.git
cd cliclippy
go build -o clippycli
```

### Install Globally

```bash
go install github.com/benmyles/cliclippy@latest
```

## Configuration

Set your Anthropic API key as an environment variable:

```bash
export ANTHROPIC_API_KEY="your_api_key_here"
```

You can add this to your shell profile (`.bashrc`, `.zshrc`, etc.) to make it permanent:

```bash
echo 'export ANTHROPIC_API_KEY="your_api_key_here"' >> ~/.zshrc
source ~/.zshrc
```

## Usage

### Basic Usage

Simply run the utility:

```bash
./clippycli
```

Or if installed globally:

```bash
clippycli
```

### Quick Mode with Auto-Generation

You can also pass your prompt as a command-line argument to automatically generate a command:

```bash
./clippycli "list all files in current directory"
```

```bash
clippycli "find large files over 100MB"
```

When you provide a prompt as an argument, ClippyCLI will immediately start generating the command and show you the result for review. This is especially useful for quick commands where you want to skip the input step.

### Getting Help

To see usage information and examples:

```bash
clippycli --help
# or
clippycli -h
```

### Interactive Flow

1. **Enter your request**: Describe what you want to do in natural language
   - Example: "list all .go files in the current directory"
   - Example: "create a backup of my config files"
   - Example: "find large files over 100MB"

2. **Review the generated command**: ClippyCLI will show you the command it generated

3. **Choose your action**:
   - **Press Enter**: Execute the command immediately
   - **Press 'e'**: Edit your original prompt and regenerate
   - **Press any other key**: Cancel and exit

### Example Sessions

#### Interactive Mode

Starting with no arguments:

```
🔧 ClippyCLI - AI Command Generator

What would you like to do?

┌─────────────────────────────────────────────────────────────────────────────┐
│ find all python files and count lines of code                              │
│                                                                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

Press Enter to generate command • Ctrl+C/Esc to quit
```

#### Quick Mode

Starting with auto-generation:

```bash
./clippycli "find all python files and count lines of code"
```

```
🔧 ClippyCLI - AI Command Generator

Generating command for:

"find all python files and count lines of code"

⠋ Thinking...
```

Then immediately shows the result:

```
🔧 ClippyCLI - AI Command Generator

Generated command:

┌─────────────────────────────────────────────────────────────────────────────┐
│ find . -name "*.py" -exec wc -l {} + | tail -1                             │
└─────────────────────────────────────────────────────────────────────────────┘

Press Enter to execute • E to edit prompt • Any other key to cancel
```

#### After Generation

In both modes, after pressing Enter:

```
🔧 ClippyCLI - AI Command Generator

Generated command:

┌─────────────────────────────────────────────────────────────────────────────┐
│ find . -name "*.py" -exec wc -l {} + | tail -1                             │
└─────────────────────────────────────────────────────────────────────────────┘

Press Enter to execute • E to edit prompt • Any other key to cancel
```

## Safety Features

ClippyCLI includes several safety measures:

- **Command Review**: Always shows the generated command before execution
- **Safe Defaults**: Avoids destructive operations unless explicitly requested
- **No Sudo by Default**: Won't suggest privileged commands unless specifically asked
- **Relative Paths**: Uses relative paths by default for file operations
- **User Confirmation**: Requires explicit confirmation before executing any command

## Examples

Here are some example prompts and the types of commands ClippyCLI might generate:

| Prompt | Generated Command |
|--------|-------------------|
| "list all files" | `ls -la` |
| "find large files" | `find . -type f -size +100M -ls` |
| "count lines in go files" | `find . -name "*.go" -exec wc -l {} + \| tail -1` |
| "create project directory" | `mkdir myproject` |
| "show disk usage" | `df -h` |
| "find recent files" | `find . -type f -mtime -1` |

### Command-line Usage Examples

```bash
# Quick file listing
clippycli "list all files"

# Find specific file types
clippycli "find all .txt files in subdirectories"

# System information
clippycli "show memory usage"

# File operations
clippycli "create a backup of config.json"

# Process management
clippycli "show running processes using port 8080"
```

## Keyboard Shortcuts

- **Ctrl+C / Esc**: Quit the application
- **Enter**: Submit prompt or execute command
- **e**: Edit the current prompt (when viewing results)
- **Any other key**: Cancel execution (when viewing results)

## Error Handling

If ClippyCLI encounters an error:

- **API Errors**: Network issues or API problems will be displayed with helpful messages
- **Invalid Commands**: The AI is prompted to generate safe, valid commands
- **Missing API Key**: Clear instructions for setting up authentication

## Development

### Project Structure

```
clippycli/
├── main.go          # Main application logic
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
└── README.md        # This file
```

### Dependencies

- **[anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go)**: Official Anthropic API client
- **[bubbletea](https://github.com/charmbracelet/bubbletea)**: Terminal UI framework
- **[bubbles](https://github.com/charmbracelet/bubbles)**: UI components for Bubble Tea
- **[lipgloss](https://github.com/charmbracelet/lipgloss)**: Terminal styling library

### Building

```bash
go build -o clippycli
```

### Testing

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Anthropic](https://www.anthropic.com/) for the Claude API
- [Charm](https://charm.sh/) for the excellent terminal UI libraries
- The Go community for the robust ecosystem

## Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/benmyles/cliclippy/issues) page
2. Create a new issue with detailed information about your problem
3. Include your Go version, OS, and any error messages

---

**⚠️ Disclaimer**: Always review generated commands before execution. While ClippyCLI includes safety measures, you are responsible for the commands you choose to run on your system. 