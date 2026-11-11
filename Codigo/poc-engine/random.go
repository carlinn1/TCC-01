package main

import (
	"math"
	"sort"
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

// RegularizedIncompleteBeta calculates I_x(a, b) via Lentz's continued fraction expansion
// (Numerical Recipes / Abramowitz & Stegun 26.5.8), with precision ~1e-14.
func RegularizedIncompleteBeta(x, a, b float64) float64 {
	if x <= 0.0 {
		return 0.0
	}
	if x >= 1.0 {
		return 1.0
	}
	lgA, _ := math.Lgamma(a)
	lgB, _ := math.Lgamma(b)
	lgAB, _ := math.Lgamma(a + b)
	lnBt := lgAB - lgA - lgB + a*math.Log(x) + b*math.Log(1.0-x)
	bt := math.Exp(lnBt)

	if x < (a+1.0)/(a+b+2.0) {
		return bt * betaContinuedFraction(x, a, b) / a
	}
	return 1.0 - bt*betaContinuedFraction(1.0-x, b, a)/b
}

func betaContinuedFraction(x, a, b float64) float64 {
	const maxIter = 200
	const eps = 3e-14
	const fpMin = 1e-30

	qab := a + b
	qap := a + 1.0
	qam := a - 1.0
	c := 1.0
	d := 1.0 - qab*x/qap
	if math.Abs(d) < fpMin {
		d = fpMin
	}
	d = 1.0 / d
	h := d

	for m := 1; m <= maxIter; m++ {
		mFloat := float64(m)
		m2 := 2.0 * mFloat
		aa := mFloat * (b - mFloat) * x / ((qam + m2) * (a + m2))
		d = 1.0 + aa*d
		if math.Abs(d) < fpMin {
			d = fpMin
		}
		c = 1.0 + aa/c
		if math.Abs(c) < fpMin {
			c = fpMin
		}
		d = 1.0 / d
		h *= d * c

		aa = -(a + mFloat) * (qab + mFloat) * x / ((a + m2) * (qap + m2))
		d = 1.0 + aa*d
		if math.Abs(d) < fpMin {
			d = fpMin
		}
		c = 1.0 + aa/c
		if math.Abs(c) < fpMin {
			c = fpMin
		}
		d = 1.0 / d
		del := d * c
		h *= del
		if math.Abs(del-1.0) < eps {
			break
		}
	}
	return h
}

// KolmogorovPValue computes the asymptotic two-sided Kolmogorov-Smirnov p-value with Stephens correction.
func KolmogorovPValue(d float64, n int) float64 {
	if d <= 0 {
		return 1.0
	}
	sqrtN := math.Sqrt(float64(n))
	z := (sqrtN + 0.12 + 0.11/sqrtN) * d
	sum := 0.0
	for k := 1; k <= 100; k++ {
		kf := float64(k)
		term := math.Exp(-2.0 * kf * kf * z * z)
		if k%2 == 1 {
			sum += term
		} else {
			sum -= term
		}
		if term < 1e-12 {
			break
		}
	}
	p := 2.0 * sum
	if p > 1.0 {
		return 1.0
	}
	if p < 0.0 {
		return 0.0
	}
	return p
}

// KSTaskResult stores exact full-sample and 2000-point grid KS statistics for a PERT-Beta task.
type KSTaskResult struct {
	TaskID         string  `json:"task_id"`
	To             float64 `json:"to"`
	Tm             float64 `json:"tm"`
	Tp             float64 `json:"tp"`
	Alpha          float64 `json:"alpha"`
	Beta           float64 `json:"beta"`
	MuTheo         float64 `json:"mu_theo"`
	MuSample       float64 `json:"mu_sample"`
	VarTheo        float64 `json:"var_theo"`
	VarSample      float64 `json:"var_sample"`
	SampleSize     int     `json:"sample_size"`
	DnFull         float64 `json:"dn_full"`
	PValueFull     float64 `json:"p_value_full"`
	DnGrid2000     float64 `json:"dn_grid_2000"`
	PValueGrid2000 float64 `json:"p_value_grid_2000"`
	DCrit005       float64 `json:"d_crit_005"`
}

// KSMultiSeedTaskSummary stores the 100-seed KS goodness-of-fit summary for a task (COD-06).
type KSMultiSeedTaskSummary struct {
	TaskID        string  `json:"task_id"`
	NumSeeds      int     `json:"num_seeds"`
	SamplePerSeed int     `json:"sample_per_seed"`
	MeanDn        float64 `json:"mean_dn"`
	MaxDn         float64 `json:"max_dn"`
	MeanPValue    float64 `json:"mean_p_value"`
	MedianPValue  float64 `json:"median_p_value"`
	Rejections005 int     `json:"rejections_005"`
	RejectRate005 float64 `json:"reject_rate_005"`
}

// ComputeKSBetaFit evaluates exact full-sample KS supremum D_n across all n sorted draws (COD-05).
func ComputeKSBetaFit(id string, to, tm, tp, lambda float64, n int, seed uint64) KSTaskResult {
	diff := tp - to
	a := 1.0 + lambda*(tm-to)/diff
	b := 1.0 + lambda*(tp-tm)/diff
	muTheo := (to + lambda*tm + tp) / (lambda + 2.0)
	varTheo := (muTheo - to) * (tp - muTheo) / (lambda + 3.0)

	rng := NewSplitMix64(seed)
	samples := make([]float64, n)
	var sum, sumSq float64
	for i := 0; i < n; i++ {
		x := rng.SampleBeta(a, b)
		samples[i] = x
		d := to + x*diff
		sum += d
		sumSq += d * d
	}
	muSample := sum / float64(n)
	varSample := sumSq/float64(n) - muSample*muSample

	sort.Float64s(samples)

	// Exact full-sample KS supremum over all n points (COD-05)
	maxDFull := 0.0
	nf := float64(n)
	for i := 0; i < n; i++ {
		cdfTheo := RegularizedIncompleteBeta(samples[i], a, b)
		empUpper := float64(i+1) / nf
		empLower := float64(i) / nf
		d1 := math.Abs(empUpper - cdfTheo)
		d2 := math.Abs(cdfTheo - empLower)
		if d1 > maxDFull {
			maxDFull = d1
		}
		if d2 > maxDFull {
			maxDFull = d2
		}
	}

	// 2000-point grid evaluation for comparison
	maxDGrid := 0.0
	step := n / 2000
	if step < 1 {
		step = 1
	}
	for i := 0; i < n; i += step {
		cdfTheo := RegularizedIncompleteBeta(samples[i], a, b)
		empUpper := float64(i+1) / nf
		empLower := float64(i) / nf
		d1 := math.Abs(empUpper - cdfTheo)
		d2 := math.Abs(cdfTheo - empLower)
		if d1 > maxDGrid {
			maxDGrid = d1
		}
		if d2 > maxDGrid {
			maxDGrid = d2
		}
	}

	dCrit := 1.35810 / math.Sqrt(nf)
	return KSTaskResult{
		TaskID:         id,
		To:             to,
		Tm:             tm,
		Tp:             tp,
		Alpha:          a,
		Beta:           b,
		MuTheo:         muTheo,
		MuSample:       muSample,
		VarTheo:        varTheo,
		VarSample:      varSample,
		SampleSize:     n,
		DnFull:         maxDFull,
		PValueFull:     KolmogorovPValue(maxDFull, n),
		DnGrid2000:     maxDGrid,
		PValueGrid2000: KolmogorovPValue(maxDGrid, n),
		DCrit005:       dCrit,
	}
}

// RunKSMultiSeedSuite runs numSeeds independent KS tests of size nPerSeed for a task (COD-06).
func RunKSMultiSeedSuite(id string, to, tm, tp, lambda float64, seedOffset int, numSeeds, nPerSeed int) KSMultiSeedTaskSummary {
	diff := tp - to
	a := 1.0 + lambda*(tm-to)/diff
	b := 1.0 + lambda*(tp-tm)/diff
	dCrit := 1.35810 / math.Sqrt(float64(nPerSeed))
	nf := float64(nPerSeed)

	pVals := make([]float64, numSeeds)
	var sumDn, maxDn, sumP float64
	rejections := 0

	samples := make([]float64, nPerSeed)
	for s := 0; s < numSeeds; s++ {
		seed := DeriveIterationSeed(uint64(2026+s*1000), seedOffset)
		rng := NewSplitMix64(seed)
		for i := 0; i < nPerSeed; i++ {
			samples[i] = rng.SampleBeta(a, b)
		}
		sort.Float64s(samples)

		maxD := 0.0
		for i := 0; i < nPerSeed; i++ {
			cdfTheo := RegularizedIncompleteBeta(samples[i], a, b)
			d1 := math.Abs(float64(i+1)/nf - cdfTheo)
			d2 := math.Abs(cdfTheo - float64(i)/nf)
			if d1 > maxD {
				maxD = d1
			}
			if d2 > maxD {
				maxD = d2
			}
		}
		p := KolmogorovPValue(maxD, nPerSeed)
		pVals[s] = p
		sumDn += maxD
		if maxD > maxDn {
			maxDn = maxD
		}
		sumP += p
		if maxD >= dCrit {
			rejections++
		}
	}
	sort.Float64s(pVals)
	medianP := pVals[numSeeds/2]
	if numSeeds%2 == 0 && numSeeds > 0 {
		medianP = 0.5 * (pVals[numSeeds/2-1] + pVals[numSeeds/2])
	}

	return KSMultiSeedTaskSummary{
		TaskID:        id,
		NumSeeds:      numSeeds,
		SamplePerSeed: nPerSeed,
		MeanDn:        sumDn / float64(numSeeds),
		MaxDn:         maxDn,
		MeanPValue:    sumP / float64(numSeeds),
		MedianPValue:  medianP,
		Rejections005: rejections,
		RejectRate005: float64(rejections) / float64(numSeeds) * 100.0,
	}
}


