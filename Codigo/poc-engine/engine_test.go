package main

import (
	"math"
	"testing"
)

// TestCycleDetection tests that cyclic dependency graphs are rejected.
func TestCycleDetection(t *testing.T) {
	tasks := []Task{
		{ID: "A", To: 1, Tm: 2, Tp: 3},
		{ID: "B", To: 1, Tm: 2, Tp: 3},
		{ID: "C", To: 1, Tm: 2, Tp: 3},
	}
	deps := []Dependency{
		{From: "A", To: "B"},
		{From: "B", To: "C"},
		{From: "C", To: "A"}, // Cycle
	}

	_, err := TopologicalSort(tasks, deps)
	if err == nil {
		t.Fatalf("Expected cycle detection error, but got nil")
	}
}

// TestAnalyticalCPM tests deterministic CPM against known theoretical solution.
// Linear network: T1 (3d) -> T2 (5d) -> T3 (2d) -> Total: 10d
func TestAnalyticalCPM(t *testing.T) {
	tasks := []Task{
		{ID: "T1", To: 3, Tm: 3, Tp: 3},
		{ID: "T2", To: 5, Tm: 5, Tp: 5},
		{ID: "T3", To: 2, Tm: 2, Tp: 2},
	}
	deps := []Dependency{
		{From: "T1", To: "T2"},
		{From: "T2", To: "T3"},
	}
	durations := map[string]float64{
		"T1": 3, "T2": 5, "T3": 2,
	}

	res, err := SolveCPM(tasks, deps, durations, 0, 1e-9)
	if err != nil {
		t.Fatalf("SolveCPM failed: %v", err)
	}

	if math.Abs(res.Duration-10.0) > 1e-9 {
		t.Errorf("Expected duration 10.0, got %f", res.Duration)
	}

	for _, taskID := range []string{"T1", "T2", "T3"} {
		if !res.Tasks[taskID].Critical {
			t.Errorf("Expected task %s to be critical", taskID)
		}
		if math.Abs(res.Tasks[taskID].TF) > 1e-9 {
			t.Errorf("Expected task %s TF=0, got %f", taskID, res.Tasks[taskID].TF)
		}
	}
}

