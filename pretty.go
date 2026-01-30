package otel

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/term"
)

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func wrapLine(s string, width int, indent string) string {
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}

	var result strings.Builder
	for len(runes) > width {
		result.WriteString(string(runes[:width]))
		result.WriteString("\n")
		result.WriteString(indent)
		runes = runes[width:]
	}
	result.WriteString(string(runes))
	return result.String()
}

func getTerminalWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return 120
}

func printPrettyLine(source, severityText, body string, attrs []Attr, traceID, spanID string) {
	now := time.Now().Format("15:04:05")
	width := getTerminalWidth()

	fmt.Fprintf(os.Stdout, "%s [%-*s] %-5s %s\n", now, serviceNameWidth, source, severityText, body)

	// 8 (time) + 1 (space) + 1 ([) + serviceNameWidth + 1 (]) + 1 (space) + 5 (severity) + 1 (space)
	indent := 18 + serviceNameWidth
	prefix := strings.Repeat(" ", indent)
	maxValueWidth := width - indent

	if traceID != "" {
		fmt.Fprintf(os.Stdout, "%strace=%s\n", prefix, traceID)
	}
	if spanID != "" {
		fmt.Fprintf(os.Stdout, "%sspan=%s\n", prefix, spanID)
	}

	for _, attr := range attrs {
		line := fmt.Sprintf("%s=%s", attr.K, formatValue(attr.V))
		if len([]rune(line)) > maxValueWidth && maxValueWidth > 20 {
			line = wrapLine(line, maxValueWidth, prefix)
		}
		fmt.Fprintf(os.Stdout, "%s%s\n", prefix, line)
	}
}

func emitPretty(ctx context.Context, severityText string, body string, attrs []Attr) {
	source := serviceName
	var filteredAttrs []Attr
	for _, attr := range attrs {
		if attr.K == "source" {
			source = fmt.Sprintf("%v", attr.V)
		} else {
			filteredAttrs = append(filteredAttrs, attr)
		}
	}

	var traceID, spanID string
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		traceID = spanCtx.TraceID().String()
		spanID = spanCtx.SpanID().String()
	}

	printPrettyLine(source, severityText, body, filteredAttrs, traceID, spanID)
}
