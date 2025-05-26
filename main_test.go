package main

import (
	"testing"
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
	model := initialModel("")

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
}

func TestInitialModelWithPrompt(t *testing.T) {
	prompt := "list all files"
	model := initialModel(prompt)

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
	model := initialModel(prompt)

	// Init should return commands including generateCommand when starting with a prompt
	cmd := model.Init()
	if cmd == nil {
		t.Error("Expected Init to return a command when starting with a prompt")
	}
}

func TestInitWithoutPrompt(t *testing.T) {
	model := initialModel("")

	// Init should return basic commands when starting without a prompt
	cmd := model.Init()
	if cmd == nil {
		t.Error("Expected Init to return a command")
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