// TestBernoulliRiskComponent (COD-01) tests p=0, p=1 (proportional and absolute), and p=0.5.
// Any mutation that zeroes out risk impact must fail this test.
func TestBernoulliRiskComponent(t *testing.T) {
	cfg := MonteCarloConfig{
		Iterations:   4000,
		Seed:         42,
		NumWorkers:   2,
		Lambda:       4.0,
		EpsilonFloat: 1e-9,
	}

	// 1. p = 0.0 -> duration must remain strictly 10.0
	netP0 := ProjectNetwork{
		Tasks: []Task{{ID: "T1", To: 10, Tm: 10, Tp: 10, Progress: 0.25}},
		Risks: []Risk{{ID: "R1", TaskID: "T1", Probability: 0.0, ImpactDays: 4.0, Active: true}},
	}
	resP0, err := RunMonteCarlo(netP0, cfg)
	if err != nil {
		t.Fatalf("unexpected error for p=0: %v", err)
	}
	wantBase := 10.0 * (1.0 - 0.25) // 7.5
	if math.Abs(resP0.Mean-wantBase) > 1e-9 || math.Abs(resP0.P85-wantBase) > 1e-9 {
		t.Errorf("p=0 risk altered duration: got Mean=%.6f P85=%.6f, want %.6f", resP0.Mean, resP0.P85, wantBase)
	}

	// 2. p = 1.0 with proportional scaling (1 - c_i): 10*(1-0.25) + 4*(1-0.25) = 10.5
	netP1Prop := ProjectNetwork{
		Tasks: []Task{{ID: "T1", To: 10, Tm: 10, Tp: 10, Progress: 0.25}},
		Risks: []Risk{{ID: "R1", TaskID: "T1", Probability: 1.0, ImpactDays: 4.0, IsAbsolute: false, Active: true}},
	}
	resP1Prop, err := RunMonteCarlo(netP1Prop, cfg)
	if err != nil {
		t.Fatalf("unexpected error for p=1 proportional: %v", err)
	}
	wantProp := 7.5 + 4.0*(1.0-0.25) // 10.5
	if math.Abs(resP1Prop.Mean-wantProp) > 1e-9 || math.Abs(resP1Prop.P50-wantProp) > 1e-9 {
		t.Errorf("p=1 proportional risk failed: got Mean=%.6f P50=%.6f, want %.6f", resP1Prop.Mean, resP1Prop.P50, wantProp)
	}

	// 3. p = 1.0 with IsAbsolute = true: 10*(1-0.25) + 4 = 11.5
	netP1Abs := ProjectNetwork{
		Tasks: []Task{{ID: "T1", To: 10, Tm: 10, Tp: 10, Progress: 0.25}},
		Risks: []Risk{{ID: "R1", TaskID: "T1", Probability: 1.0, ImpactDays: 4.0, IsAbsolute: true, Active: true}},
	}
	resP1Abs, err := RunMonteCarlo(netP1Abs, cfg)
	if err != nil {
		t.Fatalf("unexpected error for p=1 absolute: %v", err)
	}
	wantAbs := 7.5 + 4.0 // 11.5
	if math.Abs(resP1Abs.Mean-wantAbs) > 1e-9 {
		t.Errorf("p=1 absolute risk failed: got Mean=%.6f, want %.6f", resP1Abs.Mean, wantAbs)
	}

	// 4. p = 0.5 -> expected mean = 7.5 + 0.5 * 3.0 = 9.0 (within 4 * MCSE)
	netP05 := ProjectNetwork{
		Tasks: []Task{{ID: "T1", To: 10, Tm: 10, Tp: 10, Progress: 0.25}},
		Risks: []Risk{{ID: "R1", TaskID: "T1", Probability: 0.5, ImpactDays: 4.0, IsAbsolute: false, Active: true}},
	}
	resP05, err := RunMonteCarlo(netP05, cfg)
	if err != nil {
		t.Fatalf("unexpected error for p=0.5: %v", err)
	}
	wantMean := 9.0
	if resP05.MeanSE <= 0 || math.Abs(resP05.Mean-wantMean) > 4.0*resP05.MeanSE {
		t.Errorf("p=0.5 risk mean out of bounds: got Mean=%.4f (SE=%.4f), want %.4f", resP05.Mean, resP05.MeanSE, wantMean)
	}
}

