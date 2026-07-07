package transform

import (
	"encoding/json"
	"testing"
)

// --- StringInt64 ---

func TestStringInt64_Quoted(t *testing.T) {
	var v StringInt64
	if err := json.Unmarshal([]byte(`"9007199254740993"`), &v); err != nil {
		t.Fatal(err)
	}
	if int64(v) != 9007199254740993 {
		t.Fatalf("got %d", v)
	}
}

func TestStringInt64_Bare(t *testing.T) {
	var v StringInt64
	if err := json.Unmarshal([]byte(`-42`), &v); err != nil {
		t.Fatal(err)
	}
	if int64(v) != -42 {
		t.Fatalf("got %d", v)
	}
}

// --- StringUint64 ---

func TestStringUint64_Quoted(t *testing.T) {
	var v StringUint64
	if err := json.Unmarshal([]byte(`"18446744073709551615"`), &v); err != nil {
		t.Fatal(err)
	}
	if uint64(v) != 18446744073709551615 {
		t.Fatalf("got %d", v)
	}
}

func TestStringUint64_Bare(t *testing.T) {
	var v StringUint64
	if err := json.Unmarshal([]byte(`100`), &v); err != nil {
		t.Fatal(err)
	}
	if uint64(v) != 100 {
		t.Fatalf("got %d", v)
	}
}

// --- StringUint64Slice ---

func TestStringUint64Slice_Mixed(t *testing.T) {
	var s StringUint64Slice
	if err := json.Unmarshal([]byte(`["1","2",3]`), &s); err != nil {
		t.Fatal(err)
	}
	want := []uint64{1, 2, 3}
	for i, w := range want {
		if s[i] != w {
			t.Fatalf("index %d: got %d want %d", i, s[i], w)
		}
	}
}

// --- nanoToDatetime ---

func TestNanoToDatetime_Normal(t *testing.T) {
	// 1720000000000000000 ns = 1720000000 s = 2024-07-03 07:06:40 UTC
	got := nanoToDatetime("1720000000000000000")
	if got != "2024-07-03 09:46:40" {
		t.Fatalf("got %q", got)
	}
}

func TestNanoToDatetime_Empty(t *testing.T) {
	got := nanoToDatetime("")
	if got != "1970-01-01 00:00:00" {
		t.Fatalf("got %q", got)
	}
}

func TestNanoToDatetime_Zero(t *testing.T) {
	got := nanoToDatetime("0")
	if got != "1970-01-01 00:00:00" {
		t.Fatalf("got %q", got)
	}
}

// --- nanoToDatetimeNullable ---

func TestNanoToDatetimeNullable_Empty(t *testing.T) {
	got := nanoToDatetimeNullable("")
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}

func TestNanoToDatetimeNullable_Normal(t *testing.T) {
	got := nanoToDatetimeNullable("1720000000000000000")
	if got != "2024-07-03 09:46:40" {
		t.Fatalf("got %q", got)
	}
}

// --- Metrics end-to-end ---

func TestMetrics_SumMissingTimestamps(t *testing.T) {
	raw := []byte(`{
		"resourceMetrics": [{
			"resource": {"attributes": [{"key":"service.name","value":{"stringValue":"svc"}}]},
			"scopeMetrics": [{
				"scope": {"name":"test"},
				"metrics": [{
					"name": "reqs",
					"sum": {
						"aggregationTemporality": 2,
						"isMonotonic": true,
						"dataPoints": [{
							"attributes": [],
							"asInt": "42"
						}]
					}
				}]
			}]
		}]
	}`)
	batch, err := Metrics(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Sums) != 1 {
		t.Fatalf("expected 1 sum, got %d", len(batch.Sums))
	}
	sum := batch.Sums[0]
	if sum.TimeUnix != "1970-01-01 00:00:00" {
		t.Fatalf("TimeUnix: got %q", sum.TimeUnix)
	}
	if sum.Value != "42" {
		t.Fatalf("Value: got %q", sum.Value)
	}
}

