package otel

import (
	"fmt"
	"reflect"
	"strings"

	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
)

func sliceToArrayValue(v reflect.Value) *commonv1.AnyValue {
	values := make([]*commonv1.AnyValue, v.Len())
	for i := 0; i < v.Len(); i++ {
		values[i] = toAnyValue(v.Index(i).Interface())
	}
	return &commonv1.AnyValue{Value: &commonv1.AnyValue_ArrayValue{ArrayValue: &commonv1.ArrayValue{Values: values}}}
}

func toAnyValue(v any) *commonv1.AnyValue {
	switch val := v.(type) {
	case string:
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: strings.ToValidUTF8(val, "")}}
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
	case map[string]any:
		kvs := make([]*commonv1.KeyValue, 0, len(val))
		for k, v := range val {
			kvs = append(kvs, &commonv1.KeyValue{Key: k, Value: toAnyValue(v)})
		}
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_KvlistValue{KvlistValue: &commonv1.KeyValueList{Values: kvs}}}
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice {
			return sliceToArrayValue(rv)
		}
		return &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: strings.ToValidUTF8(fmt.Sprintf("%v", v), "")}}
	}
}

func anyValueToGo(v *commonv1.AnyValue) any {
	if v == nil {
		return nil
	}
	switch val := v.Value.(type) {
	case *commonv1.AnyValue_StringValue:
		return val.StringValue
	case *commonv1.AnyValue_IntValue:
		return val.IntValue
	case *commonv1.AnyValue_DoubleValue:
		return val.DoubleValue
	case *commonv1.AnyValue_BoolValue:
		return val.BoolValue
	default:
		return fmt.Sprintf("%v", v)
	}
}
