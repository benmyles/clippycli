package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Application states
type state int

const (
	stateInput state = iota
	stateLoading
	stateResult
	stateEdit
)

// Model represents the application state
type model struct {
	state           state
	textarea        textarea.Model
	spinner         spinner.Model
	prompt          string
	generatedCmd    string
	copiedCmd       string // Track the command that was copied to clipboard
	err             error
	width           int
	height          int
	anthropicClient *anthropic.Client
	verbose         bool   // Show full prompt in verbose mode
	fullPrompt      string // Store the full prompt sent to AI
}

// Messages
type cmdGeneratedMsg struct {
	cmd        string
	err        error
	fullPrompt string // Include the full prompt that was sent to AI
}

type cmdCopiedMsg struct {
	cmd string
	err error
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1)

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#059669"))

	cmdStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Foreground(lipgloss.Color("#F9FAFB")).
			Padding(1).
			MarginTop(1).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6B7280"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DC2626")).
			Bold(true).
			MarginTop(1)

	verbosePromptStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#374151")).
				Foreground(lipgloss.Color("#D1D5DB")).
				Padding(1).
				MarginTop(1).
				MarginBottom(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#4B5563"))
)

func initialModel(initialPrompt string, verbose bool) model {
	// Initialize textarea
	ta := textarea.New()
	ta.Placeholder = "Describe what you want to do..."
	ta.Focus()
	ta.SetWidth(80)
	ta.SetHeight(3)

	// If we have an initial prompt, set it and adjust the UI
	if initialPrompt != "" {
		ta.SetValue(initialPrompt)
		// Move cursor to end of text
		ta.CursorEnd()
	}

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	// Initialize Anthropic client
	client := anthropic.NewClient()

	// Determine initial state based on whether we have a prompt
	initialState := stateInput
	if initialPrompt != "" {
		initialState = stateLoading
	}

	return model{
		state:           initialState,
		textarea:        ta,
		spinner:         s,
		prompt:          initialPrompt,
		anthropicClient: &client,
		verbose:         verbose,
	}
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		textarea.Blink,
		m.spinner.Tick,
	}

	// If we start in loading state (with initial prompt), generate command immediately
	if m.state == stateLoading && m.prompt != "" {
		cmds = append(cmds, m.generateCommand())
	}

	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(min(80, msg.Width-4))

	case tea.KeyMsg:
		switch m.state {
		case stateInput:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "enter":
				if strings.TrimSpace(m.textarea.Value()) != "" {
					m.prompt = m.textarea.Value()
					m.state = stateLoading
					return m, tea.Batch(
						m.spinner.Tick,
						m.generateCommand(),
					)
				}
			default:
				var cmd tea.Cmd
				m.textarea, cmd = m.textarea.Update(msg)
				cmds = append(cmds, cmd)
			}

		case stateResult:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "enter":
				if m.generatedCmd != "" {
					return m, m.executeCommand()
				}
			case "e":
				m.state = stateEdit
				m.textarea.SetValue(m.prompt)
				m.textarea.Focus()
				cmds = append(cmds, textarea.Blink)
			default:
				// Any other key cancels
				return m, tea.Quit
			}

		case stateEdit:
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit
			case "enter":
				if strings.TrimSpace(m.textarea.Value()) != "" {
					m.prompt = m.textarea.Value()
					m.state = stateLoading
					m.err = nil
					return m, tea.Batch(
						m.spinner.Tick,
						m.generateCommand(),
					)
				}
			default:
				var cmd tea.Cmd
				m.textarea, cmd = m.textarea.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case cmdGeneratedMsg:
		m.state = stateResult
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.generatedCmd = msg.cmd
			m.fullPrompt = msg.fullPrompt
		}

	case cmdCopiedMsg:
		if msg.err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not copy command to clipboard: %v\n", msg.err)
		} else {
			m.copiedCmd = msg.cmd
		}
		return m, tea.Quit

	case spinner.TickMsg:
		if m.state == stateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("🔧 ClippyCLI - AI Command Generator"))
	content.WriteString("\n\n")

	switch m.state {
	case stateInput:
		if strings.TrimSpace(m.textarea.Value()) != "" {
			content.WriteString(promptStyle.Render("Review your prompt:"))
		} else {
			content.WriteString(promptStyle.Render("What would you like to do?"))
		}
		content.WriteString("\n\n")
		content.WriteString(m.textarea.View())
		content.WriteString("\n")
		content.WriteString(helpStyle.Render("Press Enter to generate command • Ctrl+C/Esc to quit"))

	case stateLoading:
		content.WriteString(promptStyle.Render("Generating command for:"))
		content.WriteString("\n\n")
		if m.prompt != "" {
			// Show the prompt being processed
			promptDisplay := lipgloss.NewStyle().
				Italic(true).
				Foreground(lipgloss.Color("#6B7280")).
				Render("\"" + m.prompt + "\"")
			content.WriteString(promptDisplay)
			content.WriteString("\n\n")
		}
		content.WriteString(m.spinner.View() + " Thinking...")

	case stateResult:
		if m.err != nil {
			content.WriteString(errorStyle.Render("Error: " + m.err.Error()))
			content.WriteString("\n")
			content.WriteString(helpStyle.Render("Press any key to quit"))
		} else {
			content.WriteString(promptStyle.Render("Generated command:"))
			content.WriteString("\n")
			content.WriteString(cmdStyle.Render(m.generatedCmd))

			// Show verbose prompt if verbose mode is enabled
			if m.verbose && m.fullPrompt != "" {
				content.WriteString("\n")
				content.WriteString(promptStyle.Render("Full prompt sent to AI:"))
				content.WriteString("\n")
				content.WriteString(verbosePromptStyle.Render(m.fullPrompt))
			}

			content.WriteString("\n")
			content.WriteString(helpStyle.Render("Press Enter to copy to clipboard • E to edit prompt • Any other key to cancel"))
		}

	case stateEdit:
		content.WriteString(promptStyle.Render("Edit your prompt:"))
		content.WriteString("\n\n")
		content.WriteString(m.textarea.View())
		content.WriteString("\n")
		content.WriteString(helpStyle.Render("Press Enter to regenerate • Ctrl+C/Esc to quit"))
	}

	return content.String()
}