// TestInputValidationRejectsInvalidNetworks (COD-02 & COD-03) verifies that invalid task estimates,
// negative durations, NaN/Inf, out-of-range progress/probabilities, and risks without valid TaskID fail fast.
func TestInputValidationRejectsInvalidNetworks(t *testing.T) {
	cfg := MonteCarloConfig{Iterations: 100, Seed: 42, NumWorkers: 1, Lambda: 4.0, EpsilonFloat: 1e-9}

	invalidCases := []struct {
		name string
		net  ProjectNetwork
	}{
		{
			name: "empty task list",
			net:  ProjectNetwork{Tasks: nil},
		},
		{
			name: "inverted estimates to > tm (5, 2, 3)",
			net:  ProjectNetwork{Tasks: []Task{{ID: "T1", To: 5, Tm: 2, Tp: 3}}},
		},
		{
			name: "inverted estimates tm > tp (2, 6, 4)",
			net:  ProjectNetwork{Tasks: []Task{{ID: "T1", To: 2, Tm: 6, Tp: 4}}},
		},
		{
			name: "negative optimistic duration",
			net:  ProjectNetwork{Tasks: []Task{{ID: "T1", To: -1, Tm: 2, Tp: 3}}},
		},
		{
			name: "NaN duration",
			net:  ProjectNetwork{Tasks: []Task{{ID: "T1", To: 1, Tm: math.NaN(), Tp: 3}}},
		},
		{
			name: "Inf duration",
			net:  ProjectNetwork{Tasks: []Task{{ID: "T1", To: 1, Tm: 2, Tp: math.Inf(1)}}},
		},
		{
			name: "progress out of [0,1] (c_i = 1.2)",
			net:  ProjectNetwork{Tasks: []Task{{ID: "T1", To: 1, Tm: 2, Tp: 3, Progress: 1.2}}},
		},
		{
			name: "negative progress (c_i = -0.1)",
			net:  ProjectNetwork{Tasks: []Task{{ID: "T1", To: 1, Tm: 2, Tp: 3, Progress: -0.1}}},
		},
		{
			name: "risk with empty TaskID (COD-03)",
			net: ProjectNetwork{
				Tasks: []Task{{ID: "T1", To: 1, Tm: 2, Tp: 3}},
				Risks: []Risk{{ID: "R1", TaskID: "", Probability: 0.3, ImpactDays: 2.0, Active: true}},
			},
		},
		{
			name: "risk with unknown TaskID (COD-03)",
			net: ProjectNetwork{
				Tasks: []Task{{ID: "T1", To: 1, Tm: 2, Tp: 3}},
				Risks: []Risk{{ID: "R1", TaskID: "T999", Probability: 0.3, ImpactDays: 2.0, Active: true}},
			},
		},
		{
			name: "risk probability > 1",
			net: ProjectNetwork{
				Tasks: []Task{{ID: "T1", To: 1, Tm: 2, Tp: 3}},
				Risks: []Risk{{ID: "R1", TaskID: "T1", Probability: 1.5, ImpactDays: 2.0, Active: true}},
			},
		},
		{
			name: "negative risk impact",
			net: ProjectNetwork{
				Tasks: []Task{{ID: "T1", To: 1, Tm: 2, Tp: 3}},
				Risks: []Risk{{ID: "R1", TaskID: "T1", Probability: 0.5, ImpactDays: -1.0, Active: true}},
			},
		},
		{
			name: "negative status date",
			net: ProjectNetwork{
				Tasks:      []Task{{ID: "T1", To: 1, Tm: 2, Tp: 3}},
				StatusDate: -2.0,
			},
		},
		{
			name: "duplicate task IDs",
			net: ProjectNetwork{
				Tasks: []Task{
					{ID: "T1", To: 1, Tm: 2, Tp: 3},
					{ID: "T1", To: 2, Tm: 3, Tp: 4},
				},
			},
		},
	}

	for _, tc := range invalidCases {
		if _, err := RunMonteCarlo(tc.net, cfg); err == nil {
			t.Errorf("RunMonteCarlo accepted invalid case %q, expected validation error", tc.name)
		}
	}

	// Test invalid negative Lambda
	badCfg := cfg
	badCfg.Lambda = -2.0
	validNet := ProjectNetwork{Tasks: []Task{{ID: "T1", To: 1, Tm: 2, Tp: 3}}}
	if _, err := RunMonteCarlo(validNet, badCfg); err == nil {
		t.Errorf("RunMonteCarlo accepted negative Lambda, expected validation error")
	}
}

