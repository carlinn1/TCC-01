package main

import (
	"math"
)

// SplitMix64 is a high-quality fast 64-bit PRNG.
// Each iteration uses an independent seed derived via SplitMix64(S0 ^ k).
type SplitMix64 struct {
	state uint64
}

func NewSplitMix64(seed uint64) *SplitMix64 {
	return &SplitMix64{state: seed}
}

// NextUint64 produces the next pseudo-random 64-bit integer.
func (s *SplitMix64) NextUint64() uint64 {
	s.state += 0x9e3779b97f4a7c15
	z := s.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// Float64 returns a pseudo-random float64 in [0, 1) with 53 bits of precision.
func (s *SplitMix64) Float64() float64 {
	// Shift right by 11 to keep 53 bits
	return float64(s.NextUint64()>>11) / float64(uint64(1)<<53)
}

// float64Positive returns a float in (0, 1) to avoid log(0)
func (s *SplitMix64) float64Positive() float64 {
	for {
		f := s.Float64()
		if f > 0 && f < 1 {
			return f
		}
	}
}

// StandardNormal generates Z ~ N(0, 1) using Box-Muller transform
func (s *SplitMix64) StandardNormal() float64 {
	u1 := s.float64Positive()
	u2 := s.Float64()
	return math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
}

// SampleGamma generates a Gamma(alpha, 1) variate using Marsaglia and Tsang (2000).
// In PERT-Beta with lambda >= 0, alpha >= 1 always.
func (s *SplitMix64) SampleGamma(alpha float64) float64 {
	if alpha < 1.0 {
		// Ahrens-Dieter / Marsaglia property: Gamma(alpha) = Gamma(alpha+1) * U^(1/alpha)
		g := s.SampleGamma(alpha + 1.0)
		u := s.float64Positive()
		return g * math.Pow(u, 1.0/alpha)
	}

	d := alpha - 1.0/3.0
	c := 1.0 / math.Sqrt(9.0*d)

	for {
		z := s.StandardNormal()
		v := 1.0 + c*z
		if v <= 0 {
			continue
		}
		v = v * v * v
		u := s.float64Positive()

		// Squeeze test
		if u < 1.0-0.0331*z*z*z*z {
			return d * v
		}

		if math.Log(u) < 0.5*z*z+d*(1.0-v+math.Log(v)) {
			return d * v
		}
	}
}

// SampleBeta generates X ~ Beta(alpha, beta) via two Gamma variates: X = Y1 / (Y1 + Y2)
func (s *SplitMix64) SampleBeta(alpha, beta float64) float64 {
	y1 := s.SampleGamma(alpha)
	y2 := s.SampleGamma(beta)
	sum := y1 + y2
	if sum <= 0 {
		return 0.5
	}
	val := y1 / sum
	if val < 0 {
		return 0
	}
	if val > 1 {
		return 1
	}
	return val
}

// SamplePERTBeta samples from the rescaled PERT-Beta distribution on [to, tp].
// Degenerate case (tp - to <= 1e-12) returns constant duration `to` without division by zero.
func (s *SplitMix64) SamplePERTBeta(to, tm, tp, lambda float64) float64 {
	diff := tp - to
	if diff <= 1e-12 || tm < to || tm > tp {
		return to
	}
	if lambda <= 0 {
		lambda = 4.0
	}
	alpha := 1.0 + lambda*(tm-to)/diff
	beta := 1.0 + lambda*(tp-tm)/diff
	return to + s.SampleBeta(alpha, beta)*diff
}

// SampleBernoulli returns true with probability p
func (s *SplitMix64) SampleBernoulli(p float64) bool {
	if p <= 0 {
		return false
	}
	if p >= 1 {
		return true
	}
	return s.Float64() < p
}

// DeriveIterationSeed calculates sk = SplitMix64(S0 ^ k)
func DeriveIterationSeed(s0 uint64, k int) uint64 {
	sm := NewSplitMix64(s0 ^ uint64(k))
	return sm.NextUint64()
}

