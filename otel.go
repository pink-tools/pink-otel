package otel

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"golang.org/x/term"
	"google.golang.org/protobuf/encoding/protojson"
)

// Attr is an ordered key-value pair for log attributes.
// Using a slice of Attr instead of map[string]any preserves insertion order.
type Attr struct {
	K string
	V any
}

var (
	serviceName      string
	serviceVersion   string
	serviceNameWidth int
	prettyMode       bool
)

var marshaler = protojson.MarshalOptions{EmitUnpopulated: false}
var unmarshaler = protojson.UnmarshalOptions{DiscardUnknown: true}

func Init(name, version string) {
	serviceName = name
	serviceVersion = version

	serviceNameWidth = len(name)
	if w, err := strconv.Atoi(os.Getenv("PINK_LOG_WIDTH")); err == nil && w > 0 {
		serviceNameWidth = w
	}

	switch strings.ToLower(os.Getenv("LOG_FORMAT")) {
	case "json":
		prettyMode = false
	case "pretty":
		prettyMode = true
	default:
		prettyMode = term.IsTerminal(int(os.Stdout.Fd()))
	}
}

func SetServiceNameWidth(w int) {
	if w > 0 {
		serviceNameWidth = w
	}
}

func emit(ctx context.Context, severityNumber logsv1.SeverityNumber, severityText string, body string, attrs []Attr) {
	if prettyMode {
		emitPretty(ctx, severityText, body, attrs)
		return
	}

	now := uint64(time.Now().UnixNano())

	var kvAttrs []*commonv1.KeyValue
	for _, attr := range attrs {
		kvAttrs = append(kvAttrs, &commonv1.KeyValue{
			Key:   strings.ToValidUTF8(attr.K, ""),
			Value: toAnyValue(attr.V),
		})
	}

	logRecord := &logsv1.LogRecord{
		TimeUnixNano:         now,
		ObservedTimeUnixNano: now,
		SeverityNumber:       severityNumber,
		SeverityText:         severityText,
		Body:                 &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: strings.ToValidUTF8(body, "")}},
		Attributes:           kvAttrs,
	}

	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		traceID := spanCtx.TraceID()
		spanID := spanCtx.SpanID()
		logRecord.TraceId = traceID[:]
		logRecord.SpanId = spanID[:]
		logRecord.Flags = uint32(spanCtx.TraceFlags())
	}

	data := &logsv1.LogsData{
		ResourceLogs: []*logsv1.ResourceLogs{{
			Resource: &resourcev1.Resource{
				Attributes: []*commonv1.KeyValue{
					{Key: "service.name", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: serviceName}}},
					{Key: "service.version", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: serviceVersion}}},
					{Key: "telemetry.sdk.name", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "pink-otel"}}},
					{Key: "telemetry.sdk.language", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "go"}}},
					{Key: "telemetry.sdk.version", Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: Version}}},
				},
			},
			ScopeLogs: []*logsv1.ScopeLogs{{
				Scope: &commonv1.InstrumentationScope{
					Name:    "github.com/pink-tools/pink-otel",
					Version: Version,
				},
				LogRecords: []*logsv1.LogRecord{logRecord},
			}},
		}},
	}

	jsonBytes, err := marshaler.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "otel: marshal error: %v\n", err)
		return
	}
	os.Stdout.Write(jsonBytes)
	os.Stdout.Write([]byte("\n"))
}

func Debug(ctx context.Context, body string, attrs ...Attr) {
	emit(ctx, logsv1.SeverityNumber_SEVERITY_NUMBER_DEBUG, "DEBUG", body, attrs)
}

func Info(ctx context.Context, body string, attrs ...Attr) {
	emit(ctx, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, "INFO", body, attrs)
}

func Warn(ctx context.Context, body string, attrs ...Attr) {
	emit(ctx, logsv1.SeverityNumber_SEVERITY_NUMBER_WARN, "WARN", body, attrs)
}

func Error(ctx context.Context, body string, attrs ...Attr) {
	emit(ctx, logsv1.SeverityNumber_SEVERITY_NUMBER_ERROR, "ERROR", body, attrs)
}
