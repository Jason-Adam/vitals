package widget

import (
	"time"

	"github.com/Jason-Adam/vitals/internal/config"
	"github.com/Jason-Adam/vitals/internal/model"
)

// Agents renders running and recently-completed sub-agent entries.
// Running agents show a colored robot icon, half-circle indicator, and elapsed time.
// Completed agents show a dim colored robot icon, check mark, and duration.
// Returns an empty WidgetResult when ctx.Transcript is nil or there are no agents to show.
// FgColor is left empty because the widget composes multiple styles internally;
// the renderer passes the pre-styled Text through as-is.
func Agents(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	if ctx.Transcript == nil {
		return WidgetResult{}
	}

	icons := IconsFor(cfg.Style.Icons)
	agents := ctx.Transcript.Agents

	var running []model.AgentEntry
	var completed []model.AgentEntry
	for _, a := range agents {
		if a.Status == "running" {
			running = append(running, a)
		} else if a.DurationMs >= 1000 && !isStaleAgent(a) {
			// Only show completed agents that ran for >= 1s and finished
			// within the last 60s. Older agents are no longer actionable.
			completed = append(completed, a)
		}
	}

	// Show all running agents + last 2 completed, max 5 total.
	recent := completed
	if len(recent) > 2 {
		recent = recent[len(recent)-2:]
	}
	toShow := append(running, recent...)
	if len(toShow) > 5 {
		toShow = toShow[len(toShow)-5:]
	}

	if len(toShow) == 0 {
		return WidgetResult{}
	}

	// Stack agents vertically: first agent is the main result, remaining
	// agents are emitted as extra lines (one per row).
	first := toShow[0]
	fgColor := agentColors[first.ColorIndex%8]

	result := WidgetResult{
		Text:      formatAgentEntry(first, icons),
		PlainText: formatAgentEntryPlain(first, icons),
		FgColor:   fgColor,
	}

	for _, a := range toShow[1:] {
		color := agentColors[a.ColorIndex%8]
		result.ExtraLines = append(result.ExtraLines, WidgetResult{
			Text:      formatAgentEntry(a, icons),
			PlainText: formatAgentEntryPlain(a, icons),
			FgColor:   color,
		})
	}

	return result
}

// agentStaleThreshold is how long after completion before an agent entry
// is considered stale and hidden from the statusline.
const agentStaleThreshold = 60 * time.Second

// isStaleAgent reports whether a completed agent finished more than
// agentStaleThreshold ago. Returns false for running agents or agents
// without timing data.
func isStaleAgent(a model.AgentEntry) bool {
	if a.StartTime.IsZero() || a.DurationMs == 0 {
		return false
	}
	completedAt := a.StartTime.Add(time.Duration(a.DurationMs) * time.Millisecond)
	return time.Since(completedAt) > agentStaleThreshold
}

// formatAgentEntryPlain renders a single agent entry as unstyled text.
func formatAgentEntryPlain(a model.AgentEntry, icons Icons) string {
	displayName := a.Name
	if a.Description != "" {
		displayName = a.Description
	}
	// No fixed width cap — the render pipeline truncates at terminal width.
	modelSuffix := modelFamilySuffix(a.Model)
	label := icons.Task + " " + displayName + modelSuffix

	if a.Status == "running" {
		elapsed := formatElapsed(time.Since(a.StartTime))
		return label + " " + icons.Running + " " + elapsed
	}
	return label + " " + icons.Check + " " + formatDuration(a.DurationMs)
}

// formatAgentEntry renders a single agent entry with colored icon, running
// indicator or check mark, and elapsed/duration time.
func formatAgentEntry(a model.AgentEntry, icons Icons) string {
	style := AgentColorStyle(a.ColorIndex)
	icon := icons.Task
	modelSuffix := modelFamilySuffix(a.Model)

	// Prefer the description ("Structural completeness review") over the
	// subagent_type ("general-purpose") when available. The description is
	// the human-readable task label from the Agent tool_use input.
	displayName := a.Name
	if a.Description != "" {
		displayName = a.Description
	}
	// No fixed width cap — the render pipeline truncates at terminal width.

	if a.Status == "running" {
		elapsed := formatElapsed(time.Since(a.StartTime))
		label := icon + " " + displayName + modelSuffix
		return style.Render(label) + " " + style.Render(icons.Running) + " " + DimStyle.Render(elapsed)
	}

	// Completed: dim the colored icon, show check + duration.
	dimColorStyle := style.Faint(true)
	label := icon + " " + displayName + modelSuffix
	duration := formatDuration(a.DurationMs)
	return dimColorStyle.Render(label) + " " + greenStyle.Render(icons.Check) + DimStyle.Render(duration)
}

// modelFamilySuffix returns a parenthetical suffix for the model family if
// recognized, e.g. " (haiku)". Returns "" for unrecognized models.
func modelFamilySuffix(modelName string) string {
	family := ModelFamily(modelName)
	if family == "" {
		return ""
	}
	return " (" + family + ")"
}
