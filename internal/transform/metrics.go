package transform

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type MetricsBatch struct {
	Gauges                []FlatGauge
	Sums                  []FlatSum
	Summaries             []FlatSummary
	Histograms            []FlatHistogram
	ExponentialHistograms []FlatExponentialHistogram
}

func Metrics(raw []byte) (MetricsBatch, error) {
	var payload OTLPMetricsPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return MetricsBatch{}, fmt.Errorf("metrics unmarshal: %w", err)
	}

	var batch MetricsBatch
	for _, rm := range payload.ResourceMetrics {
		svc := serviceNameFrom(rm.Resource.Attributes)
		resAttrs := attrsToJSON(rm.Resource.Attributes)

		for _, sm := range rm.ScopeMetrics {
			scopeAttrs := attrsToJSON(sm.Scope.Attributes)

			for _, m := range sm.Metrics {
				base := metricBase{
					ServiceName:           svc,
					MetricName:            m.Name,
					ResourceAttributes:    resAttrs,
					ResourceSchemaUrl:     rm.SchemaUrl,
					ScopeName:             sm.Scope.Name,
					ScopeVersion:          sm.Scope.Version,
					ScopeAttributes:       scopeAttrs,
					ScopeDroppedAttrCount: int64(sm.Scope.DroppedAttributesCount),
					ScopeSchemaUrl:        sm.SchemaUrl,
					MetricDescription:     m.Description,
					MetricUnit:            m.Unit,
				}

				switch {
				case m.Gauge != nil:
					for _, dp := range m.Gauge.DataPoints {
						batch.Gauges = append(batch.Gauges, flatGauge(base, dp))
					}
				case m.Sum != nil:
					for _, dp := range m.Sum.DataPoints {
						g := flatGauge(base, dp)
						batch.Sums = append(batch.Sums, FlatSum{
							FlatGauge:              g,
							AggregationTemporality: m.Sum.AggregationTemporality,
							IsMonotonic:            m.Sum.IsMonotonic,
						})
					}
				case m.Summary != nil:
					for _, dp := range m.Summary.DataPoints {
						batch.Summaries = append(batch.Summaries, flatSummary(base, dp))
					}
				case m.Histogram != nil:
					for _, dp := range m.Histogram.DataPoints {
						batch.Histograms = append(batch.Histograms, flatHistogram(base, dp, m.Histogram.AggregationTemporality))
					}
				case m.ExponentialHistogram != nil:
					for _, dp := range m.ExponentialHistogram.DataPoints {
						batch.ExponentialHistograms = append(batch.ExponentialHistograms, flatExponentialHistogram(base, dp, m.ExponentialHistogram.AggregationTemporality))
					}
				}
			}
		}
	}
	return batch, nil
}

type metricBase struct {
	ServiceName           string
	MetricName            string
	ResourceAttributes    string
	ResourceSchemaUrl     string
	ScopeName             string
	ScopeVersion          string
	ScopeAttributes       string
	ScopeDroppedAttrCount int64
	ScopeSchemaUrl        string
	MetricDescription     string
	MetricUnit            string
}

func flatGauge(b metricBase, dp NumberDataPoint) FlatGauge {
	var val string
	switch {
	case dp.AsDouble != nil:
		val = strconv.FormatFloat(*dp.AsDouble, 'f', -1, 64)
	case dp.AsInt != nil:
		val = strconv.FormatInt(int64(*dp.AsInt), 10)
	}
	return FlatGauge{
		ServiceName:           b.ServiceName,
		MetricName:            b.MetricName,
		TimeUnix:              nanoToDatetime(coalesceStr(dp.TimeUnixNano, dp.StartTimeUnixNano)),
		ResourceAttributes:    b.ResourceAttributes,
		ResourceSchemaUrl:     b.ResourceSchemaUrl,
		ScopeName:             b.ScopeName,
		ScopeVersion:          b.ScopeVersion,
		ScopeAttributes:       b.ScopeAttributes,
		ScopeDroppedAttrCount: b.ScopeDroppedAttrCount,
		ScopeSchemaUrl:        b.ScopeSchemaUrl,
		MetricDescription:     b.MetricDescription,
		MetricUnit:            b.MetricUnit,
		Attributes:            attrsToJSON(dp.Attributes),
		StartTimeUnix:         nanoToDatetimeNullable(dp.StartTimeUnixNano),
		Value:                 val,
		Flags:                 dp.Flags,
		Exemplars:             "[]",
	}
}

