package main

import (
	"math"
	"testing"
)

// TestCycleDetection tests that cyclic dependency graphs are rejected.
func TestCycleDetection(t *testing.T) {
	tasks := []Task{
		{ID: "A"}, {ID: "B"}, {ID: "C"},
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