func TestMetrics_GaugeWithTimestamp(t *testing.T) {
	raw := []byte(`{
		"resourceMetrics": [{
			"resource": {"attributes": [{"key":"service.name","value":{"stringValue":"svc"}}]},
			"scopeMetrics": [{
				"scope": {"name":"test"},
				"metrics": [{
					"name": "cpu",
					"gauge": {
						"dataPoints": [{
							"attributes": [],
							"timeUnixNano": "1720000000000000000",
							"asDouble": 0.75
						}]
					}
				}]
			}]
		}]
	}`)
	batch, err := Metrics(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Gauges) != 1 {
		t.Fatalf("expected 1 gauge, got %d", len(batch.Gauges))
	}
	if batch.Gauges[0].TimeUnix != "2024-07-03 09:46:40" {
		t.Fatalf("TimeUnix: got %q", batch.Gauges[0].TimeUnix)
	}
}

func TestMetrics_HistogramBucketCounts(t *testing.T) {
	raw := []byte(`{
		"resourceMetrics": [{
			"resource": {"attributes": []},
			"scopeMetrics": [{
				"scope": {"name":"test"},
				"metrics": [{
					"name": "latency",
					"histogram": {
						"aggregationTemporality": 2,
						"dataPoints": [{
							"attributes": [],
							"count": "10",
							"bucketCounts": ["1","2","7"]
						}]
					}
				}]
			}]
		}]
	}`)
	batch, err := Metrics(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Histograms) != 1 {
		t.Fatalf("expected 1 histogram, got %d", len(batch.Histograms))
	}
	if batch.Histograms[0].Count != 10 {
		t.Fatalf("Count: got %d", batch.Histograms[0].Count)
	}
}

// --- Traces end-to-end ---

func TestTraces_MissingStartTime(t *testing.T) {
	raw := []byte(`{
		"resourceSpans": [{
			"resource": {"attributes": [{"key":"service.name","value":{"stringValue":"svc"}}]},
			"scopeSpans": [{
				"scope": {"name":"test"},
				"spans": [{
					"traceId": "abc",
					"spanId": "def",
					"name": "op",
					"endTimeUnixNano": "1720000000000000000"
				}]
			}]
		}]
	}`)
	out, err := Traces(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 trace, got %d", len(out))
	}
	if out[0].Timestamp != "2024-07-03 09:46:40" {
		t.Fatalf("Timestamp: got %q", out[0].Timestamp)
	}
}

func TestTraces_BothTimestampsMissing(t *testing.T) {
	raw := []byte(`{
		"resourceSpans": [{
			"resource": {"attributes": []},
			"scopeSpans": [{
				"scope": {"name":"test"},
				"spans": [{"traceId":"a","spanId":"b","name":"op"}]
			}]
		}]
	}`)
	out, err := Traces(raw)
	if err != nil {
		t.Fatal(err)
	}
	if out[0].Timestamp != "1970-01-01 00:00:00" {
		t.Fatalf("Timestamp: got %q", out[0].Timestamp)
	}
}

// --- Logs end-to-end ---

func TestLogs_MissingTimestamp(t *testing.T) {
	raw := []byte(`{
		"resourceLogs": [{
			"resource": {"attributes": [{"key":"service.name","value":{"stringValue":"svc"}}]},
			"scopeLogs": [{
				"scope": {"name":"test"},
				"logRecords": [{
					"body": {"stringValue": "hello"},
					"severityText": "INFO"
				}]
			}]
		}]
	}`)
	out, err := Logs(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 log, got %d", len(out))
	}
	if out[0].Timestamp != "1970-01-01 00:00:00" {
		t.Fatalf("Timestamp: got %q", out[0].Timestamp)
	}
}

func TestLogs_WithTimestamp(t *testing.T) {
	raw := []byte(`{
		"resourceLogs": [{
			"resource": {"attributes": []},
			"scopeLogs": [{
				"scope": {"name":"test"},
				"logRecords": [{
					"timeUnixNano": "1720000000000000000",
					"body": {"stringValue": "msg"}
				}]
			}]
		}]
	}`)
	out, err := Logs(raw)
	if err != nil {
		t.Fatal(err)
	}
	if out[0].Timestamp != "2024-07-03 09:46:40" {
		t.Fatalf("Timestamp: got %q", out[0].Timestamp)
	}
}
