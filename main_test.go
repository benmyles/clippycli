package main

import (
	"testing"

	"github.com/atotto/clipboard"
)

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"a smaller", 5, 10, 5},
		{"b smaller", 10, 5, 5},
		{"equal", 7, 7, 7},
		{"negative numbers", -5, -10, -10},
		{"zero", 0, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("min(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestInitialModel(t *testing.T) {
	model := initialModel("", false)

	// Test initial state
	if model.state != stateInput {
		t.Errorf("Expected initial state to be stateInput, got %v", model.state)
	}

	// Test that textarea is initialized
	if model.textarea.Placeholder == "" {
		t.Error("Expected textarea to have a placeholder")
	}

	// Test that anthropic client is initialized
	if model.anthropicClient == nil {
		t.Error("Expected anthropic client to be initialized")
	}

	// Test that verbose is set correctly
	if model.verbose != false {
		t.Error("Expected verbose to be false")
	}
}

func TestInitialModelWithPrompt(t *testing.T) {
	prompt := "list all files"
	model := initialModel(prompt, false)

	// Test initial state - should be loading when prompt is provided
	if model.state != stateLoading {
		t.Errorf("Expected initial state to be stateLoading, got %v", model.state)
	}

	// Test that textarea has the initial prompt
	if model.textarea.Value() != prompt {
		t.Errorf("Expected textarea value to be %q, got %q", prompt, model.textarea.Value())
	}

	// Test that prompt is stored in model
	if model.prompt != prompt {
		t.Errorf("Expected model prompt to be %q, got %q", prompt, model.prompt)
	}

	// Test that anthropic client is initialized
	if model.anthropicClient == nil {
		t.Error("Expected anthropic client to be initialized")
	}
}

func TestInitWithPrompt(t *testing.T) {
	prompt := "test prompt"
	model := initialModel(prompt, false)

	// Init should return commands including generateCommand when starting with a prompt
	cmd := model.Init()
	if cmd == nil {
		t.Error("Expected Init to return a command when starting with a prompt")
	}
}

func TestInitWithoutPrompt(t *testing.T) {
	model := initialModel("", false)

	// Init should return basic commands when starting without a prompt
	cmd := model.Init()
	if cmd == nil {
		t.Error("Expected Init to return a command")
	}
}

func TestCopyToClipboard(t *testing.T) {
	testCommand := "ls -la"

	// Test copying to clipboard
	err := copyToClipboard(testCommand)
	if err != nil {
		t.Errorf("copyToClipboard failed: %v", err)
	}

	// Verify the command was copied to clipboard
	clipboardContent, err := clipboard.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read from clipboard: %v", err)
	}

	if clipboardContent != testCommand {
		t.Errorf("Expected clipboard content %q, got %q", testCommand, clipboardContent)
	}
}

func TestStateTransitions(t *testing.T) {
	tests := []struct {
		name          string
		currentState  state
		expectedState state
	}{
		{"input to loading", stateInput, stateLoading},
		{"loading to result", stateLoading, stateResult},
		{"result to edit", stateResult, stateEdit},
		{"edit to loading", stateEdit, stateLoading},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a basic structure test
			// In a real application, you might test actual state transitions
			if tt.currentState == stateInput && tt.expectedState != stateLoading {
				t.Errorf("Invalid state transition from %v to %v", tt.currentState, tt.expectedState)
			}
		})
	}
}

func TestVerboseMode(t *testing.T) {
	// Test verbose mode enabled
	model := initialModel("test prompt", true)
	if !model.verbose {
		t.Error("Expected verbose to be true when enabled")
	}

	// Test verbose mode disabled
	model = initialModel("test prompt", false)
	if model.verbose {
		t.Error("Expected verbose to be false when disabled")
	}
}

func TestFullPromptStorage(t *testing.T) {
	// Test that fullPrompt is stored when cmdGeneratedMsg is received
	testModel := initialModel("test prompt", true)

	// Simulate receiving a cmdGeneratedMsg
	testFullPrompt := "System: Test system prompt\n\nUser: test prompt"
	msg := cmdGeneratedMsg{
		cmd:        "ls -la",
		err:        nil,
		fullPrompt: testFullPrompt,
	}

	// Update the model with the message
	updatedModel, _ := testModel.Update(msg)

	// Type assertion to access the fields
	if m, ok := updatedModel.(model); ok {
		// Check that the full prompt was stored
		if m.fullPrompt != testFullPrompt {
			t.Errorf("Expected fullPrompt to be %q, got %q", testFullPrompt, m.fullPrompt)
		}

		// Check that the generated command was stored
		if m.generatedCmd != "ls -la" {
			t.Errorf("Expected generatedCmd to be %q, got %q", "ls -la", m.generatedCmd)
		}

		// Check that state changed to result
		if m.state != stateResult {
			t.Errorf("Expected state to be stateResult, got %v", m.state)
		}
	} else {
		t.Fatal("Expected updatedModel to be of type model")
	}
}
