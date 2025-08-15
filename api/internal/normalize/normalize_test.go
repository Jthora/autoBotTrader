package normalize

import "testing"

func TestAstrologyScoreBounds(t *testing.T) {
    if AstrologyScore(-10) != 0 { t.Fatal("below min should clamp to 0") }
    if AstrologyScore(0) != 0 { t.Fatal("0 -> 0") }
    if AstrologyScore(720) != 100 { t.Fatalf("720 -> 100 got %d", AstrologyScore(720)) }
    if AstrologyScore(800) != 100 { t.Fatal("above max clamp 100") }
}

func TestGravimetricScoreBounds(t *testing.T) {
    if GravimetricScore(10) != 0 { t.Fatal("below min tide -> 0") }
    if GravimetricScore(80) != 0 { t.Fatalf("80 -> 0 got %d", GravimetricScore(80)) }
    mid := GravimetricScore(105) // midpoint of 80-130 should approximate 50
    if mid < 48 || mid > 52 { t.Fatalf("midpoint expected ~50 got %d", mid) }
    if GravimetricScore(130) != 100 { t.Fatalf("130 -> 100 got %d", GravimetricScore(130)) }
    if GravimetricScore(1000) != 100 { t.Fatal("above max -> 100") }
}
