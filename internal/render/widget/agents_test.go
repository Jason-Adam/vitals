package widget

import (
	"strings"
	"testing"
	"time"

	"github.com/Jason-Adam/vitals/internal/model"
)

func longAgent(name string, status string) model.AgentEntry {
	return model.AgentEntry{
		Name:       name,
		Status:     status,
		StartTime:  time.Now().Add(-30 * time.Second),
		DurationMs: 5000,
		ColorIndex: 0,
	}
}

func TestAgents_SingleAgent_NoExtraLines(t *testing.T) {
	agents := []model.AgentEntry{
		longAgent("Explore", "running"),
	}
	ctx := &model.RenderContext{
		TerminalWidth: 120,
		Transcript:    &model.TranscriptData{Agents: agents},
	}
	cfg := defaultCfg()

	result := Agents(ctx, cfg)
	if result.IsEmpty() {
		t.Fatal("expected non-empty result for 1 agent")
	}

	if len(result.ExtraLines) != 0 {
		t.Errorf("expected no extra lines for single agent, got %d", len(result.ExtraLines))
	}

	if !strings.Contains(result.PlainText, "Explore") {
		t.Errorf("expected 'Explore' in PlainText, got %q", result.PlainText)
	}
}

func TestAgents_MultipleAgents_StackedVertically(t *testing.T) {
	agents := []model.AgentEntry{
		longAgent("Explore", "running"),
		longAgent("Plan", "running"),
		longAgent("Review", "running"),
	}
	ctx := &model.RenderContext{
		TerminalWidth: 120,
		Transcript:    &model.TranscriptData{Agents: agents},
	}
	cfg := defaultCfg()

	result := Agents(ctx, cfg)
	if result.IsEmpty() {
		t.Fatal("expected non-empty result for 3 agents")
	}

	// First agent in main result.
	if !strings.Contains(result.PlainText, "Explore") {
		t.Errorf("expected first agent 'Explore' in PlainText, got %q", result.PlainText)
	}

	// Remaining agents in ExtraLines.
	if len(result.ExtraLines) != 2 {
		t.Fatalf("expected 2 extra lines, got %d", len(result.ExtraLines))
	}

	// Main result should NOT contain agents 2 and 3 — they're in ExtraLines.
	if strings.Contains(result.PlainText, "Plan") {
		t.Error("agent 'Plan' should be in ExtraLines, not PlainText")
	}
	if strings.Contains(result.PlainText, "Review") {
		t.Error("agent 'Review' should be in ExtraLines, not PlainText")
	}
}

func TestAgents_LongNames_NotTruncatedByWidget(t *testing.T) {
	agents := []model.AgentEntry{
		longAgent("Implement comprehensive test coverage for the entire parser subsystem", "running"),
	}
	ctx := &model.RenderContext{
		TerminalWidth: 200,
		Transcript:    &model.TranscriptData{Agents: agents},
	}
	cfg := defaultCfg()

	result := Agents(ctx, cfg)
	if result.IsEmpty() {
		t.Fatal("expected non-empty result")
	}

	// Long names should appear verbatim — no widget-level truncation.
	if !strings.Contains(result.PlainText, agents[0].Name) {
		t.Errorf("expected long name verbatim in PlainText, got %q", result.PlainText)
	}
}

func TestAgents_MaxFiveTotal(t *testing.T) {
	var agents []model.AgentEntry
	for i := 0; i < 7; i++ {
		agents = append(agents, longAgent("Agent"+string(rune('A'+i)), "running"))
	}
	ctx := &model.RenderContext{
		TerminalWidth: 120,
		Transcript:    &model.TranscriptData{Agents: agents},
	}
	cfg := defaultCfg()

	result := Agents(ctx, cfg)
	// 1 main + ExtraLines should total at most 5.
	total := 1 + len(result.ExtraLines)
	if total > 5 {
		t.Errorf("expected at most 5 agents total, got %d", total)
	}
}
