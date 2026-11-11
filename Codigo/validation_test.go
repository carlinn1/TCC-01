package main

import (
	"math"
	"strings"
	"testing"
)

// TestRegularizedIncompleteBetaAnalytical (COD-05) verifies RegularizedIncompleteBeta(x, a, b)
// against closed-form analytical identities with tolerance < 1e-12.
func TestRegularizedIncompleteBetaAnalytical(t *testing.T) {
	// 1. Uniform Beta(1, 1): I_x(1, 1) = x
	for _, x := range []float64{0.0, 0.1, 0.25, 0.5, 0.75, 0.9, 1.0} {
		got := RegularizedIncompleteBeta(x, 1.0, 1.0)
		if math.Abs(got-x) > 1e-13 {
			t.Errorf("I_%.2f(1,1) = %.15f, want %.15f", x, got, x)
		}
	}

	// 2. Symmetry at x = 0.5: I_{0.5}(a, a) = 0.5
	for _, a := range []float64{1.0, 2.0, 2.5, 3.5, 5.0} {
		got := RegularizedIncompleteBeta(0.5, a, a)
		if math.Abs(got-0.5) > 1e-13 {
			t.Errorf("I_0.5(%.1f, %.1f) = %.15f, want 0.5", a, a, got)
		}
	}

	// 3. Closed-form polynomials:
	//    I_x(2, 2) = 3x^2 - 2x^3
	//    I_x(2, 3) = 6x^2 - 8x^3 + 3x^4
	//    I_x(3, 2) = 4x^3 - 3x^4
	for _, x := range []float64{0.1, 0.2, 0.35, 0.5, 0.7, 0.85} {
		want22 := 3*x*x - 2*x*x*x
		got22 := RegularizedIncompleteBeta(x, 2.0, 2.0)
		if math.Abs(got22-want22) > 1e-12 {
			t.Errorf("I_%.2f(2,2) = %.15f, want %.15f", x, got22, want22)
		}

		want23 := 6*x*x - 8*x*x*x + 3*x*x*x*x
		got23 := RegularizedIncompleteBeta(x, 2.0, 3.0)
		if math.Abs(got23-want23) > 1e-12 {
			t.Errorf("I_%.2f(2,3) = %.15f, want %.15f", x, got23, want23)
		}

		want32 := x * x * x * (4.0 - 3.0*x)
		got32 := RegularizedIncompleteBeta(x, 3.0, 2.0)
		if math.Abs(got32-want32) > 1e-12 {
			t.Errorf("I_%.2f(3,2) = %.15f, want %.15f", x, got32, want32)
		}

		// Reflection identity: I_x(a, b) + I_{1-x}(b, a) = 1
		sumSym := RegularizedIncompleteBeta(x, 2.333333333333, 3.666666666667) +
			RegularizedIncompleteBeta(1.0-x, 3.666666666667, 2.333333333333)
		if math.Abs(sumSym-1.0) > 1e-12 {
			t.Errorf("reflection identity failed at x=%.2f: sum=%.15f", x, sumSym)
		}
	}
}

// TestPertBetaMomentsMatchAnalytical validates the Marsaglia-Tsang + Beta sampler
// against closed-form moments mu = (to + 4*tm + tp)/6 and sigma^2 = (mu - to)*(tp - mu)/7.
func TestPertBetaMomentsMatchAnalytical(t *testing.T) {
	cases := []struct {
		id         string
		to, tm, tp float64
	}{
		{"T1", 2, 3, 5},
		{"T2", 4, 5, 9},
		{"T3", 3, 5, 10},
		{"T4", 1, 2, 4},
	}
	const n = 1_000_000
	const lambda = 4.0

	for _, c := range cases {
		diff := c.tp - c.to
		a := 1 + lambda*(c.tm-c.to)/diff
		b := 1 + lambda*(c.tp-c.tm)/diff
		wantMu := (c.to + 4*c.tm + c.tp) / 6
		wantVar := (wantMu - c.to) * (c.tp - wantMu) / (lambda + 3)

		rng := NewSplitMix64(DeriveIterationSeed(2026, 1))
		var sum, sumSq float64
		for i := 0; i < n; i++ {
			d := c.to + rng.SampleBeta(a, b)*diff
			sum += d
			sumSq += d * d
		}
		mu := sum / n
		variance := sumSq/n - mu*mu

		seMu := math.Sqrt(wantVar / n)
		if math.Abs(mu-wantMu) > 5*seMu {
			t.Errorf("%s: sample mean %.5f != analytical %.5f (5*SE = %.5f)", c.id, mu, wantMu, 5*seMu)
		}
		if math.Abs(variance-wantVar)/wantVar > 0.02 {
			t.Errorf("%s: sample variance %.5f != analytical %.5f (>2%%)", c.id, variance, wantVar)
		}
		t.Logf("%s  mu=%.5f (analytical %.5f)  var=%.5f (analytical %.5f)", c.id, mu, wantMu, variance, wantVar)
	}
}

