package transform

import (
	"encoding/json"
	"strconv"

	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

type MetricsBatch struct {
	Gauges                []FlatGauge
	Sums                  []FlatSum
	Summaries             []FlatSummary
	Histograms            []FlatHistogram
	ExponentialHistograms []FlatExponentialHistogram
}

func Metrics(payload *metricspb.MetricsData) (MetricsBatch, error) {
	var batch MetricsBatch
	for _, rm := range payload.GetResourceMetrics() {
		svc := serviceNameFrom(rm.GetResource().GetAttributes())
		resAttrs := attrsToJSON(rm.GetResource().GetAttributes())

		for _, sm := range rm.GetScopeMetrics() {
			scope := sm.GetScope()
			scopeAttrs := attrsToJSON(scope.GetAttributes())

			for _, m := range sm.GetMetrics() {
				base := metricBase{
					ServiceName:           svc,
					MetricName:            m.GetName(),
					ResourceAttributes:    resAttrs,
					ResourceSchemaUrl:     rm.GetSchemaUrl(),
					ScopeName:             scope.GetName(),
					ScopeVersion:          scope.GetVersion(),
					ScopeAttributes:       scopeAttrs,
					ScopeDroppedAttrCount: int64(scope.GetDroppedAttributesCount()),
					ScopeSchemaUrl:        sm.GetSchemaUrl(),
					MetricDescription:     m.GetDescription(),
					MetricUnit:            m.GetUnit(),
				}

				switch data := m.GetData().(type) {
				case *metricspb.Metric_Gauge:
					for _, dp := range data.Gauge.GetDataPoints() {
						batch.Gauges = append(batch.Gauges, flatGauge(base, dp))
					}
				case *metricspb.Metric_Sum:
					for _, dp := range data.Sum.GetDataPoints() {
						g := flatGauge(base, dp)
						batch.Sums = append(batch.Sums, FlatSum{
							FlatGauge:              g,
							AggregationTemporality: int32(data.Sum.GetAggregationTemporality()),
							IsMonotonic:            data.Sum.GetIsMonotonic(),
						})
					}
				case *metricspb.Metric_Summary:
					for _, dp := range data.Summary.GetDataPoints() {
						batch.Summaries = append(batch.Summaries, flatSummary(base, dp))
					}
				case *metricspb.Metric_Histogram:
					for _, dp := range data.Histogram.GetDataPoints() {
						batch.Histograms = append(batch.Histograms, flatHistogram(base, dp, int32(data.Histogram.GetAggregationTemporality())))
					}
				case *metricspb.Metric_ExponentialHistogram:
					for _, dp := range data.ExponentialHistogram.GetDataPoints() {
						batch.ExponentialHistograms = append(batch.ExponentialHistograms, flatExponentialHistogram(base, dp, int32(data.ExponentialHistogram.GetAggregationTemporality())))
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

func flatGauge(b metricBase, dp *metricspb.NumberDataPoint) FlatGauge {
	var val string
	switch v := dp.GetValue().(type) {
	case *metricspb.NumberDataPoint_AsDouble:
		val = strconv.FormatFloat(v.AsDouble, 'f', -1, 64)
	case *metricspb.NumberDataPoint_AsInt:
		val = strconv.FormatInt(v.AsInt, 10)
	}
	return FlatGauge{
		ServiceName:           b.ServiceName,
		MetricName:            b.MetricName,
		TimeUnix:              nanoToDatetime(coalesceNano(dp.GetTimeUnixNano(), dp.GetStartTimeUnixNano())),
		ResourceAttributes:    b.ResourceAttributes,
		ResourceSchemaUrl:     b.ResourceSchemaUrl,
		ScopeName:             b.ScopeName,
		ScopeVersion:          b.ScopeVersion,
		ScopeAttributes:       b.ScopeAttributes,
		ScopeDroppedAttrCount: b.ScopeDroppedAttrCount,
		ScopeSchemaUrl:        b.ScopeSchemaUrl,
		MetricDescription:     b.MetricDescription,
		MetricUnit:            b.MetricUnit,
		Attributes:            attrsToJSON(dp.GetAttributes()),
		StartTimeUnix:         nanoToDatetimeNullable(dp.GetStartTimeUnixNano()),
		Value:                 val,
		Flags:                 int32(dp.GetFlags()),
		Exemplars:             "[]",
	}
}

func flatSummary(b metricBase, dp *metricspb.SummaryDataPoint) FlatSummary {
	qvJSON := marshalJSON(quantileValues(dp.GetQuantileValues()))
	return FlatSummary{
		ServiceName:           b.ServiceName,
		MetricName:            b.MetricName,
		TimeUnix:              nanoToDatetime(coalesceNano(dp.GetTimeUnixNano(), dp.GetStartTimeUnixNano())),
		ResourceAttributes:    b.ResourceAttributes,
		ResourceSchemaUrl:     b.ResourceSchemaUrl,
		ScopeName:             b.ScopeName,
		ScopeVersion:          b.ScopeVersion,
		ScopeAttributes:       b.ScopeAttributes,
		ScopeDroppedAttrCount: b.ScopeDroppedAttrCount,
		ScopeSchemaUrl:        b.ScopeSchemaUrl,
		MetricDescription:     b.MetricDescription,
		MetricUnit:            b.MetricUnit,
		Attributes:            attrsToJSON(dp.GetAttributes()),
		StartTimeUnix:         nanoToDatetimeNullable(dp.GetStartTimeUnixNano()),
		Count:                 dp.GetCount(),
		Sum:                   dp.GetSum(),
		ValueAtQuantiles:      qvJSON,
		Flags:                 int32(dp.GetFlags()),
	}
}

type outQuantile struct {
	Quantile float64 `json:"quantile"`
	Value    float64 `json:"value"`
}

func quantileValues(qs []*metricspb.SummaryDataPoint_ValueAtQuantile) []outQuantile {
	out := make([]outQuantile, 0, len(qs))
	for _, q := range qs {
		out = append(out, outQuantile{Quantile: q.GetQuantile(), Value: q.GetValue()})
	}
	return out
}

func flatHistogram(b metricBase, dp *metricspb.HistogramDataPoint, aggTemp int32) FlatHistogram {
	bcJSON, _ := json.Marshal(dp.GetBucketCounts())
	ebJSON, _ := json.Marshal(dp.GetExplicitBounds())
	exJSON := marshalJSON(exemplars(dp.GetExemplars()))
	return FlatHistogram{
		ServiceName:            b.ServiceName,
		MetricName:             b.MetricName,
		TimeUnix:               nanoToString(coalesceNano(dp.GetTimeUnixNano(), dp.GetStartTimeUnixNano())),
		ResourceAttributes:     b.ResourceAttributes,
		ResourceSchemaUrl:      b.ResourceSchemaUrl,
		ScopeName:              b.ScopeName,
		ScopeVersion:           b.ScopeVersion,
		ScopeAttributes:        b.ScopeAttributes,
		ScopeDroppedAttrCount:  b.ScopeDroppedAttrCount,
		ScopeSchemaUrl:         b.ScopeSchemaUrl,
		MetricDescription:      b.MetricDescription,
		MetricUnit:             b.MetricUnit,
		Attributes:             attrsToJSON(dp.GetAttributes()),
		StartTimeUnix:          nanoToString(dp.GetStartTimeUnixNano()),
		Count:                  dp.GetCount(),
		Sum:                    dp.GetSum(),
		BucketCounts:           string(bcJSON),
		ExplicitBounds:         string(ebJSON),
		Exemplars:              exJSON,
		Flags:                  int32(dp.GetFlags()),
		Min:                    dp.GetMin(),
		Max:                    dp.GetMax(),
		AggregationTemporality: aggTemp,
	}
}

func flatExponentialHistogram(b metricBase, dp *metricspb.ExponentialHistogramDataPoint, aggTemp int32) FlatExponentialHistogram {
	posBC, _ := json.Marshal(dp.GetPositive().GetBucketCounts())
	negBC, _ := json.Marshal(dp.GetNegative().GetBucketCounts())
	exJSON := marshalJSON(exemplars(dp.GetExemplars()))
	return FlatExponentialHistogram{
		ServiceName:            b.ServiceName,
		MetricName:             b.MetricName,
		TimeUnix:               nanoToString(coalesceNano(dp.GetTimeUnixNano(), dp.GetStartTimeUnixNano())),
		ResourceAttributes:     b.ResourceAttributes,
		ResourceSchemaUrl:      b.ResourceSchemaUrl,
		ScopeName:              b.ScopeName,
		ScopeVersion:           b.ScopeVersion,
		ScopeAttributes:        b.ScopeAttributes,
		ScopeDroppedAttrCount:  b.ScopeDroppedAttrCount,
		ScopeSchemaUrl:         b.ScopeSchemaUrl,
		MetricDescription:      b.MetricDescription,
		MetricUnit:             b.MetricUnit,
		Attributes:             attrsToJSON(dp.GetAttributes()),
		StartTimeUnix:          nanoToString(dp.GetStartTimeUnixNano()),
		Count:                  dp.GetCount(),
		Sum:                    dp.GetSum(),
		Scale:                  dp.GetScale(),
		ZeroCount:              dp.GetZeroCount(),
		PositiveOffset:         dp.GetPositive().GetOffset(),
		PositiveBucketCounts:   string(posBC),
		NegativeOffset:         dp.GetNegative().GetOffset(),
		NegativeBucketCounts:   string(negBC),
		Exemplars:              exJSON,
		Flags:                  int32(dp.GetFlags()),
		Min:                    dp.GetMin(),
		Max:                    dp.GetMax(),
		AggregationTemporality: aggTemp,
	}
}

type outExemplar struct {
	TimeUnixNano string   `json:"timeUnixNano"`
	AsDouble     *float64 `json:"asDouble,omitempty"`
	AsInt        *int64   `json:"asInt,omitempty"`
	TraceId      string   `json:"traceId"`
	SpanId       string   `json:"spanId"`
}

func exemplars(exs []*metricspb.Exemplar) []outExemplar {
	out := make([]outExemplar, 0, len(exs))
	for _, e := range exs {
		oe := outExemplar{
			TimeUnixNano: nanoToString(e.GetTimeUnixNano()),
			TraceId:      hexID(e.GetTraceId()),
			SpanId:       hexID(e.GetSpanId()),
		}
		switch v := e.GetValue().(type) {
		case *metricspb.Exemplar_AsDouble:
			d := v.AsDouble
			oe.AsDouble = &d
		case *metricspb.Exemplar_AsInt:
			i := v.AsInt
			oe.AsInt = &i
		}
		out = append(out, oe)
	}
	return out
}
