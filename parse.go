package otel

import (
	"encoding/hex"
	"fmt"
	"os"

	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	"golang.org/x/term"
)

// PrintServiceLog parses OTLP JSON and prints pretty output if in TTY.
// If not in TTY or not valid OTLP JSON, prints line as-is.
func PrintServiceLog(line string) {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Println(line)
		return
	}

	var data logsv1.LogsData
	if err := unmarshaler.Unmarshal([]byte(line), &data); err != nil {
		fmt.Println(line)
		return
	}

	if len(data.ResourceLogs) == 0 || len(data.ResourceLogs[0].ScopeLogs) == 0 {
		fmt.Println(line)
		return
	}

	svcName := "unknown"
	for _, attr := range data.ResourceLogs[0].Resource.Attributes {
		if attr.Key == "service.name" {
			svcName = attr.Value.GetStringValue()
			break
		}
	}

	for _, scopeLog := range data.ResourceLogs[0].ScopeLogs {
		for _, record := range scopeLog.LogRecords {
			var attrs []Attr
			for _, kv := range record.Attributes {
				attrs = append(attrs, Attr{K: kv.Key, V: anyValueToGo(kv.Value)})
			}

			var traceID, spanID string
			if len(record.TraceId) > 0 {
				traceID = hex.EncodeToString(record.TraceId)
			}
			if len(record.SpanId) > 0 {
				spanID = hex.EncodeToString(record.SpanId)
			}

			printPrettyLine(svcName, record.SeverityText, record.Body.GetStringValue(), attrs, traceID, spanID)
		}
	}
}
