# pink-otel

Structured logging for Go following [OTLP Logs](https://opentelemetry.io/docs/specs/otlp/) specification. Used in all [pink-tools](https://github.com/pink-tools) services.

Output format: [LogsData](https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto) protobuf serialized to JSON.

## Install

```bash
go get github.com/pink-tools/pink-otel
```

## Usage

```go
package main

import (
    "context"
    "github.com/pink-tools/pink-otel"
)

func main() {
    otel.Init("my-service", "1.0.0")

    ctx := context.Background()
    otel.Info(ctx, "server started", otel.Attr{K: "port", V: 8080})
    otel.Error(ctx, "request failed",
        otel.Attr{K: "status", V: 500},
        otel.Attr{K: "path", V: "/api/users"},
    )
}
```

## Output

Auto-detects terminal vs file/pipe:

| Running | Format |
|---------|--------|
| `./myapp` in terminal | pretty |
| `./myapp > app.log` | JSON |
| `./myapp \| jq` | JSON |
| Docker / systemd | JSON |

Override: `LOG_FORMAT=json` or `LOG_FORMAT=pretty`.

**Pretty (terminal):**
```
15:04:05 [my-service] INFO  server started [port=8080]
15:04:05 [my-service] ERROR request failed [status=500, path=/api/users]
```

**JSON (file/pipe):**
```json
{
  "resourceLogs": [{
    "resource": {
      "attributes": [
        {"key": "service.name", "value": {"stringValue": "my-service"}},
        {"key": "service.version", "value": {"stringValue": "1.0.0"}}
      ]
    },
    "scopeLogs": [{
      "logRecords": [{
        "severityText": "ERROR",
        "body": {"stringValue": "request failed"},
        "attributes": [
          {"key": "status", "value": {"intValue": "500"}}
        ]
      }]
    }]
  }]
}
```

TraceID/SpanID included when context contains OpenTelemetry span.

## API

```go
func Init(name, version string)
func Debug(ctx context.Context, body string, attrs ...Attr)
func Info(ctx context.Context, body string, attrs ...Attr)
func Warn(ctx context.Context, body string, attrs ...Attr)
func Error(ctx context.Context, body string, attrs ...Attr)

// Ordered key-value attribute
type Attr struct { K string; V any }

// Log parsing (for service log forwarding)
func ParseLogMessage(line []byte) string
func PrintServiceLog(line []byte)
```

## Attribute Types

| Go | OTLP |
|----|------|
| string | stringValue |
| int, int8-64, uint, uint8-64 | intValue |
| float32, float64 | doubleValue |
| bool | boolValue |
| []byte | bytesValue |
| []any, []string, []int | arrayValue |
| map[string]any | kvlistValue |

## Spec References

- [OTLP Protocol](https://opentelemetry.io/docs/specs/otlp/)
- [Logs Data Model](https://opentelemetry.io/docs/specs/otel/logs/data-model/)
- [LogsData Proto](https://github.com/open-telemetry/opentelemetry-proto/blob/main/opentelemetry/proto/logs/v1/logs.proto)
- [Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/)
