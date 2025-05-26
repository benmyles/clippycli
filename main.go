package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
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
	err             error
	width           int
	height          int
	anthropicClient *anthropic.Client
}

// Messages
type cmdGeneratedMsg struct {
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
)

func initialModel(initialPrompt string) model {
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
		}

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
			content.WriteString("\n")
			content.WriteString(helpStyle.Render("Press Enter to execute • E to edit prompt • Any other key to cancel"))
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

		systemPrompt := `You are a helpful command-line assistant. Given a user's description of what they want to do, generate a single, safe command that accomplishes their goal. 

Rules:
1. Return ONLY the command, no explanations or markdown
2. Make sure the command is safe and won't cause harm
3. Use common Unix/Linux commands when possible
4. If the request is unclear or potentially dangerous, suggest a safer alternative
5. For file operations, use relative paths unless absolute paths are specifically requested
6. Don't include commands that require sudo unless explicitly requested

Examples:
User: "list all files in current directory"
Response: ls -la

User: "find all .go files"
Response: find . -name "*.go"

User: "create a new directory called myproject"
Response: mkdir myproject`

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
			return cmdGeneratedMsg{err: err}
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

		return cmdGeneratedMsg{cmd: cmdText}
	}
}

func (m model) executeCommand() tea.Cmd {
	return tea.ExecProcess(exec.Command("sh", "-c", m.generatedCmd), func(err error) tea.Msg {
		// After command execution, quit the program
		return tea.Quit()
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// Handle help flags
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		fmt.Printf(`ClippyCLI - AI Command Generator

Usage:
  clippycli [prompt]

Examples:
  clippycli                           # Interactive mode
  clippycli "list all files"          # Quick mode with auto-generation
  clippycli "find large files"        # Generate command for finding large files

Options:
  -h, --help                          # Show this help message

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

	// Handle command-line arguments
	var initialPrompt string
	if len(os.Args) > 1 {
		// Join all arguments after the program name as the initial prompt
		initialPrompt = strings.Join(os.Args[1:], " ")
	}

	p := tea.NewProgram(
		initialModel(initialPrompt),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