// TestPertBetaKSGoodnessOfFit performs the exact full-sample Kolmogorov-Smirnov (KS) test
// over all n = 1,000,000 samples against the theoretical Beta CDF for T1..T4 (COD-05).
func TestPertBetaKSGoodnessOfFit(t *testing.T) {
	cases := []struct {
		id         string
		to, tm, tp float64
		seedOffset int
	}{
		{"T1", 2, 3, 5, 101},
		{"T2", 4, 5, 9, 202},
		{"T3", 3, 5, 10, 303},
		{"T4", 1, 2, 4, 404},
	}
	const n = 1_000_000
	const lambda = 4.0

	for _, c := range cases {
		res := ComputeKSBetaFit(c.id, c.to, c.tm, c.tp, lambda, n, DeriveIterationSeed(2026, c.seedOffset))
		if res.DnFull >= res.DCrit005 {
			t.Errorf("%s: exact full-sample KS D_n = %.6f >= D_crit = %.6f (p = %.4f)", c.id, res.DnFull, res.DCrit005, res.PValueFull)
		}
		if res.DnFull < res.DnGrid2000 {
			t.Errorf("%s: full supremum %.6f cannot be smaller than 2000-point grid %.6f", c.id, res.DnFull, res.DnGrid2000)
		}
		t.Logf("%s  KS full D_n=%.5f (p=%.3f) | grid2000 D_n=%.5f (p=%.3f) | D_crit=%.5f",
			c.id, res.DnFull, res.PValueFull, res.DnGrid2000, res.PValueGrid2000, res.DCrit005)
	}
}

// TestPertBetaKSMultiSeed100 (COD-06) runs 100 independent seeds (n = 10,000 per seed)
// for T1..T4 and verifies that the alpha=0.05 rejection rate lies within binomial bounds
// and that p-values are well-calibrated around 0.50.
func TestPertBetaKSMultiSeed100(t *testing.T) {
	cases := []struct {
		id         string
		to, tm, tp float64
		seedOffset int
	}{
		{"T1", 2, 3, 5, 101},
		{"T2", 4, 5, 9, 202},
		{"T3", 3, 5, 10, 303},
		{"T4", 1, 2, 4, 404},
	}
	for _, c := range cases {
		summary := RunKSMultiSeedSuite(c.id, c.to, c.tm, c.tp, 4.0, c.seedOffset, 100, 10000)
		// For 100 independent tests at alpha=0.05, Binomial(100, 0.05) has 99.9% interval [0, 13]
		if summary.Rejections005 > 12 {
			t.Errorf("%s: excessive KS rejections across 100 seeds: %d/100 (%.1f%%)", c.id, summary.Rejections005, summary.RejectRate005)
		}
		if summary.MeanPValue < 0.35 || summary.MeanPValue > 0.65 {
			t.Errorf("%s: 100-seed mean p-value %.3f outside [0.35, 0.65]", c.id, summary.MeanPValue)
		}
		t.Logf("%s (100 seeds x 10k): mean D_n=%.5f, max D_n=%.5f, mean p=%.3f, median p=%.3f, rejections@0.05=%d/100",
			c.id, summary.MeanDn, summary.MaxDn, summary.MeanPValue, summary.MedianPValue, summary.Rejections005)
	}
}

