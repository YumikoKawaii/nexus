package transform

import (
	"encoding/json"
	"fmt"
	"strconv"
)

var spanKindNames = map[int32]string{
	0: "UNSPECIFIED",
	1: "INTERNAL",
	2: "SERVER",
	3: "CLIENT",
	4: "PRODUCER",
	5: "CONSUMER",
}

var statusCodeNames = map[int32]string{
	0: "STATUS_CODE_UNSET",
	1: "STATUS_CODE_OK",
	2: "STATUS_CODE_ERROR",
}

func Traces(raw []byte) ([]FlatTrace, error) {
	var payload OTLPTracePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("traces unmarshal: %w", err)
	}

	var out []FlatTrace
	for _, rs := range payload.ResourceSpans {
		svc := serviceNameFrom(rs.Resource.Attributes)
		resAttrs := attrsToJSON(rs.Resource.Attributes)

		for _, ss := range rs.ScopeSpans {
			for _, span := range ss.Spans {
				startNs, _ := strconv.ParseInt(span.StartTimeUnixNano, 10, 64)
				endNs, _ := strconv.ParseInt(span.EndTimeUnixNano, 10, 64)

				spanKind := spanKindNames[span.Kind]
				if spanKind == "" {
					spanKind = strconv.Itoa(int(span.Kind))
				}
				statusCode := statusCodeNames[span.Status.Code]
				if statusCode == "" {
					statusCode = strconv.Itoa(int(span.Status.Code))
				}

				out = append(out, FlatTrace{
					ServiceName:        svc,
					SpanName:           span.Name,
					Timestamp:          nanoToDatetime(coalesceStr(span.StartTimeUnixNano, span.EndTimeUnixNano)),
					TraceId:            span.TraceId,
					SpanId:             span.SpanId,
					ParentSpanId:       span.ParentSpanId,
					TraceState:         span.TraceState,
					SpanKind:           spanKind,
					ResourceAttributes: resAttrs,
					ScopeName:          ss.Scope.Name,
					ScopeVersion:       ss.Scope.Version,
					SpanAttributes:     attrsToJSON(span.Attributes),
					Duration:           endNs - startNs,
					StatusCode:         statusCode,
					StatusMessage:      span.Status.Message,
					Events:             marshalJSON(span.Events),
					Links:              marshalJSON(span.Links),
				})
			}
		}
	}
	return out, nil
}
