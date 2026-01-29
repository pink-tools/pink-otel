# pink-otel

Structured logging for Go following [OTLP Logs](https://opentelemetry.io/docs/specs/otlp/#otlpgrpc-protocol) specification. Used in all [pink-tools](https://github.com/pink-tools) services.

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
    otel.Info(ctx, "server started")
    otel.Error(ctx, "request failed", map[string]any{
        "status":  500,
        "path":    "/api/users",
        "user_id": 12345,
    })
}
```

## Output

Checks if stdout is a terminal (`isatty`):

| Running | stdout | Format |
|---------|--------|--------|
| `./myapp` in terminal | terminal | pretty |
| `./myapp > app.log` | file | JSON |
| `./myapp \| jq` | pipe | JSON |
| Docker container | not a terminal | JSON |
| systemd service | journal | JSON |
| AWS CloudWatch | not a terminal | JSON |

Environment variable override:
```bash
LOG_FORMAT=json ./myapp   # force JSON in terminal
LOG_FORMAT=pretty ./myapp # force pretty in Docker
```

**Pretty (terminal):**
```
15:04:05 [my-service] INFO  server started
15:04:05 [my-service] ERROR request failed [status=500, path=/api/users, user_id=12345]
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
        "timeUnixNano": 1706234567000000000,
        "severityNumber": 17,
        "severityText": "ERROR",
        "body": {"stringValue": "request failed"},
        "attributes": [
          {"key": "status", "value": {"intValue": "500"}}
        ],
        "traceId": "...",
        "spanId": "..."
      }]
    }]
  }]
}
```

TraceID/SpanID included when context contains OpenTelemetry span.

## API

```go
func Init(name, version string)
func Debug(ctx context.Context, body string, attrs ...map[string]any)
func Info(ctx context.Context, body string, attrs ...map[string]any)
func Warn(ctx context.Context, body string, attrs ...map[string]any)
func Error(ctx context.Context, body string, attrs ...map[string]any)
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