func flatSummary(b metricBase, dp SummaryDataPoint) FlatSummary {
	qvJSON, _ := json.Marshal(dp.QuantileValues)
	return FlatSummary{
		ServiceName:           b.ServiceName,
		MetricName:            b.MetricName,
		TimeUnix:              nanoToDatetime(coalesceStr(dp.TimeUnixNano, dp.StartTimeUnixNano)),
		ResourceAttributes:    b.ResourceAttributes,
		ResourceSchemaUrl:     b.ResourceSchemaUrl,
		ScopeName:             b.ScopeName,
		ScopeVersion:          b.ScopeVersion,
		ScopeAttributes:       b.ScopeAttributes,
		ScopeDroppedAttrCount: b.ScopeDroppedAttrCount,
		ScopeSchemaUrl:        b.ScopeSchemaUrl,
		MetricDescription:     b.MetricDescription,
		MetricUnit:            b.MetricUnit,
		Attributes:            attrsToJSON(dp.Attributes),
		StartTimeUnix:         nanoToDatetimeNullable(dp.StartTimeUnixNano),
		Count:                 uint64(dp.Count),
		Sum:                   dp.Sum,
		ValueAtQuantiles:      string(qvJSON),
		Flags:                 dp.Flags,
	}
}

func flatHistogram(b metricBase, dp HistogramDataPoint, aggTemp int32) FlatHistogram {
	bcJSON, _ := json.Marshal(dp.BucketCounts)
	ebJSON, _ := json.Marshal(dp.ExplicitBounds)
	exJSON, _ := json.Marshal(dp.Exemplars)
	return FlatHistogram{
		ServiceName:            b.ServiceName,
		MetricName:             b.MetricName,
		TimeUnix:               coalesceStr(dp.TimeUnixNano, dp.StartTimeUnixNano),
		ResourceAttributes:     b.ResourceAttributes,
		ResourceSchemaUrl:      b.ResourceSchemaUrl,
		ScopeName:              b.ScopeName,
		ScopeVersion:           b.ScopeVersion,
		ScopeAttributes:        b.ScopeAttributes,
		ScopeDroppedAttrCount:  b.ScopeDroppedAttrCount,
		ScopeSchemaUrl:         b.ScopeSchemaUrl,
		MetricDescription:      b.MetricDescription,
		MetricUnit:             b.MetricUnit,
		Attributes:             attrsToJSON(dp.Attributes),
		StartTimeUnix:          dp.StartTimeUnixNano,
		Count:                  uint64(dp.Count),
		Sum:                    float64OrZero(dp.Sum),
		BucketCounts:           string(bcJSON),
		ExplicitBounds:         string(ebJSON),
		Exemplars:              string(exJSON),
		Flags:                  dp.Flags,
		Min:                    float64OrZero(dp.Min),
		Max:                    float64OrZero(dp.Max),
		AggregationTemporality: aggTemp,
	}
}

func flatExponentialHistogram(b metricBase, dp ExponentialHistogramDataPoint, aggTemp int32) FlatExponentialHistogram {
	posBC, _ := json.Marshal(dp.Positive.BucketCounts)
	negBC, _ := json.Marshal(dp.Negative.BucketCounts)
	exJSON, _ := json.Marshal(dp.Exemplars)
	return FlatExponentialHistogram{
		ServiceName:            b.ServiceName,
		MetricName:             b.MetricName,
		TimeUnix:               coalesceStr(dp.TimeUnixNano, dp.StartTimeUnixNano),
		ResourceAttributes:     b.ResourceAttributes,
		ResourceSchemaUrl:      b.ResourceSchemaUrl,
		ScopeName:              b.ScopeName,
		ScopeVersion:           b.ScopeVersion,
		ScopeAttributes:        b.ScopeAttributes,
		ScopeDroppedAttrCount:  b.ScopeDroppedAttrCount,
		ScopeSchemaUrl:         b.ScopeSchemaUrl,
		MetricDescription:      b.MetricDescription,
		MetricUnit:             b.MetricUnit,
		Attributes:             attrsToJSON(dp.Attributes),
		StartTimeUnix:          dp.StartTimeUnixNano,
		Count:                  uint64(dp.Count),
		Sum:                    float64OrZero(dp.Sum),
		Scale:                  dp.Scale,
		ZeroCount:              uint64(dp.ZeroCount),
		PositiveOffset:         dp.Positive.Offset,
		PositiveBucketCounts:   string(posBC),
		NegativeOffset:         dp.Negative.Offset,
		NegativeBucketCounts:   string(negBC),
		Exemplars:              string(exJSON),
		Flags:                  dp.Flags,
		Min:                    float64OrZero(dp.Min),
		Max:                    float64OrZero(dp.Max),
		AggregationTemporality: aggTemp,
	}
}
