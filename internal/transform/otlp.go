package transform

import (
	"encoding/hex"
	"encoding/json"
	"strconv"

	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
)

func attrsToJSON(kvs []*commonpb.KeyValue) string {
	m := make(map[string]any, len(kvs))
	for _, kv := range kvs {
		m[kv.GetKey()] = anyVal(kv.GetValue())
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func attrsToMap(kvs []*commonpb.KeyValue) map[string]any {
	m := make(map[string]any, len(kvs))
	for _, kv := range kvs {
		m[kv.GetKey()] = anyVal(kv.GetValue())
	}
	return m
}

func anyVal(v *commonpb.AnyValue) any {
	if v == nil {
		return nil
	}
	switch val := v.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		return val.StringValue
	case *commonpb.AnyValue_IntValue:
		return val.IntValue
	case *commonpb.AnyValue_DoubleValue:
		return val.DoubleValue
	case *commonpb.AnyValue_BoolValue:
		return val.BoolValue
	case *commonpb.AnyValue_BytesValue:
		return hex.EncodeToString(val.BytesValue)
	case *commonpb.AnyValue_ArrayValue:
		out := make([]any, 0, len(val.ArrayValue.GetValues()))
		for _, e := range val.ArrayValue.GetValues() {
			out = append(out, anyVal(e))
		}
		return out
	case *commonpb.AnyValue_KvlistValue:
		m := make(map[string]any, len(val.KvlistValue.GetValues()))
		for _, e := range val.KvlistValue.GetValues() {
			m[e.GetKey()] = anyVal(e.GetValue())
		}
		return m
	default:
		return nil
	}
}

func anyValStr(v *commonpb.AnyValue) string {
	if v == nil {
		return ""
	}
	switch val := v.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		return val.StringValue
	case *commonpb.AnyValue_IntValue:
		return strconv.FormatInt(val.IntValue, 10)
	case *commonpb.AnyValue_DoubleValue:
		return strconv.FormatFloat(val.DoubleValue, 'f', -1, 64)
	case *commonpb.AnyValue_BoolValue:
		return strconv.FormatBool(val.BoolValue)
	default:
		return ""
	}
}

func serviceNameFrom(attrs []*commonpb.KeyValue) string {
	return attrValue(attrs, "service.name")
}

func attrValue(attrs []*commonpb.KeyValue, key string) string {
	for _, kv := range attrs {
		if kv.GetKey() == key {
			return anyValStr(kv.GetValue())
		}
	}
	return ""
}

func marshalJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func hexID(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return hex.EncodeToString(b)
}

func coalesceNano(vals ...uint64) uint64 {
	for _, v := range vals {
		if v != 0 {
			return v
		}
	}
	return 0
}

func nanoToString(ns uint64) string {
	return strconv.FormatUint(ns, 10)
}
