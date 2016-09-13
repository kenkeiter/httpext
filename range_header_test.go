package httpext

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkRangeParsing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseRange("resources=0-99")
	}
}

func TestRangeSuffix(t *testing.T) {
	rng, err := ParseRange("resources=-100")
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, rng.IsUnbounded(), "Suffix range should be unbounded.")
	assert.True(t, rng.IsSuffix(), "Suffix range should be a suffix.")
	assert.Equal(t, 0, rng.Offset(), "Suffix range cannot have a non-zero offset.")
	assert.Equal(t, 100, rng.Limit(), "Suffix range should have a positive limit.")
	assert.Equal(t, RangeUnconstrained, rng.First(), "First index should be 0 for suffix range.")
	assert.Equal(t, -100, rng.Last(), "Suffix range upper bound should be RangeUnconstrained.")

	_, err = rng.Format()
	assert.Error(t, err, "Format should fail when no Upper bound has been set for suffix ranges.")

	rng.SetTotal(200)
	fmt, err := rng.Format()
	assert.NoError(t, err, "No error should occur when formatting a suffix range with bounds.")
	assert.Equal(t, "resources 100-199/200", fmt, "")
}

func TestRangeUnbounded(t *testing.T) {
	rng, err := ParseRange("resources=100-")
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, rng.IsUnbounded(), "Unbounded range should be unbounded.")
	assert.False(t, rng.IsSuffix(), "Range '100-' is not a suffix.")
	assert.Equal(t, 100, rng.Offset(), "Range '100-' should have an Offset of 100.")
	assert.Equal(t, RangeUnconstrained, rng.Limit(), "When unbounded in length, range should have a -1 limit.")
	assert.Equal(t, 100, rng.First(), "Range lower bound should be 100.")
	assert.Equal(t, RangeUnconstrained, rng.Last(), "Range upper bound should be unconstrained.")

	fmt, err := rng.Format()
	assert.Error(t, err, "Formatting without total should fail.")

	rng.SetTotal(300)
	fmt, err = rng.Format()
	assert.NoError(t, err, "Formatting with total should not encounter an error.")
	assert.Equal(t, "resources 100-299/300", fmt, "Range should be formattable with total.")
}

func TestRangeBounded(t *testing.T) {
	rng, err := ParseRange("resources=100-199")
	if err != nil {
		t.Fatal(err)
	}

	assert.False(t, rng.IsUnbounded(), "Bounded range should not be unbounded.")
	assert.False(t, rng.IsSuffix(), "Bounded range is not a suffix.")
	assert.Equal(t, 100, rng.Offset(), "Bounded range's Offset should be correct.")
	assert.Equal(t, 100, rng.Limit(), "Bounded range's Limit should be correct.")
	assert.Equal(t, 100, rng.First(), "Bounded range's lower bound should be 100.")
	assert.Equal(t, 200, rng.Last(), "Bounded range's upper bound should be 200.")

	fmt, err := rng.Format()
	assert.NoError(t, err, "Range formatting should not fail when range is bounded.")
	assert.Equal(t, "resources 100-199/*", fmt, "Range should be formattable without total.")

	err = rng.SetTotal(200)
	assert.NoError(t, err, "An error should not occur when setting total.")
	fmt, err = rng.Format()
	assert.NoError(t, err, "Range should be formattable with total.")
	assert.Equal(t, "resources 100-199/200", fmt, "Range should be formattable without total.")
}