func (m model) generateCommand() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Get environment information
		envInfo := getEnvironmentInfo()

		systemPrompt := fmt.Sprintf(`You are a helpful command-line assistant. Given a user's description of what they want to do, generate a single, safe command that accomplishes their goal.

Environment Information:
%s

Rules:
1. Return ONLY the command, no explanations or markdown
2. Make sure the command is safe and won't cause harm
3. Use commands appropriate for the user's platform and shell
4. If the request is unclear or potentially dangerous, suggest a safer alternative
5. For file operations, use relative paths unless absolute paths are specifically requested
6. Don't include commands that require sudo unless explicitly requested
7. Consider the user's shell when generating commands (e.g., use appropriate syntax for bash, zsh, fish, etc.)
8. Take advantage of available environment variables when relevant

Examples:
User: "list all files in current directory"
Response: ls -la

User: "find all .go files"
Response: find . -name "*.go"

User: "create a new directory called myproject"
Response: mkdir myproject`, envInfo)

		// Create the full prompt that includes both system and user messages
		fullPrompt := fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, m.prompt)

		message, err := m.anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_20250514,
			MaxTokens: 1024,
			System: []anthropic.TextBlockParam{
				{Text: systemPrompt},
			},
			Messages: []anthropic.MessageParam{
				{
					Role: anthropic.MessageParamRoleUser,
					Content: []anthropic.ContentBlockParamUnion{
						{
							OfText: &anthropic.TextBlockParam{
								Text: m.prompt,
							},
						},
					},
				},
			},
		})

		if err != nil {
			return cmdGeneratedMsg{err: err, fullPrompt: fullPrompt}
		}

		// Extract the text from the response
		var cmdText string
		for _, block := range message.Content {
			if textBlock := block.AsAny(); textBlock != nil {
				if tb, ok := textBlock.(anthropic.TextBlock); ok {
					cmdText = strings.TrimSpace(tb.Text)
					break
				}
			}
		}

		return cmdGeneratedMsg{cmd: cmdText, fullPrompt: fullPrompt}
	}
}

