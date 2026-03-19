package widget

import (
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Usage renders 5-hour and 7-day rate-limit utilization from the Anthropic
// OAuth API.
//
// Returns empty when:
//   - ctx.Usage is nil (no credentials, API user, widget not configured)
//   - Both windows are below their configured thresholds
func Usage(ctx *model.RenderContext, cfg *config.Config) WidgetResult {
	if ctx.Usage == nil {
		return WidgetResult{}
	}

	u := ctx.Usage

	// API errors (except rate-limited with stale data) get a warning label.
	if u.APIUnavailable && u.APIError != "rate-limited" {
		return usageError(u.APIError)
	}

	// Limit reached: bold critical with reset countdown.
	if u.FiveHourPercent >= 100 || u.SevenDayPercent >= 100 {
		return usageLimitReached(u, cfg)
	}

	// Hide when below threshold.
	effectiveUsage := max(u.FiveHourPercent, u.SevenDayPercent)
	if effectiveUsage < cfg.Usage.FiveHourThreshold {
		return WidgetResult{}
	}

	// Assemble visible windows.
	var windows []usageSegment
	if u.FiveHourPercent >= 0 {
		windows = append(windows, usageWindow("5h", u.FiveHourPercent, u.FiveHourResetAt, cfg))
	}
	if u.SevenDayPercent >= 0 && u.SevenDayPercent >= cfg.Usage.SevenDayThreshold {
		windows = append(windows, usageWindow("7d", u.SevenDayPercent, u.SevenDayResetAt, cfg))
	}
	if len(windows) == 0 {
		return WidgetResult{}
	}

	// Join windows, append syncing hint, pick highest-severity color.
	return usageJoin(windows, u.APIError == "rate-limited")
}

// ---------------------------------------------------------------------------
// Segment: the intermediate representation between data and final string.
// Each helper below produces one segment; the top-level function joins them.
// ---------------------------------------------------------------------------

// usageSegment holds the styled and plain text for one window, plus its
// foreground color so the joiner can pick the most urgent.
type usageSegment struct {
	text      string
	plainText string
	fgColor   string
}

// ---------------------------------------------------------------------------
// Window assembly — composes the per-element helpers into a single segment.
// Change the order of calls here to rearrange the visual layout.
// ---------------------------------------------------------------------------

func usageWindow(label string, pct int, resetAt time.Time, cfg *config.Config) usageSegment {
	_ = label // available for re-enabling usageLabel
	if pct < 0 {
		pct = 0
	}
	fg := usageFgColor(pct, cfg)
	pctStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(fg))

	var styled, plain strings.Builder

	append := func(s, p string) { appendPair(&styled, &plain, s, p) }

	// 1. Circle-fill icon (nerdfont) or percentage fallback
	if cfg.Style.Icons == "nerdfont" {
		append(usageIcon(pct, pctStyle))
	} else {
		append(usagePercent(pct, pctStyle))
	}

	// 2. Reset countdown
	append(usageReset(resetAt))

	return usageSegment{text: styled.String(), plainText: plain.String(), fgColor: fg}
}

// ---------------------------------------------------------------------------
// Per-element helpers. Each returns (styled, plain). Return ("", "") to omit.
// ---------------------------------------------------------------------------

// usageLabel renders the window identifier: "5h" or "7d".
func usageLabel(label string) (string, string) {
	return DimStyle.Render(label), label
}

// usageIcon renders the circle-slice fill icon colored by severity.
func usageIcon(pct int, style lipgloss.Style) (string, string) {
	icon := percentToIcon(pct)
	return style.Render(icon), icon
}

// usagePercent renders " NN%", colored by severity.
func usagePercent(pct int, style lipgloss.Style) (string, string) {
	s := fmt.Sprintf(" %d%%", pct)
	return style.Render(s), s
}

// usageReset renders the reset countdown. Returns ("", "") when the reset
// time is zero or in the past.
func usageReset(resetAt time.Time) (string, string) {
	r := formatResetTime(resetAt)
	if r == "" {
		return "", ""
	}
	s := " (" + r + ")"
	return DimStyle.Render(s), s
}

