package ecosystem_test

import (
	"testing"

	"github.com/intentd/kel/ecosystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// vector is a Prometheus API instant-query response with one series per
// value; each series gets an "instance" label.
func vector(values ...string) map[string]interface{} {
	result := make([]interface{}, len(values))
	for i, v := range values {
		result[i] = map[string]interface{}{
			"metric": map[string]interface{}{"instance": "i" + v},
			"value":  []interface{}{1753228800.0, v},
		}
	}
	return map[string]interface{}{
		"status": "success",
		"data":   map[string]interface{}{"resultType": "vector", "result": result},
	}
}

func scalar(value string) map[string]interface{} {
	return map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"resultType": "scalar",
			"result":     []interface{}{1753228800.0, value},
		},
	}
}

func TestPromValue(t *testing.T) {
	v, err := ecosystem.PromValue(vector("42", "7"))
	require.NoError(t, err)
	assert.Equal(t, 42.0, v, "first series")

	v, err = ecosystem.PromValue(scalar("3.14"))
	require.NoError(t, err)
	assert.Equal(t, 3.14, v)

	v, err = ecosystem.PromValue(nil)
	require.NoError(t, err, "no data is not an error")
	assert.Zero(t, v)

	_, err = ecosystem.PromValue(vector("not-a-number"))
	assert.Error(t, err)
}

func TestPromSumAndMax(t *testing.T) {
	sum, err := ecosystem.PromSum(vector("10", "20", "30"))
	require.NoError(t, err)
	assert.Equal(t, 60.0, sum)

	max, err := ecosystem.PromMax(vector("10", "30", "20"))
	require.NoError(t, err)
	assert.Equal(t, 30.0, max)

	sum, err = ecosystem.PromSum(vector())
	require.NoError(t, err)
	assert.Zero(t, sum)
}

func TestPromThresholds(t *testing.T) {
	above, err := ecosystem.PromAboveThreshold(vector("1500"), 1000)
	require.NoError(t, err)
	assert.True(t, above)

	below, err := ecosystem.PromBelowThreshold(vector("1500"), "1000")
	require.NoError(t, err)
	assert.False(t, below)

	above, err = ecosystem.PromAboveThreshold(nil, 1000)
	require.NoError(t, err)
	assert.False(t, above, "no data is not above anything")

	_, err = ecosystem.PromAboveThreshold(vector("1"), "lots")
	assert.Error(t, err, "a threshold that is not a number is a mistake")
}

func TestPromSeriesCountAndLabels(t *testing.T) {
	assert.Equal(t, 3, ecosystem.PromSeriesCount(vector("1", "2", "3")))
	assert.Equal(t, 1, ecosystem.PromSeriesCount(scalar("1")))
	assert.Equal(t, 0, ecosystem.PromSeriesCount(nil))

	assert.Equal(t, []string{"i1", "i2"}, ecosystem.PromLabelValues(vector("1", "2"), "instance"))
	assert.Empty(t, ecosystem.PromLabelValues(vector("1"), "missing"))
}
