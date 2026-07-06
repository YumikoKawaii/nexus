package transform

import (
	"encoding/json"
	"fmt"
	"strconv"
)

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
				durationNs := endNs - startNs

				out = append(out, FlatTrace{
					ServiceName:        svc,
					TraceId:            span.TraceId,
					SpanId:             span.SpanId,
					ParentSpanId:       span.ParentSpanId,
					SpanName:           span.Name,
					SpanKind:           span.Kind,
					StartTime:          span.StartTimeUnixNano,
					EndTime:            span.EndTimeUnixNano,
					Duration:           durationNs,
					StatusCode:         span.Status.Code,
					StatusMessage:      span.Status.Message,
					ResourceAttributes: resAttrs,
					SpanAttributes:     attrsToJSON(span.Attributes),
					Events:             marshalJSON(span.Events),
					Links:              marshalJSON(span.Links),
				})
			}
		}
	}
	return out, nil
}
