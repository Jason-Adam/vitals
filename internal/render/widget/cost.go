package widget

import (
	"fmt"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Cost renders the session cost as a dollar amount. The color shifts from the
// normal context color to warning at cfg.Thresholds.CostWarning USD, and to
// critical at cfg.Thresholds.CostCritical USD.
//
// Returns an empty WidgetResult when SessionCostUSD is zero (no cost data available).
// FgColor is left empty because the widget selects among multiple styles dynamically;
// the renderer passes the pre-styled Text through as-is.
func Cost(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	if ctx.SessionCostUSD == 0 {
		return WidgetResult{}
	}

	// Resolve colors: prefer config overrides, fall back to package-level defaults.
	contextColor := colorStyle(cfg.Style.Colors.Context, greenStyle)
	warningColor := colorStyle(cfg.Style.Colors.Warning, yellowStyle)
	criticalColor := colorStyle(cfg.Style.Colors.Critical, redStyle)

	// Resolve thresholds with safe fallbacks.
	warnAt := cfg.Thresholds.CostWarning
	critAt := cfg.Thresholds.CostCritical
	if warnAt <= 0 {
		warnAt = 5.00
	}
	if critAt <= 0 {
		critAt = 10.00
	}

	cost := ctx.SessionCostUSD
	activeStyle := contextColor
	if cost >= critAt {
		activeStyle = criticalColor
	} else if cost >= warnAt {
		activeStyle = warningColor
	}

	return WidgetResult{Text: activeStyle.Render(fmt.Sprintf("$%.2f", cost))}
}
