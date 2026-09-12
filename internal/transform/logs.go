package transform

import (
	"encoding/json"

	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
)

func Logs(payload *logspb.LogsData) ([]FlatLog, error) {
	var out []FlatLog
	for _, rl := range payload.GetResourceLogs() {
		resAttrs := rl.GetResource().GetAttributes()
		svc := serviceNameFrom(resAttrs)
		resAttrsJSON := attrsToJSON(resAttrs)
		svcVersion := attrValue(resAttrs, "service.version")
		deployEnv := attrValue(resAttrs, "deployment.environment")

		for _, sl := range rl.GetScopeLogs() {
			scope := sl.GetScope()
			for _, rec := range sl.GetLogRecords() {
				bodyBytes, _ := json.Marshal(anyVal(rec.GetBody()))
				out = append(out, FlatLog{
					ServiceName:           svc,
					Timestamp:             nanoToDatetime(coalesceNano(rec.GetTimeUnixNano(), rec.GetObservedTimeUnixNano())),
					TraceId:               hexID(rec.GetTraceId()),
					SpanId:                hexID(rec.GetSpanId()),
					SeverityText:          rec.GetSeverityText(),
					SeverityNumber:        int32(rec.GetSeverityNumber()),
					Body:                  string(bodyBytes),
					ScopeName:             scope.GetName(),
					ServiceVersion:        svcVersion,
					DeploymentEnvironment: deployEnv,
					ResourceAttributes:    resAttrsJSON,
					LogAttributes:         attrsToJSON(rec.GetAttributes()),
					EventName:             attrValue(rec.GetAttributes(), "event.name"),
				})
			}
		}
	}
	return out, nil
}
