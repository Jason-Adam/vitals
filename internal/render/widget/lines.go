package widget

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

var (
	linesAddedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))  // bright green
	linesRemovedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // bright red
)

// Lines renders the lines added and removed during the current session.
// Format: "+N -M" with green for additions and red for removals.
// Returns "" when both counts are zero or no cost data was provided.
func Lines(ctx *model.RenderContext, cfg *config.Config) string {
	if ctx.LinesAdded == 0 && ctx.LinesRemoved == 0 {
		return ""
	}

	var parts []string

	if ctx.LinesAdded > 0 {
		parts = append(parts, linesAddedStyle.Render(fmt.Sprintf("+%d", ctx.LinesAdded)))
	}
	if ctx.LinesRemoved > 0 {
		parts = append(parts, linesRemovedStyle.Render(fmt.Sprintf("-%d", ctx.LinesRemoved)))
	}

	return strings.Join(parts, " ")
}
