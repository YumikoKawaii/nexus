package transform

import (
	"encoding/json"
	"fmt"
	"strconv"
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
		return *v.IntValue
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
		return strconv.FormatInt(*v.IntValue, 10)
	case v.DoubleValue != nil:
		return strconv.FormatFloat(*v.DoubleValue, 'f', -1, 64)
	case v.BoolValue != nil:
		return strconv.FormatBool(*v.BoolValue)
	default:
		return ""
	}
}

func serviceNameFrom(attrs []OTLPKv) string {
	for _, kv := range attrs {
		if kv.Key == "service.name" {
			return anyValStr(kv.Value)
		}
	}
	return ""
}

func nanoToTime(nanoStr string) string {
	if nanoStr == "" {
		return ""
	}
	ns, err := strconv.ParseInt(nanoStr, 10, 64)
	if err != nil {
		return nanoStr
	}
	sec := ns / 1_000_000_000
	nano := ns % 1_000_000_000
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d.%09d",
		1970+sec/31557600, 1, 1, 0, 0, sec%60, nano) // rough; SR accepts epoch strings too
}

func marshalJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func float64OrZero(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}
