package transform

import (
	"encoding/json"
	"fmt"
)

func Logs(raw []byte) ([]FlatLog, error) {
	var payload OTLPLogsPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("logs unmarshal: %w", err)
	}

	var out []FlatLog
	for _, rl := range payload.ResourceLogs {
		svc := serviceNameFrom(rl.Resource.Attributes)
		resAttrs := attrsToJSON(rl.Resource.Attributes)
		svcVersion := attrValue(rl.Resource.Attributes, "service.version")
		deployEnv := attrValue(rl.Resource.Attributes, "deployment.environment")

		for _, sl := range rl.ScopeLogs {
			for _, rec := range sl.LogRecords {
				bodyBytes, _ := json.Marshal(rec.Body)
				out = append(out, FlatLog{
					ServiceName:           svc,
					Timestamp:             nanoToDatetime(coalesceStr(rec.TimeUnixNano)),
					TraceId:               rec.TraceId,
					SpanId:                rec.SpanId,
					SeverityText:          rec.SeverityText,
					SeverityNumber:        rec.SeverityNumber,
					Body:                  string(bodyBytes),
					ScopeName:             sl.Scope.Name,
					ServiceVersion:        svcVersion,
					DeploymentEnvironment: deployEnv,
					ResourceAttributes:    resAttrs,
					LogAttributes:         attrsToJSON(rec.Attributes),
					EventName:             attrValue(rec.Attributes, "event.name"),
				})
			}
		}
	}
	return out, nil
}
