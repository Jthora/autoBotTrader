package normalize

import (
    "math"
    "testing"
)

func TestNonFiniteInputs(t *testing.T) {
    if AstrologyScore(math.NaN()) != 0 { t.Error("NaN should coerce to 0") }
    if GravimetricScore(math.NaN()) != 0 { t.Error("NaN to 0") }
    if AstrologyScore(math.Inf(1)) != 100 { t.Error("+Inf clamp 100") }
    if GravimetricScore(math.Inf(1)) != 100 { t.Error("+Inf clamp 100") }
    if AstrologyScore(math.Inf(-1)) != 0 { t.Error("-Inf clamp 0") }
    if GravimetricScore(math.Inf(-1)) != 0 { t.Error("-Inf clamp 0") }
}

func TestPrecisionNearBoundaries(t *testing.T) {
    mid := AstrologyScore(360.0)
    if mid < 49 || mid > 51 { t.Fatalf("mid expected ~50 got %d", mid) }
    lower := AstrologyScore(7.19)
    upper := AstrologyScore(7.21)
    if !(lower <= upper) { t.Fatalf("monotonicity violated lower=%d upper=%d", lower, upper) }
}