// TestSplitMix64StreamIndependence (COD-06) verifies near-zero correlation between
// consecutive iteration sub-streams (S0 ^ k vs S0 ^ (k+1)) and across tasks within iterations.
func TestSplitMix64StreamIndependence(t *testing.T) {
	const numIters = 10000
	const s0 uint64 = 42

	uCur := make([]float64, numIters)
	uNext := make([]float64, numIters)
	task1 := make([]float64, numIters)
	task2 := make([]float64, numIters)

	for k := 0; k < numIters; k++ {
		rngK := NewSplitMix64(DeriveIterationSeed(s0, k))
		rngK1 := NewSplitMix64(DeriveIterationSeed(s0, k+1))
		uCur[k] = rngK.Float64()
		uNext[k] = rngK1.Float64()

		rngTasks := NewSplitMix64(DeriveIterationSeed(s0, k))
		task1[k] = rngTasks.SamplePERTBeta(2, 3, 5, 4.0)
		task2[k] = rngTasks.SamplePERTBeta(4, 5, 9, 4.0)
	}

	pearson := func(xs, ys []float64) float64 {
		n := float64(len(xs))
		var sx, sy, sxx, syy, sxy float64
		for i := range xs {
			x, y := xs[i], ys[i]
			sx += x
			sy += y
			sxx += x * x
			syy += y * y
			sxy += x * y
		}
		cov := sxy/n - (sx/n)*(sy/n)
		vx := sxx/n - (sx/n)*(sx/n)
		vy := syy/n - (sy/n)*(sy/n)
		return cov / math.Sqrt(vx*vy)
	}

	rLag1 := pearson(uCur, uNext)
	rTasks := pearson(task1, task2)
	if math.Abs(rLag1) > 0.03 {
		t.Errorf("excessive lag-1 stream correlation between k and k+1: r = %.5f", rLag1)
	}
	if math.Abs(rTasks) > 0.03 {
		t.Errorf("excessive inter-task correlation between T1 and T2 draws: r = %.5f", rTasks)
	}
	t.Logf("Stream independence verified: lag-1 r=%.5f, inter-task r=%.5f (threshold |r| <= 0.03)", rLag1, rTasks)
}

// TestDegeneratePertBetaZeroVariance validates the degenerate case to = tm = tp.
func TestDegeneratePertBetaZeroVariance(t *testing.T) {
	rng := NewSplitMix64(2026)
	for i := 0; i < 100; i++ {
		s := rng.SamplePERTBeta(5.0, 5.0, 5.0, 4.0)
		if math.IsNaN(s) || math.Abs(s-5.0) > 1e-12 {
			t.Fatalf("SamplePERTBeta degenerate invalid: got %v, want 5.0", s)
		}
	}

	net := ProjectNetwork{
		StatusDate: 0,
		Tasks: []Task{
			{ID: "T_DEG", To: 5.0, Tm: 5.0, Tp: 5.0, Progress: 0.0},
		},
	}
	cfg := MonteCarloConfig{
		Iterations:   1000,
		Seed:         42,
		NumWorkers:   4,
		Lambda:       4.0,
		EpsilonFloat: 1e-9,
	}
	res, err := RunMonteCarlo(net, cfg)
	if err != nil {
		t.Fatalf("unexpected error in degenerate case: %v", err)
	}
	for _, d := range res.Durations {
		if math.IsNaN(d) || math.Abs(d-5.0) > 1e-12 {
			t.Fatalf("invalid degenerate duration: got %v, want 5.0", d)
		}
	}
}

