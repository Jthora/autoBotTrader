package normalize

const (
    volatilityMin = 0.0
    volatilityMax = 720.0
    tideMin       = 80.0
    tideMax       = 130.0
)

func scaleToScore(x, min, max float64) uint32 {
    if x < min { x = min }
    if x > max { x = max }
    span := max - min
    if span <= 0 { return 0 }
    bp := ((x - min) / span) * 10000.0
    if bp < 0 { bp = 0 }
    if bp > 10000 { bp = 10000 }
    return uint32(bp) / 100
}

func AstrologyScore(volatilityIndex float64) uint32 {
    return scaleToScore(volatilityIndex, volatilityMin, volatilityMax)
}

func GravimetricScore(tideForce float64) uint32 {
    return scaleToScore(tideForce, tideMin, tideMax)
}
