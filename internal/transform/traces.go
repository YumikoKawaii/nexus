package transform

import (
	"strconv"

	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
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

func Traces(payload *tracepb.TracesData) ([]FlatTrace, error) {
	var out []FlatTrace
	for _, rs := range payload.GetResourceSpans() {
		svc := serviceNameFrom(rs.GetResource().GetAttributes())
		resAttrs := attrsToJSON(rs.GetResource().GetAttributes())

		for _, ss := range rs.GetScopeSpans() {
			scope := ss.GetScope()
			for _, span := range ss.GetSpans() {
				kind := int32(span.GetKind())
				spanKind := spanKindNames[kind]
				if spanKind == "" {
					spanKind = strconv.Itoa(int(kind))
				}
				code := int32(span.GetStatus().GetCode())
				statusCode := statusCodeNames[code]
				if statusCode == "" {
					statusCode = strconv.Itoa(int(code))
				}

				out = append(out, FlatTrace{
					ServiceName:        svc,
					SpanName:           span.GetName(),
					TimeUnix:           nanoToString(coalesceNano(span.GetStartTimeUnixNano(), span.GetEndTimeUnixNano())),
					TraceId:            hexID(span.GetTraceId()),
					SpanId:             hexID(span.GetSpanId()),
					ParentSpanId:       hexID(span.GetParentSpanId()),
					TraceState:         span.GetTraceState(),
					SpanKind:           spanKind,
					ResourceAttributes: resAttrs,
					ScopeName:          scope.GetName(),
					ScopeVersion:       scope.GetVersion(),
					SpanAttributes:     attrsToJSON(span.GetAttributes()),
					Duration:           int64(span.GetEndTimeUnixNano()) - int64(span.GetStartTimeUnixNano()),
					StatusCode:         statusCode,
					StatusMessage:      span.GetStatus().GetMessage(),
					Events:             marshalJSON(spanEvents(span.GetEvents())),
					Links:              marshalJSON(spanLinks(span.GetLinks())),
				})
			}
		}
	}
	return out, nil
}

type outEvent struct {
	TimeUnixNano string         `json:"timeUnixNano"`
	Name         string         `json:"name"`
	Attributes   map[string]any `json:"attributes"`
}

type outLink struct {
	TraceId    string         `json:"traceId"`
	SpanId     string         `json:"spanId"`
	Attributes map[string]any `json:"attributes"`
}

func spanEvents(events []*tracepb.Span_Event) []outEvent {
	out := make([]outEvent, 0, len(events))
	for _, e := range events {
		out = append(out, outEvent{
			TimeUnixNano: nanoToString(e.GetTimeUnixNano()),
			Name:         e.GetName(),
			Attributes:   attrsToMap(e.GetAttributes()),
		})
	}
	return out
}

func spanLinks(links []*tracepb.Span_Link) []outLink {
	out := make([]outLink, 0, len(links))
	for _, l := range links {
		out = append(out, outLink{
			TraceId:    hexID(l.GetTraceId()),
			SpanId:     hexID(l.GetSpanId()),
			Attributes: attrsToMap(l.GetAttributes()),
		})
	}
	return out
}