// TestDeterministicCPMScenarios checks deterministic CPM of scenarios A and C against manual PERT mean sums.
func TestDeterministicCPMScenarios(t *testing.T) {
	cfg := MonteCarloConfig{Iterations: 100, Seed: 1, NumWorkers: 1, Lambda: 4, EpsilonFloat: 1e-9}

	resA, err := RunMonteCarlo(buildBaseNetwork(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(resA.CPMDeterministic-(19.0/6+5.5+13.0/6)) > 1e-9 {
		t.Errorf("Scenario A: CPM = %.6f, want %.6f", resA.CPMDeterministic, 19.0/6+5.5+13.0/6)
	}

	netC := buildBaseNetwork()
	netC.StatusDate = 5
	netC.Tasks[0].Completed, netC.Tasks[0].Progress, netC.Tasks[0].ActualDuration = true, 1, 3
	netC.Tasks[1].Progress = 0.40
	netC.Tasks[2].Progress = 0.30
	resC, err := RunMonteCarlo(netC, cfg)
	if err != nil {
		t.Fatal(err)
	}
	want := 5 + 0.7*5.5 + 13.0/6
	if math.Abs(resC.CPMDeterministic-want) > 1e-9 {
		t.Errorf("Scenario C: CPM = %.6f, want %.6f", resC.CPMDeterministic, want)
	}
}

// TestPrintReproducibilityValues logs real values for the Reproducibility Table (RNF08).
func TestPrintReproducibilityValues(t *testing.T) {
	net := buildBaseNetwork()
	net.Risks = []Risk{{ID: "R1", TaskID: "T3", Probability: 0.30, ImpactDays: 2.0, Active: true}}
	for _, w := range []int{1, 4, 16} {
		r, err := RunMonteCarlo(net, MonteCarloConfig{
			Iterations: 5000, Seed: 12345, NumWorkers: w,
			Lambda: 4, EpsilonFloat: 1e-9, TargetDate: 11,
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("workers=%2d  Mean=%.6f  P50=%.6f  P85=%.6f  P95=%.6f  P(F<=11)=%.4f%%  IC(T3)=%.4f%%",
			w, r.Mean, r.P50, r.P85, r.P95, r.ProbTarget, r.CriticalityIndex["T3"])
	}
}

// TestTCC2ScheduleMonteCarlo (COD-10) verifies the deterministic CPM duration, track sums,
// critical path, total floats, and all 4 stochastic configurations (Base, WithRisks, Stress, StressWithCuts).
func TestTCC2ScheduleMonteCarlo(t *testing.T) {
	audit, err := AuditTCC2Schedule()
	if err != nil {
		t.Fatalf("AuditTCC2Schedule failed: %v", err)
	}

	// 1. Assert deterministic CPM duration = 16.75 weeks
	if math.Abs(audit.CPMDuration-16.75) > 1e-6 {
		t.Errorf("TCC2 CPMDuration = %.6f, want 16.75", audit.CPMDuration)
	}

	// 2. Assert Track A sum = 133/12 (~11.0833), Track B sum = 129/12 (10.75), Joint sum = 49/12 (~4.0833)
	if math.Abs(audit.TrackASum-133.0/12.0) > 1e-6 {
		t.Errorf("TrackASum = %.6f, want %.6f (11.08)", audit.TrackASum, 133.0/12.0)
	}
	if math.Abs(audit.TrackBSum-10.75) > 1e-6 {
		t.Errorf("TrackBSum = %.6f, want 10.75", audit.TrackBSum)
	}
	if math.Abs(audit.JointSum-49.0/12.0) > 1e-6 {
		t.Errorf("JointSum = %.6f, want %.6f (4.08)", audit.JointSum, 49.0/12.0)
	}

	// 3. Assert critical path is E02->E03->E05->E08->E09->E10->E12->E13->E14
	wantCP := "E02->E03->E05->E08->E09->E10->E12->E13->E14"
	gotCP := strings.Join(audit.CriticalPath, "->")
	if gotCP != wantCP {
		t.Errorf("CriticalPath = %s, want %s", gotCP, wantCP)
	}

	// 4. Assert total float of E11 = 19/12 (~1.5833) and E06/E07 = 0.50
	if math.Abs(audit.TaskCPM["E11"].TF-19.0/12.0) > 1e-6 {
		t.Errorf("TF(E11) = %.6f, want %.6f (1.58)", audit.TaskCPM["E11"].TF, 19.0/12.0)
	}
	if math.Abs(audit.TaskCPM["E06"].TF-0.50) > 1e-6 {
		t.Errorf("TF(E06) = %.6f, want 0.50", audit.TaskCPM["E06"].TF)
	}
	if math.Abs(audit.TaskCPM["E07"].TF-0.50) > 1e-6 {
		t.Errorf("TF(E07) = %.6f, want 0.50", audit.TaskCPM["E07"].TF)
	}

	// 5. Assert Stress scenario exceeds 20 weeks at P85, and StressWithCuts brings P85 <= 20 weeks (DEC-4)
	if audit.Stress.P85 <= 20.0 {
		t.Errorf("Expected Stress P85 > 20.0 weeks, got %.2f", audit.Stress.P85)
	}
	if audit.StressWithCuts.P85 > 20.0 {
		t.Errorf("Expected StressWithCuts P85 <= 20.0 weeks after DEC-4 cuts, got %.2f", audit.StressWithCuts.P85)
	}

	t.Logf("TCC2 Base: CPM=%.2f | Mean=%.2f | P50=%.2f | P85=%.2f | P95=%.2f | P(F<=20)=%.1f%%",
		audit.Base.CPMDeterministic, audit.Base.Mean, audit.Base.P50, audit.Base.P85, audit.Base.P95, audit.Base.ProbTarget)
	t.Logf("TCC2 Risks: CPM=%.2f | Mean=%.2f | P50=%.2f | P85=%.2f | P95=%.2f | P(F<=20)=%.1f%%",
		audit.WithRisks.CPMDeterministic, audit.WithRisks.Mean, audit.WithRisks.P50, audit.WithRisks.P85, audit.WithRisks.P95, audit.WithRisks.ProbTarget)
	t.Logf("TCC2 Stress (+25%%): CPM=%.2f | Mean=%.2f | P50=%.2f | P85=%.2f | P95=%.2f | P(F<=20)=%.1f%%",
		audit.Stress.CPMDeterministic, audit.Stress.Mean, audit.Stress.P50, audit.Stress.P85, audit.Stress.P95, audit.Stress.ProbTarget)
	t.Logf("TCC2 Stress+Cuts: CPM=%.2f | Mean=%.2f | P50=%.2f | P85=%.2f | P95=%.2f | P(F<=20)=%.1f%%",
		audit.StressWithCuts.CPMDeterministic, audit.StressWithCuts.Mean, audit.StressWithCuts.P50, audit.StressWithCuts.P85, audit.StressWithCuts.P95, audit.StressWithCuts.ProbTarget)
}

// TestBenchmarkDispersion5Runs (COD-09 & BASE-02) measures 5-run dispersion across
// 1, 2, 4, and 8 workers and enforces the RNF06 performance gate (< 500 ms median uninstrumented).
func TestBenchmarkDispersion5Runs(t *testing.T) {
	suite, err := RunBenchmarkSuite(5)
	if err != nil {
		t.Fatalf("RunBenchmarkSuite failed: %v", err)
	}
	budgetMs := 500.0
	if testing.CoverMode() != "" {
		// Under -cover, Go instruments every basic block with shared global counters
		budgetMs = 900.0
	}
	t.Logf("Environment: OS=%s ARCH=%s Go=%s NumCPU=%d (cover=%q)", suite.GOOS, suite.GOARCH, suite.GoVersion, suite.NumCPU, testing.CoverMode())
	for _, r := range suite.Results {
		t.Logf("%s (workers=%d): median=%.2f ms | mean=%.2f±%.2f ms | [%.2f, %.2f] ms | ~%.0f it/s",
			r.Name, r.Workers, r.MedianMs, r.MeanMs, r.StdDevMs, r.MinMs, r.MaxMs, r.ThroughputItS)
		if r.Tasks == 200 && r.Workers >= 4 && r.MedianMs > budgetMs {
			t.Errorf("RNF06 gate violated for %s: median %.2f ms > %.0f ms", r.Name, r.MedianMs, budgetMs)
		}
	}
}

// TestRunAllPoCValidationAndExport verifies the end-to-end JSON and LaTeX export pipeline (COD-08 & FOR-01).
func TestRunAllPoCValidationAndExport(t *testing.T) {
	tmp := t.TempDir()
	if err := RunAllPoCValidationAndExportToDir(tmp, tmp, tmp, 3); err != nil {
		t.Fatalf("RunAllPoCValidationAndExportToDir failed: %v", err)
	}
}
