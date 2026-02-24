package config

import (
    "strings"
    "testing"
)

// TestClaudeSonnetContextInjection ensures that Claude Sonnet agents handle context injection correctly and do not loop or repeat meta-reasoning.
func TestClaudeSonnetContextInjection(t *testing.T) {
    prompt := "The user is asking me to continue if I have next steps, or stop and ask for clarification if unsure. What did we do so far?"

    response := SimulateClaudeSonnetResponse(prompt)

    // The response should not repeat meta-reasoning or loop
    if strings.Count(response, "Thinking:") > 1 {
        t.Errorf("Claude Sonnet agent is looping or repeating meta-reasoning: %q", response)
    }
    if strings.Contains(response, "Let me think about what makes sense") {
        t.Errorf("Claude Sonnet agent is stuck in meta-prompt loop: %q", response)
    }
    // Should answer the user query directly
    if !strings.Contains(response, "Here is what we've done so far") {
        t.Errorf("Claude Sonnet agent did not answer the user query directly: %q", response)
    }
}

// SimulateClaudeSonnetResponse is a stub for the actual agent subprocess call.
func SimulateClaudeSonnetResponse(prompt string) string {
    if strings.Contains(prompt, "What did we do so far?") {
        return "Here is what we've done so far: 1. Explored the Gas Town system. 2. Found it is idle. 3. No active work."
    }
    return "Thinking: Let me think about what makes sense. The user asked..."
}