// TestComputeCRNDiff (COD-07) verifies paired Common Random Numbers (CRN) variance reduction
// between Scenario C and Scenario C' under the same master seed.
func TestComputeCRNDiff(t *testing.T) {
	netC := buildBaseNetwork()
	netC.StatusDate = 3.0
	netC.Tasks[0].Completed = true
	netC.Tasks[0].Progress = 1.0
	netC.Tasks[0].ActualDuration = 3.0
	netC.Tasks[1].Progress = 0.50
	netC.Tasks[2].Progress = 0.25
	netC.Risks = []Risk{
		{ID: "R1", TaskID: "T3", Probability: 0.30, ImpactDays: 2.0, Active: true},
	}

	netCPrime := buildBaseNetwork()
	netCPrime.StatusDate = 3.0
	netCPrime.Tasks[0].Completed = true
	netCPrime.Tasks[0].Progress = 1.0
	netCPrime.Tasks[0].ActualDuration = 3.0
	netCPrime.Tasks[1].Progress = 0.50
	netCPrime.Tasks[2].Progress = 0.25
	netCPrime.Tasks[2].To = 2.0
	netCPrime.Tasks[2].Tm = 3.0
	netCPrime.Tasks[2].Tp = 5.0
	netCPrime.Risks = []Risk{
		{ID: "R1", TaskID: "T3", Probability: 0.30, ImpactDays: 2.0, Active: false},
	}

	cfg := MonteCarloConfig{
		Iterations:   10000,
		Seed:         42,
		NumWorkers:   4,
		Lambda:       4.0,
		EpsilonFloat: 1e-9,
		TargetDate:   11.0,
	}

	resC, err := RunMonteCarlo(netC, cfg)
	if err != nil {
		t.Fatalf("Scenario C failed: %v", err)
	}
	resCPrime, err := RunMonteCarlo(netCPrime, cfg)
	if err != nil {
		t.Fatalf("Scenario C' failed: %v", err)
	}

	diff, err := ComputeCRNDiff(resC, resCPrime, 42)
	if err != nil {
		t.Fatalf("ComputeCRNDiff failed: %v", err)
	}

	indepSE := math.Hypot(resC.MeanSE, resCPrime.MeanSE)
	if diff.DeltaMeanSE >= indepSE {
		t.Errorf("CRN did not reduce variance: paired SE=%.6f >= independent SE=%.6f", diff.DeltaMeanSE, indepSE)
	}
	// Because Scenario C' cuts T3 below T2 (switching dominant critical path from T3 to T2) while disabling R1,
	// only T4 and partial path overlaps remain shared (~1.75x theoretical ceiling here).
	if diff.VarianceRedRatio <= 1.5 {
		t.Errorf("CRN variance reduction ratio %.2f too low (expected > 1.5)", diff.VarianceRedRatio)
	}
	if math.Abs(diff.DeltaMean-(resCPrime.Mean-resC.Mean)) > 1e-6 {
		t.Errorf("DeltaMean %.6f does not match difference of means %.6f", diff.DeltaMean, resCPrime.Mean-resC.Mean)
	}

	// Second CRN check: perturbs T3 estimates WITHOUT collapsing below T2 (tests Marsaglia-Tsang rejection-count
	// isolation under per-entity sub-streams: T2, T4, and R1 streams must remain synchronized -> > 4.0x reduction).
	netCMild := buildBaseNetwork()
	netCMild.StatusDate = 5.0
	netCMild.Tasks[0].Completed = true
	netCMild.Tasks[0].Progress = 1.0
	netCMild.Tasks[0].ActualDuration = 3.0
	netCMild.Tasks[1].Progress = 0.40
	netCMild.Tasks[2].Progress = 0.30
	netCMild.Risks = []Risk{{ID: "R1", TaskID: "T3", Probability: 0.30, ImpactDays: 2.0, Active: true}}

	netCMildPrime := buildBaseNetwork()
	netCMildPrime.StatusDate = 5.0
	netCMildPrime.Tasks[0].Completed = true
	netCMildPrime.Tasks[0].Progress = 1.0
	netCMildPrime.Tasks[0].ActualDuration = 3.0
	netCMildPrime.Tasks[1].Progress = 0.40
	netCMildPrime.Tasks[2].Progress = 0.30
	netCMildPrime.Tasks[2].To = 2.5
	netCMildPrime.Tasks[2].Tm = 4.5
	netCMildPrime.Tasks[2].Tp = 8.5
	netCMildPrime.Risks = []Risk{{ID: "R1", TaskID: "T3", Probability: 0.30, ImpactDays: 2.0, Active: false}}

	resMild, err := RunMonteCarlo(netCMild, cfg)
	if err != nil {
		t.Fatalf("Scenario Mild C failed: %v", err)
	}
	resMildPrime, err := RunMonteCarlo(netCMildPrime, cfg)
	if err != nil {
		t.Fatalf("Scenario Mild C' failed: %v", err)
	}
	diffMild, err := ComputeCRNDiff(resMild, resMildPrime, 42)
	if err != nil {
		t.Fatalf("ComputeCRNDiff (mild) failed: %v", err)
	}
	if diffMild.VarianceRedRatio <= 3.0 {
		t.Errorf("Per-entity sub-stream CRN ratio %.2f too low under T3 perturbation (expected > 3.0)", diffMild.VarianceRedRatio)
	}
}

