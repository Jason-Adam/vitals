package widget

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

var (
	dimStyle    = lipgloss.NewStyle().Faint(true)
	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	yellowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// Context renders a filled/empty progress bar representing context window usage,
// followed by the percentage value. The bar color shifts from green to yellow at
// 70% and from yellow to red at 85%.
// Returns "" when both ContextPercent and ContextWindowSize are zero.
func Context(ctx *model.RenderContext, cfg *config.Config) string {
	if ctx.ContextPercent == 0 && ctx.ContextWindowSize == 0 {
		return ""
	}

	barWidth := cfg.Context.BarWidth
	if barWidth <= 0 {
		barWidth = 10
	}

	pct := ctx.ContextPercent

	// Select color based on usage thresholds.
	colorStyle := greenStyle
	if pct >= 85 {
		colorStyle = redStyle
	} else if pct >= 70 {
		colorStyle = yellowStyle
	}

	filled := (pct * barWidth) / 100
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	bar := colorStyle.Render(strings.Repeat("█", filled)) +
		dimStyle.Render(strings.Repeat("░", empty))

	return bar + " " + colorStyle.Render(fmt.Sprintf("%d%%", pct))
}
