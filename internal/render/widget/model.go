package widget

import (
	"fmt"

	"charm.land/lipgloss/v2"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

var cyanStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("87"))

// Model renders the model display name in cyan, optionally suffixed with the
// context window size when cfg.Model.ShowContextSize is true.
// Returns "" when ctx.ModelDisplayName is empty.
func Model(ctx *model.RenderContext, cfg *config.Config) string {
	if ctx.ModelDisplayName == "" {
		return ""
	}

	name := ctx.ModelDisplayName

	if cfg.Model.ShowContextSize && ctx.ContextWindowSize > 0 {
		name = fmt.Sprintf("%s (%s context)", name, formatTokens(ctx.ContextWindowSize))
	}

	return cyanStyle.Render(fmt.Sprintf("[%s]", name))
}

// formatTokens converts a raw token count to a human-readable string.
func formatTokens(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.0fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%dk", n/1_000)
	}
	return fmt.Sprintf("%d", n)
}