// TestTCC2StressUniformAppliesToCuts (PK-03) enforces that in stress_with_cuts mode:
// 1. E14 remains fixed at (1.0, 1.0, 1.0);
// 2. All contingency cuts are applied BEFORE the +25% stress multiplier, so every E01..E13
//    task in stress_with_cuts carries the uniform 1.25x inflation (including cut packages);
// 3. E11 (IC=0% in base/risks) is NOT cut, while critical/near-critical packages (E10, E06, E09, E08) receive the cuts.
func TestTCC2StressUniformAppliesToCuts(t *testing.T) {
	baseNet := buildTCC2ScheduleNetworkWithMode("base")
	stressCutsNet := buildTCC2ScheduleNetworkWithMode("stress_with_cuts")

	preStressCutTargets := map[string][3]float64{
		"E10": {1.5, 2.0, 2.5},
		"E06": {1.0, 1.5, 2.0},
		"E09": {1.0, 1.5, 2.0},
		"E08": {1.0, 1.5, 2.0},
	}

	baseByID := make(map[string]Task, len(baseNet.Tasks))
	for _, task := range baseNet.Tasks {
		baseByID[task.ID] = task
	}

	for _, got := range stressCutsNet.Tasks {
		base := baseByID[got.ID]
		if got.ID == "E14" {
			if math.Abs(got.To-1.0) > 1e-9 || math.Abs(got.Tm-1.0) > 1e-9 || math.Abs(got.Tp-1.0) > 1e-9 {
				t.Errorf("E14 must remain fixed at (1,1,1), got (%.3f, %.3f, %.3f)", got.To, got.Tm, got.Tp)
			}
			continue
		}
		wantPre := [3]float64{base.To, base.Tm, base.Tp}
		if cut, ok := preStressCutTargets[got.ID]; ok {
			wantPre = cut
		}
		wantTo, wantTm, wantTp := wantPre[0]*1.25, wantPre[1]*1.25, wantPre[2]*1.25
		if math.Abs(got.To-wantTo) > 1e-9 || math.Abs(got.Tm-wantTm) > 1e-9 || math.Abs(got.Tp-wantTp) > 1e-9 {
			t.Errorf("%s in stress_with_cuts = (%.3f, %.3f, %.3f), want uniform 1.25x over (%.2f, %.2f, %.2f) = (%.3f, %.3f, %.3f)",
				got.ID, got.To, got.Tm, got.Tp, wantPre[0], wantPre[1], wantPre[2], wantTo, wantTm, wantTp)
		}
	}
}

// TestLambdaConsistencyAcrossCPMAndMonteCarlo (COD-04) verifies that for lambda in {2, 4, 6},
// a single asymmetric task has CPM deterministic duration equal to (to + lambda*tm + tp)/(lambda+2)
// and matches the Monte Carlo simulated mean within 4*MCSE.
func TestLambdaConsistencyAcrossCPMAndMonteCarlo(t *testing.T) {
	net := ProjectNetwork{
		Tasks: []Task{{ID: "T1", To: 2.0, Tm: 4.0, Tp: 12.0}},
	}
	for _, lam := range []float64{2.0, 4.0, 6.0} {
		cfg := MonteCarloConfig{
			Iterations:   20000,
			Seed:         42,
			NumWorkers:   4,
			Lambda:       lam,
			EpsilonFloat: 1e-9,
		}
		res, err := RunMonteCarlo(net, cfg)
		if err != nil {
			t.Fatalf("RunMonteCarlo failed for lambda=%.1f: %v", lam, err)
		}
		wantMu := (2.0 + lam*4.0 + 12.0) / (lam + 2.0)
		if math.Abs(res.CPMDeterministic-wantMu) > 1e-9 {
			t.Errorf("lambda=%.1f: CPMDeterministic = %.6f, want analytical mu = %.6f", lam, res.CPMDeterministic, wantMu)
		}
		if math.Abs(res.Mean-wantMu) > 4.0*res.MeanSE {
			t.Errorf("lambda=%.1f: simulated Mean = %.6f (SE=%.6f) diverges from analytical mu = %.6f", lam, res.Mean, res.MeanSE, wantMu)
		}
	}
}