// ---------------------------------------------------------------------------
// Composite results: error, limit-reached, and multi-window join.
// ---------------------------------------------------------------------------

// usageError renders the API-unavailable warning state.
func usageError(apiError string) WidgetResult {
	hint := formatUsageError(apiError)
	label := "Usage " + hint
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	return WidgetResult{
		Text:      style.Render(label),
		PlainText: label,
		FgColor:   "3",
	}
}

// usageLimitReached renders bold critical text with a reset countdown.
func usageLimitReached(u *model.UsageInfo, cfg *config.Config) WidgetResult {
	critFg := "1"
	if cfg.Style.Colors.Critical != "" {
		critFg = cfg.Style.Colors.Critical
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(critFg)).Bold(true)

	var resetAt time.Time
	if u.FiveHourPercent >= 100 {
		resetAt = u.FiveHourResetAt
	} else {
		resetAt = u.SevenDayResetAt
	}

	label := "Limit reached"
	if r := formatResetTime(resetAt); r != "" {
		label += fmt.Sprintf(" (resets %s)", r)
	}

	return WidgetResult{
		Text:      style.Render(label),
		PlainText: label,
		FgColor:   critFg,
	}
}

// usageJoin combines multiple window segments with a separator and appends
// a syncing hint when rate-limited. Picks the highest-severity foreground color.
func usageJoin(windows []usageSegment, syncing bool) WidgetResult {
	styledParts := make([]string, len(windows))
	plainParts := make([]string, len(windows))
	fgColor := ""

	for i, w := range windows {
		styledParts[i] = w.text
		plainParts[i] = w.plainText
		if usageColorPriority(w.fgColor) > usageColorPriority(fgColor) {
			fgColor = w.fgColor
		}
	}

	text := strings.Join(styledParts, DimStyle.Render(" | "))
	plain := strings.Join(plainParts, " | ")

	if syncing {
		text += " " + DimStyle.Render("(syncing...)")
		plain += " (syncing...)"
	}

	return WidgetResult{Text: text, PlainText: plain, FgColor: fgColor}
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// appendPair appends a (styled, plain) pair to the respective builders.
// Skips when both are empty, so element helpers can return ("", "") to omit.
func appendPair(styled, plain *strings.Builder, s, p string) {
	if s == "" && p == "" {
		return
	}
	styled.WriteString(s)
	plain.WriteString(p)
}

// usageFgColor returns the ANSI foreground color for a usage percentage.
// 0-49%: green, 50-79%: yellow, 80-100%: red.
func usageFgColor(pct int, cfg *config.Config) string {
	return thresholdFgColor(pct, 50, 80,
		cfg.Style.Colors.Context, cfg.Style.Colors.Warning, cfg.Style.Colors.Critical)
}

// usageColorPriority returns a numeric priority so the joiner picks the
// most urgent color when combining windows.
func usageColorPriority(fg string) int {
	switch fg {
	case "1":
		return 3 // red / critical
	case "3":
		return 2 // yellow / warning
	default:
		return 1 // green / normal
	}
}

// formatResetTime formats a future timestamp as a compact duration string.
// Returns "" when the time is zero or in the past.
func formatResetTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	diff := time.Until(t)
	if diff <= 0 {
		return ""
	}

	totalMins := int(math.Ceil(diff.Minutes()))
	if totalMins < 60 {
		return fmt.Sprintf("%dm", totalMins)
	}

	hours := totalMins / 60
	mins := totalMins % 60

	if hours >= 24 {
		days := hours / 24
		remHours := hours % 24
		if remHours > 0 {
			return fmt.Sprintf("%dd %dh", days, remHours)
		}
		return fmt.Sprintf("%dd", days)
	}

	if mins > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dh", hours)
}

// formatUsageError formats an API error string for display.
func formatUsageError(apiError string) string {
	if apiError == "" {
		return ""
	}
	if apiError == "rate-limited" {
		return "(syncing...)"
	}
	if strings.HasPrefix(apiError, "http-") {
		return "(" + apiError[5:] + ")"
	}
	return "(" + apiError + ")"
}
