package ecosystem

import (
	"fmt"
	"strconv"
	"text/template"

	"github.com/intentd/kel/core"
)

// prometheusFuncs — Prometheus
//
// Domain: prometheus
//
// Read the response of a Prometheus HTTP API instant query (/api/v1/query),
// the object with "status" and "data": the value, sums and maxima across
// series, thresholds, series counts and label values. No data gives zero
// values; a sample that does not parse is an error.
func prometheusFuncs() template.FuncMap {
	return template.FuncMap{
		"promValue":          PromValue,
		"promSum":            PromSum,
		"promMax":            PromMax,
		"promAboveThreshold": PromAboveThreshold,
		"promBelowThreshold": PromBelowThreshold,
		"promSeriesCount":    PromSeriesCount,
		"promLabelValues":    PromLabelValues,
	}
}

// PromValue returns the query's value: a scalar result, or the first series of a vector result. 0 when there is no data; an error when the sample is not a number.
//
// Keywords: prometheus, metric, query, value, scalar, vector
//
// Example:
//
//	{{ promValue .errorRate }}  → 0.02
func PromValue(resp interface{}) (float64, error) {
	vals := promSamples(resp)
	if len(vals) == 0 {
		return 0, nil
	}
	return parseSample(vals[0])
}

// PromSum returns the sum of every series value in the result. 0 when there is no data.
//
// Keywords: prometheus, metric, sum, total, vector
//
// Example:
//
//	{{ promSum .requests }}  → 1250
func PromSum(resp interface{}) (float64, error) {
	var sum float64
	for _, v := range promSamples(resp) {
		f, err := parseSample(v)
		if err != nil {
			return 0, err
		}
		sum += f
	}
	return sum, nil
}

// PromMax returns the largest series value in the result. 0 when there is no data.
//
// Keywords: prometheus, metric, max, peak, vector
//
// Example:
//
//	{{ promMax .requests }}  → 800
func PromMax(resp interface{}) (float64, error) {
	var max float64
	for i, v := range promSamples(resp) {
		f, err := parseSample(v)
		if err != nil {
			return 0, err
		}
		if i == 0 || f > max {
			max = f
		}
	}
	return max, nil
}

// PromAboveThreshold reports whether the query's value is above threshold. False when there is no data; an error when threshold is not a number.
//
// Keywords: prometheus, metric, threshold, above, alert, compare
//
// Example:
//
//	{{ promAboveThreshold .requests 1000 }}  → false
func PromAboveThreshold(resp interface{}, threshold interface{}) (bool, error) {
	return compareValue(resp, threshold, func(v, t float64) bool { return v > t })
}

// PromBelowThreshold reports whether the query's value is below threshold. False when there is no data; an error when threshold is not a number.
//
// Keywords: prometheus, metric, threshold, below, compare
//
// Example:
//
//	{{ promBelowThreshold .errorRate 0.05 }}  → true
func PromBelowThreshold(resp interface{}, threshold interface{}) (bool, error) {
	return compareValue(resp, threshold, func(v, t float64) bool { return v < t })
}

// PromSeriesCount returns the number of series in the result: 1 for a scalar, 0 when there is no data.
//
// Keywords: prometheus, metric, series, count
//
// Example:
//
//	{{ promSeriesCount .requests }}  → 2
func PromSeriesCount(resp interface{}) int {
	return len(promSamples(resp))
}

// PromLabelValues returns the value of label on each series of a vector result, in order; series without the label are skipped.
//
// Keywords: prometheus, metric, label, values, series, list
//
// Example:
//
//	{{ promLabelValues .requests "pod" }}  → [web-0 web-1]
func PromLabelValues(resp interface{}, label string) []string {
	var vals []string
	for _, series := range promSeries(resp) {
		metric, _ := series["metric"].(map[string]interface{})
		if v, ok := metric[label].(string); ok {
			vals = append(vals, v)
		}
	}
	return vals
}

func compareValue(resp, threshold interface{}, cmp func(v, t float64) bool) (bool, error) {
	t, err := core.ToFloat(threshold)
	if err != nil {
		return false, fmt.Errorf("threshold %v is not a number", threshold)
	}
	vals := promSamples(resp)
	if len(vals) == 0 {
		return false, nil
	}
	v, err := parseSample(vals[0])
	if err != nil {
		return false, err
	}
	return cmp(v, t), nil
}

func parseSample(s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("prometheus sample %q is not a number", s)
	}
	return f, nil
}

// promData returns the response's data object, or nil.
func promData(resp interface{}) map[string]interface{} {
	m, _ := resp.(map[string]interface{})
	d, _ := m["data"].(map[string]interface{})
	return d
}

// promSeries returns the series of a vector result.
func promSeries(resp interface{}) []map[string]interface{} {
	d := promData(resp)
	if d["resultType"] != "vector" {
		return nil
	}
	result, _ := d["result"].([]interface{})
	out := make([]map[string]interface{}, 0, len(result))
	for _, r := range result {
		if s, ok := r.(map[string]interface{}); ok {
			out = append(out, s)
		}
	}
	return out
}

// promSamples returns the sample values of a scalar or vector result, as
// Prometheus sends them (strings).
func promSamples(resp interface{}) []string {
	d := promData(resp)
	switch d["resultType"] {
	case "scalar":
		pair, _ := d["result"].([]interface{})
		if len(pair) == 2 {
			if v, ok := pair[1].(string); ok {
				return []string{v}
			}
		}
	case "vector":
		var vals []string
		for _, series := range promSeries(resp) {
			pair, _ := series["value"].([]interface{})
			if len(pair) == 2 {
				if v, ok := pair[1].(string); ok {
					vals = append(vals, v)
				}
			}
		}
		return vals
	}
	return nil
}