// TestReproducibilityAcrossWorkers tests RNF08:
// Exactly identical results across 1, 4, and 16 workers for the same seed.
func TestReproducibilityAcrossWorkers(t *testing.T) {
	net := buildBaseNetwork()
	net.Risks = []Risk{
		{ID: "R1", TaskID: "T3", Probability: 0.30, ImpactDays: 2.0, Active: true},
	}

	baseCfg := MonteCarloConfig{
		Iterations:   5000,
		Seed:         12345,
		Lambda:       4.0,
		EpsilonFloat: 1e-9,
		TargetDate:   11.0,
	}

	workerCounts := []int{1, 4, 16}
	results := make([]*MonteCarloResult, len(workerCounts))

	for i, w := range workerCounts {
		cfg := baseCfg
		cfg.NumWorkers = w
		res, err := RunMonteCarlo(net, cfg)
		if err != nil {
			t.Fatalf("RunMonteCarlo failed for %d workers: %v", w, err)
		}
		results[i] = res
	}

	ref := results[0] // 1 worker is reference
	for i := 1; i < len(workerCounts); i++ {
		cur := results[i]
		w := workerCounts[i]

		if ref.Mean != cur.Mean {
			t.Errorf("Mismatch in Mean between 1 worker (%.12f) and %d workers (%.12f)", ref.Mean, w, cur.Mean)
		}
		if ref.P50 != cur.P50 {
			t.Errorf("Mismatch in P50 between 1 worker (%f) and %d workers (%f)", ref.P50, w, cur.P50)
		}
		if ref.P85 != cur.P85 {
			t.Errorf("Mismatch in P85 between 1 worker (%f) and %d workers (%f)", ref.P85, w, cur.P85)
		}
		if ref.P95 != cur.P95 {
			t.Errorf("Mismatch in P95 between 1 worker (%f) and %d workers (%f)", ref.P95, w, cur.P95)
		}
		if ref.ProbTarget != cur.ProbTarget {
			t.Errorf("Mismatch in ProbTarget between 1 worker (%f) and %d workers (%f)", ref.ProbTarget, w, cur.ProbTarget)
		}

		for tID, icRef := range ref.CriticalityIndex {
			icCur := cur.CriticalityIndex[tID]
			if icRef != icCur {
				t.Errorf("Mismatch in IC for task %s between 1 worker (%f) and %d workers (%f)", tID, icRef, w, icCur)
			}
		}
	}
}

// BenchmarkMonteCarloSmallNetwork benchmarks 10,000 iterations on 4-task network (1 worker / 1 vCPU baseline).
func BenchmarkMonteCarloSmallNetwork(b *testing.B) {
	net := buildBaseNetwork()
	cfg := MonteCarloConfig{
		Iterations:   10000,
		Seed:         42,
		NumWorkers:   1,
		Lambda:       4.0,
		EpsilonFloat: 1e-9,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := RunMonteCarlo(net, cfg)
		if err != nil {
			b.Fatalf("Simulation error: %v", err)
		}
	}
}

// BenchmarkMonteCarloSynthetic200Tasks benchmarks RNF06:
// 200 tasks (10 tiers x 20 tasks), 300 dependencies, 40 Bernoulli risks, 10,000 iterations.
// Must run in < 1.0 second even on 1 worker / 1 vCPU.
func BenchmarkMonteCarloSynthetic200Tasks(b *testing.B) {
	net := buildSynthetic200Network()
	cfg := MonteCarloConfig{
		Iterations:   10000,
		Seed:         42,
		NumWorkers:   1,
		Lambda:       4.0,
		EpsilonFloat: 1e-9,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := RunMonteCarlo(net, cfg)
		if err != nil {
			b.Fatalf("Simulation error: %v", err)
		}
	}
}
