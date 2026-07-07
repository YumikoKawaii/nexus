package transform

import (
	"encoding/json"
	"strconv"
	"time"
)

func attrsToJSON(kvs []OTLPKv) string {
	m := make(map[string]any, len(kvs))
	for _, kv := range kvs {
		m[kv.Key] = anyVal(kv.Value)
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func anyVal(v OTLPAny) any {
	switch {
	case v.StringValue != nil:
		return *v.StringValue
	case v.IntValue != nil:
		return int64(*v.IntValue)
	case v.DoubleValue != nil:
		return *v.DoubleValue
	case v.BoolValue != nil:
		return *v.BoolValue
	default:
		return nil
	}
}

func anyValStr(v OTLPAny) string {
	switch {
	case v.StringValue != nil:
		return *v.StringValue
	case v.IntValue != nil:
		return strconv.FormatInt(int64(*v.IntValue), 10)
	case v.DoubleValue != nil:
		return strconv.FormatFloat(*v.DoubleValue, 'f', -1, 64)
	case v.BoolValue != nil:
		return strconv.FormatBool(*v.BoolValue)
	default:
		return ""
	}
}

func serviceNameFrom(attrs []OTLPKv) string {
	return attrValue(attrs, "service.name")
}

func attrValue(attrs []OTLPKv, key string) string {
	for _, kv := range attrs {
		if kv.Key == key {
			return anyValStr(kv.Value)
		}
	}
	return ""
}

func marshalJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func coalesceStr(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return "0"
}

func nanoToDatetime(nanoStr string) string {
	if nanoStr == "" || nanoStr == "0" {
		return "1970-01-01 00:00:00"
	}
	ns, err := strconv.ParseInt(nanoStr, 10, 64)
	if err != nil || ns == 0 {
		return "1970-01-01 00:00:00"
	}
	sec := ns / 1_000_000_000
	return time.Unix(sec, 0).UTC().Format("2006-01-02 15:04:05")
}

// nanoToDatetimeNullable returns empty string (→ NULL) when nanoStr is absent/zero.
func nanoToDatetimeNullable(nanoStr string) string {
	if nanoStr == "" || nanoStr == "0" {
		return ""
	}
	ns, err := strconv.ParseInt(nanoStr, 10, 64)
	if err != nil || ns == 0 {
		return ""
	}
	sec := ns / 1_000_000_000
	return time.Unix(sec, 0).UTC().Format("2006-01-02 15:04:05")
}

func float64OrZero(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
