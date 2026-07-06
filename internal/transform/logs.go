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

		for _, sl := range rl.ScopeLogs {
			for _, rec := range sl.LogRecords {
				bodyBytes, _ := json.Marshal(rec.Body)
				out = append(out, FlatLog{
					ServiceName:        svc,
					Timestamp:          rec.TimeUnixNano,
					SeverityNumber:     rec.SeverityNumber,
					SeverityText:       rec.SeverityText,
					Body:               string(bodyBytes),
					ResourceAttributes: resAttrs,
					LogAttributes:      attrsToJSON(rec.Attributes),
					TraceId:            rec.TraceId,
					SpanId:             rec.SpanId,
				})
			}
		}
	}
	return out, nil
}
