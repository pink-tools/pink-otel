package otel

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	serviceName    string
	serviceVersion string
)

var marshaler = protojson.MarshalOptions{EmitUnpopulated: false}

func Init(name, version string) {
	serviceName = name
	serviceVersion = version
}

func toAnyValue(v any) *commonv1.AnyValue {
	switch val := v.(type) {
	case string:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: val}}
	case bool:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_BoolValue{BoolValue: val}}
	case int:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: int64(val)}}
	case int8:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: int64(val)}}
	case int16:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: int64(val)}}
	case int32:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: int64(val)}}
	case int64:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: val}}
	case uint:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: int64(val)}}
	case uint8:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: int64(val)}}
	case uint16:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: int64(val)}}
	case uint32:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: int64(val)}}
	case uint64:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_IntValue{IntValue: int64(val)}}
	case float32:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_DoubleValue{DoubleValue: float64(val)}}
	case float64:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_DoubleValue{DoubleValue: val}}
	case []byte:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_BytesValue{BytesValue: val}}
	case []any:
		values := make([]*commonv1.AnyValue, len(val))
		for i, item := range val {
			values[i] = toAnyValue(item)
		}
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_ArrayValue{ArrayValue: &commonv1.ArrayValue{Values: values}}}
	case []string:
		values := make([]*commonv1.AnyValue, len(val))
		for i, item := range val {
			values[i] = toAnyValue(item)
		}
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_ArrayValue{ArrayValue: &commonv1.ArrayValue{Values: values}}}
	case []int:
		values := make([]*commonv1.AnyValue, len(val))
		for i, item := range val {
			values[i] = toAnyValue(item)
		}
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_ArrayValue{ArrayValue: &commonv1.ArrayValue{Values: values}}}
	case map[string]any:
		kvs := make([]*commonv1.KeyValue, 0, len(val))
		for k, v := range val {
			kvs = append(kvs, &commonv1.KeyValue{Key: k, Value: toAnyValue(v)})
		}
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_KvlistValue{KvlistValue: &commonv1.KeyValueList{Values: kvs}}}
	default:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: fmt.Sprintf("%v", v)}}
	}
}

func emit(ctx context.Context, severityNumber logsv1.SeverityNumber, severityText string, body string, attrs map[string]any) {
	now := uint64(time.Now().UnixNano())

	var kvAttrs []*commonv1.KeyValue
	for k, v := range attrs {
		kvAttrs = append(kvAttrs, &commonv1.KeyValue{
			Key:   k,
			Value: toAnyValue(v),
		})
	}

	logRecord := &logsv1.LogRecord{
		TimeUnixNano:         now,
		ObservedTimeUnixNano: now,
		SeverityNumber:       severityNumber,
		SeverityText:         severityText,
		Body:                 &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: body}},
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

	jsonBytes, _ := marshaler.Marshal(data)
	os.Stdout.Write(jsonBytes)
	os.Stdout.Write([]byte("\n"))
}

func Debug(ctx context.Context, body string, attrs ...map[string]any) {
	var a map[string]any
	if len(attrs) > 0 {
		a = attrs[0]
	}
	emit(ctx, logsv1.SeverityNumber_SEVERITY_NUMBER_DEBUG, "DEBUG", body, a)
}

func Info(ctx context.Context, body string, attrs ...map[string]any) {
	var a map[string]any
	if len(attrs) > 0 {
		a = attrs[0]
	}
	emit(ctx, logsv1.SeverityNumber_SEVERITY_NUMBER_INFO, "INFO", body, a)
}

func Warn(ctx context.Context, body string, attrs ...map[string]any) {
	var a map[string]any
	if len(attrs) > 0 {
		a = attrs[0]
	}
	emit(ctx, logsv1.SeverityNumber_SEVERITY_NUMBER_WARN, "WARN", body, a)
}

func Error(ctx context.Context, body string, attrs ...map[string]any) {
	var a map[string]any
	if len(attrs) > 0 {
		a = attrs[0]
	}
	emit(ctx, logsv1.SeverityNumber_SEVERITY_NUMBER_ERROR, "ERROR", body, a)
}