func (m model) executeCommand() tea.Cmd {
	return func() tea.Msg {
		// Copy command to clipboard
		if err := copyToClipboard(m.generatedCmd); err != nil {
			return cmdCopiedMsg{cmd: "", err: err}
		}

		// Return success message with the copied command
		return cmdCopiedMsg{cmd: m.generatedCmd, err: nil}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// copyToClipboard copies the command to the clipboard
func copyToClipboard(command string) error {
	return clipboard.WriteAll(command)
}

// getEnvironmentInfo gathers environment information for the LLM prompt
func getEnvironmentInfo() string {
	var envInfo strings.Builder

	// Get current shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "unknown"
	}
	envInfo.WriteString(fmt.Sprintf("Shell: %s\n", shell))

	// Get platform and architecture
	envInfo.WriteString(fmt.Sprintf("Platform: %s\n", runtime.GOOS))
	envInfo.WriteString(fmt.Sprintf("Architecture: %s\n", runtime.GOARCH))

	// Get environment variable keys (but not values for security)
	envVars := os.Environ()
	var envKeys []string
	for _, env := range envVars {
		if parts := strings.SplitN(env, "=", 2); len(parts) == 2 {
			envKeys = append(envKeys, parts[0])
		}
	}

	// Sort environment variable keys for consistent output
	sort.Strings(envKeys)

	envInfo.WriteString("Available environment variables: ")
	envInfo.WriteString(strings.Join(envKeys, ", "))

	return envInfo.String()
}

func main() {
	// Handle help flags
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Printf(`ClippyCLI - AI Command Generator

Usage:
  clippycli [options] [prompt]

Examples:
  clippycli                           # Interactive mode
  clippycli "list all files"          # Quick mode with auto-generation
  clippycli -v "find large files"     # Verbose mode showing full AI prompt

Options:
  -h, --help                          # Show this help message
  -v                                  # Verbose mode: show full prompt sent to AI

Environment Variables:
  ANTHROPIC_API_KEY                   # Required: Your Anthropic API key

For more information, visit: https://github.com/benmyles/cliclippy
`)
		os.Exit(0)
	}

	// Check for API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		fmt.Fprintf(os.Stderr, "Error: ANTHROPIC_API_KEY environment variable is required\n")
		fmt.Fprintf(os.Stderr, "Please set your Anthropic API key: export ANTHROPIC_API_KEY=your_key_here\n")
		os.Exit(1)
	}

	// Parse command-line arguments
	var verbose bool
	var initialPrompt string
	var promptArgs []string

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "-v" {
			verbose = true
		} else {
			promptArgs = append(promptArgs, arg)
		}
	}

	if len(promptArgs) > 0 {
		initialPrompt = strings.Join(promptArgs, " ")
	}

	p := tea.NewProgram(
		initialModel(initialPrompt, verbose),
		tea.WithAltScreen(),
	)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}

	// Show the actual command that was copied to clipboard with styling
	if m, ok := finalModel.(model); ok && m.copiedCmd != "" {
		// Print styled success message
		successHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#059669")).
			Render("✓ Command copied to clipboard:")

		commandDisplay := lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Foreground(lipgloss.Color("#F9FAFB")).
			Padding(0, 1).
			MarginTop(1).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6B7280")).
			Render(m.copiedCmd)

		helpText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Italic(true).
			Render("Paste with Ctrl+V (or Cmd+V on macOS)")

		fmt.Printf("\n%s\n%s\n%s\n\n", successHeader, commandDisplay, helpText)
	}
}
