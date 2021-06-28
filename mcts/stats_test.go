package mcts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSampleStats(t *testing.T) {
	samples := []float64{1, 2, 2, 4, 6}

	stats := sampleStats(samples)

	require.Equal(t, 6.0, stats.max)
	require.Equal(t, 1.0, stats.min)
	require.Equal(t, 15.0, stats.sum)
	require.Equal(t, 5.0, stats.count)
	require.Equal(t, 3.0, stats.mean)
	require.Equal(t, 2.0, stats.stddev)
}
