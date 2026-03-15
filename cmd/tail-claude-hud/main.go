package main

import (
	"encoding/json"
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
)

// Smoke test: can lipgloss render styled output in a piped-stdout process?
// Claude Code invokes this every ~300ms, pipes JSON to stdin, reads stdout.

type stdinData struct {
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	Model          *struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"model"`
	ContextWindow *struct {
		Size         int      `json:"context_window_size"`
		UsedPercent  *float64 `json:"used_percentage"`
		CurrentUsage *struct {
			InputTokens              int `json:"input_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		} `json:"current_usage"`
	} `json:"context_window"`
}

func main() {
	// Detect TTY (no pipe) -- print initializing message and exit
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Println("[tail-claude-hud] waiting for claude-code...")
		return
	}

	// Read exactly one JSON object from stdin
	var input stdinData
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Println("[tail-claude-hud] initializing...")
		return
	}

	// Extract data
	modelName := "Unknown"
	if input.Model != nil && input.Model.DisplayName != "" {
		modelName = input.Model.DisplayName
	}

	contextSize := ""
	if input.ContextWindow != nil && input.ContextWindow.Size > 0 {
		contextSize = fmt.Sprintf(" (%s context)", formatTokens(input.ContextWindow.Size))
	}

	percent := 0
	if input.ContextWindow != nil && input.ContextWindow.UsedPercent != nil {
		percent = int(*input.ContextWindow.UsedPercent)
	}

	// Lipgloss styles
	dim := lipgloss.NewStyle().Faint(true)
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	cyan := lipgloss.NewStyle().Foreground(lipgloss.Color("87"))
	magenta := lipgloss.NewStyle().Foreground(lipgloss.Color("141"))

	// Pick context color by threshold
	contextColor := green
	if percent >= 85 {
		contextColor = red
	} else if percent >= 70 {
		contextColor = yellow
	}

	// Build context bar
	barWidth := 10
	filled := (percent * barWidth) / 100
	empty := barWidth - filled
	bar := contextColor.Render(repeat("█", filled)) + dim.Render(repeat("░", empty))

	// Model badge
	badge := fmt.Sprintf("[%s%s]", modelName, contextSize)

	// Project name from cwd
	projectName := "unknown"
	if input.Cwd != "" {
		projectName = lastPathSegment(input.Cwd)
	}

	// Compose line 1: [Model (size)] percent | project | env placeholders
	line1Parts := []string{
		cyan.Render(badge),
		bar + " " + contextColor.Render(fmt.Sprintf("%d%%", percent)),
		magenta.Bold(true).Render(projectName),
	}

	sep := dim.Render(" | ")
	line1 := join(line1Parts, sep)

	// Line 2: tool summary placeholder with nerdfont icons
	checkIcon := green.Render("") // nf-fa-check
	line2Parts := []string{
		checkIcon + " Bash " + dim.Render("x0"),
		checkIcon + " Edit " + dim.Render("x0"),
		checkIcon + " Read " + dim.Render("x0"),
	}
	line2 := join(line2Parts, sep)

	fmt.Println(line1)
	fmt.Println(line2)
}

func repeat(s string, n int) string {
	out := ""
	for range n {
		out += s
	}
	return out
}

func join(parts []string, sep string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += sep
		}
		out += p
	}
	return out
}

func lastPathSegment(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func formatTokens(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.0fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprintf("%d", n)
}
