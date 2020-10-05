package main

import (
	"math"
)

func withinEpsilon(a, b float64) bool {
	if a == b {
		return true
	}

	diff := math.Abs(a - b)
	epsilon := math.Nextafter(1, 2) - 1
	minNorm := math.Nextafter(1, 2)

	if a == 0 || b == 0 || (math.Abs(a) + math.Abs(b) < minNorm) {
		return diff < (epsilon * minNorm)
	}

	return diff / math.Min((math.Abs(a) + math.Abs(b)), math.MaxFloat64) < epsilon;
}