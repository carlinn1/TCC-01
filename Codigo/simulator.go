package main

import (
	"math"
	"sort"
	"sync"
	"time"
)

// RunMonteCarlo executes the concurrent Monte Carlo simulation.
// Guaranteed exact reproducibility regardless of worker count due to
// independent per-iteration seed derivation: sk = SplitMix64(S0 ^ k).
func RunMonteCarlo(net ProjectNetwork, cfg MonteCarloConfig) (*MonteCarloResult, error) {
	startTime := time.Now()

	if cfg.Iterations <= 0 {
		cfg.Iterations = 10000
	}
	if cfg.NumWorkers <= 0 {
		cfg.NumWorkers = 4
	}
	if cfg.Lambda <= 0 {
		cfg.Lambda = 4.0
	}
	if cfg.EpsilonFloat <= 0 {
		cfg.EpsilonFloat = 1e-9
	}

	// 1. Sort tasks and risks deterministically by ID
	sortedTasks := make([]Task, len(net.Tasks))
	copy(sortedTasks, net.Tasks)
	sort.Slice(sortedTasks, func(i, j int) bool {
		return sortedTasks[i].ID < sortedTasks[j].ID
	})

	sortedRisks := make([]Risk, len(net.Risks))
	copy(sortedRisks, net.Risks)
	sort.Slice(sortedRisks, func(i, j int) bool {
		return sortedRisks[i].ID < sortedRisks[j].ID
	})

	// Pre-compile graph topology for zero-allocation CPM
	compiledGraph, err := CompileGraph(sortedTasks, net.Dependencies)
	if err != nil {
		return nil, err
	}

	// Pre-calculate PERT-Beta parameters per task index (safe for degenerate To == Tp)
	type betaParams struct {
		alpha    float64
		beta     float64
		rangeVal float64
		isConst  bool
	}
	taskParams := make([]betaParams, len(sortedTasks))
	for i, t := range sortedTasks {
		diff := t.Tp - t.To
		if diff > 1e-12 && t.Tm >= t.To && t.Tm <= t.Tp {
			a := 1.0 + cfg.Lambda*(t.Tm-t.To)/diff
			b := 1.0 + cfg.Lambda*(t.Tp-t.Tm)/diff
			taskParams[i] = betaParams{
				alpha:    a,
				beta:     b,
				rangeVal: diff,
				isConst:  false,
			}
		} else {
			taskParams[i] = betaParams{
				isConst: true,
			}
		}
	}

	// Calculate deterministic CPM baseline
	detDurations := make(map[string]float64)
	for _, t := range sortedTasks {
		if t.Completed || t.Progress >= 1.0 {
			detDurations[t.ID] = 0
		} else {
			exp := CalculateExpectedDuration(t.To, t.Tm, t.Tp)
			detDurations[t.ID] = exp * (1.0 - t.Progress)
		}
	}
	detCPM, err := SolveCPM(sortedTasks, net.Dependencies, detDurations, net.StatusDate, cfg.EpsilonFloat)
	if err != nil {
		return nil, err
	}

	// Pre-map risks to integer task indices
	type compiledRisk struct {
		taskIdx    int
		hasTask    bool
		prob       float64
		impact     float64
		isAbsolute bool
	}
	activeRisks := make([]compiledRisk, 0, len(sortedRisks))
	for _, r := range sortedRisks {
		if !r.Active {
			continue
		}
		cr := compiledRisk{
			prob:       r.Probability,
			impact:     r.ImpactDays,
			isAbsolute: r.IsAbsolute,
			hasTask:    r.TaskID != "",
		}
		if cr.hasTask {
			if idx, ok := compiledGraph.TaskMap[r.TaskID]; ok {
				cr.taskIdx = idx
				activeRisks = append(activeRisks, cr)
			}
		} else {
			activeRisks = append(activeRisks, cr)
		}
	}

	// 2. Allocate pre-sized arrays for thread-safe positional writes
	finishDates := make([]float64, cfg.Iterations)
	criticalMatrix := make([][]uint8, compiledGraph.NumTasks)
	for i := 0; i < compiledGraph.NumTasks; i++ {
		if compiledGraph.IsActive[i] {
			criticalMatrix[i] = make([]uint8, cfg.Iterations)
		}
	}

	// 3. Divide iterations across workers
	var wg sync.WaitGroup
	chunkSize := (cfg.Iterations + cfg.NumWorkers - 1) / cfg.NumWorkers

	for w := 0; w < cfg.NumWorkers; w++ {
		startK := w * chunkSize
		endK := startK + chunkSize
		if endK > cfg.Iterations {
			endK = cfg.Iterations
		}
		if startK >= endK {
			continue
		}

		wg.Add(1)
		go func(workerStart, workerEnd int) {
			defer wg.Done()

			n := compiledGraph.NumTasks
			durs := make([]float64, n)
			es := make([]float64, n)
			ef := make([]float64, n)
			ls := make([]float64, n)
			lf := make([]float64, n)
			crit := make([]bool, n)

			for k := workerStart; k < workerEnd; k++ {
				iterSeed := DeriveIterationSeed(cfg.Seed, k)
				rng := NewSplitMix64(iterSeed)

				// A. Sample task base durations in fixed ID order
				for i, t := range sortedTasks {
					if !compiledGraph.IsActive[i] {
						durs[i] = 0
						continue
					}
					bp := taskParams[i]
					var d float64
					if bp.isConst {
						d = t.To
					} else {
						x := rng.SampleBeta(bp.alpha, bp.beta)
						d = t.To + x*bp.rangeVal
					}
					durs[i] = d * (1.0 - t.Progress)
				}

				// B. Sample active risks in fixed ID order
				for _, r := range activeRisks {
					if rng.SampleBernoulli(r.prob) {
						if r.hasTask {
							i := r.taskIdx
							if compiledGraph.IsActive[i] {
								scale := 1.0
								if !r.isAbsolute {
									scale = 1.0 - sortedTasks[i].Progress
								}
								durs[i] += r.impact * scale
							}
						}
					}
				}

				// C. Run FastCPM for this iteration
				tCPM := FastCPM(compiledGraph, durs, es, ef, ls, lf, crit, cfg.EpsilonFloat)
				finishDates[k] = net.StatusDate + tCPM

				for i := 0; i < n; i++ {
					if compiledGraph.IsActive[i] && crit[i] {
						criticalMatrix[i][k] = 1
					}
				}
			}
		}(startK, endK)
	}

	wg.Wait()

	// 4. Compute statistics in canonical index order (k = 0..N-1) with Kahan compensated summation
	var kahanSum, kahanC float64
	for k := 0; k < cfg.Iterations; k++ {
		y := finishDates[k] - kahanC
		t := kahanSum + y
		kahanC = (t - kahanSum) - y
		kahanSum = t
	}
	mean := kahanSum / float64(cfg.Iterations)

	var varSum, varC float64
	for k := 0; k < cfg.Iterations; k++ {
		diff := finishDates[k] - mean
		y := (diff * diff) - varC
		t := varSum + y
		varC = (t - varSum) - y
		varSum = t
	}
	var meanSE float64
	if cfg.Iterations > 1 {
		sd := math.Sqrt(varSum / float64(cfg.Iterations-1))
		meanSE = sd / math.Sqrt(float64(cfg.Iterations))
	}

	sortedFinishes := make([]float64, cfg.Iterations)
	copy(sortedFinishes, finishDates)
	sort.Float64s(sortedFinishes)

	idxP50 := int(math.Ceil(0.50*float64(cfg.Iterations))) - 1
	idxP85 := int(math.Ceil(0.85*float64(cfg.Iterations))) - 1
	idxP95 := int(math.Ceil(0.95*float64(cfg.Iterations))) - 1

	p50 := sortedFinishes[idxP50]
	p85 := sortedFinishes[idxP85]
	p95 := sortedFinishes[idxP95]

	// Distribution-free Monte Carlo Standard Error (MCSE) for quantiles via binomial order statistics
	orderStatSE := func(tau float64) float64 {
		nf := float64(cfg.Iterations)
		halfWidth := 1.96 * math.Sqrt(nf*tau*(1.0-tau))
		r := int(math.Floor(nf*tau-halfWidth)) - 1
		s := int(math.Ceil(nf*tau+halfWidth)) - 1
		if r < 0 {
			r = 0
		}
		if s >= cfg.Iterations {
			s = cfg.Iterations - 1
		}
		return (sortedFinishes[s] - sortedFinishes[r]) / (2.0 * 1.96)
	}
	p50SE := orderStatSE(0.50)
	p85SE := orderStatSE(0.85)
	p95SE := orderStatSE(0.95)

	targetCount := 0
	if cfg.TargetDate > 0 {
		for _, f := range finishDates {
			if f <= cfg.TargetDate+1e-9 {
				targetCount++
			}
		}
	}
	probTarget := float64(targetCount) / float64(cfg.Iterations) * 100.0

	critIndex := make(map[string]float64)
	for i, arr := range criticalMatrix {
		if !compiledGraph.IsActive[i] {
			continue
		}
		count := 0
		for _, c := range arr {
			if c == 1 {
				count++
			}
		}
		critIndex[sortedTasks[i].ID] = float64(count) / float64(cfg.Iterations) * 100.0
	}

	elapsed := time.Since(startTime)

	return &MonteCarloResult{
		CPMDeterministic: detCPM.FinishDate,
		Mean:             mean,
		MeanSE:           meanSE,
		P50:              p50,
		P50SE:            p50SE,
		P85:              p85,
		P85SE:            p85SE,
		P95:              p95,
		P95SE:            p95SE,
		ProbTarget:       probTarget,
		CriticalityIndex: critIndex,
		Durations:        sortedFinishes,
		DurationMs:       float64(elapsed.Microseconds()) / 1000.0,
		Iterations:       cfg.Iterations,
		NumWorkers:       cfg.NumWorkers,
	}, nil
}
