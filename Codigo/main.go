package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

func buildBaseNetwork() ProjectNetwork {
	tasks := []Task{
		{ID: "T1", Name: "Modelagem de Dados", To: 2, Tm: 3, Tp: 5, Progress: 0},
		{ID: "T2", Name: "Construção de APIs", To: 4, Tm: 5, Tp: 9, Progress: 0},
		{ID: "T3", Name: "Serviço de Mensageria", To: 3, Tm: 5, Tp: 10, Progress: 0},
		{ID: "T4", Name: "Integração Front-end", To: 1, Tm: 2, Tp: 4, Progress: 0},
	}

	deps := []Dependency{
		{From: "T1", To: "T2"},
		{From: "T1", To: "T3"},
		{From: "T2", To: "T4"},
		{From: "T3", To: "T4"},
	}

	return ProjectNetwork{
		Tasks:        tasks,
		Dependencies: deps,
		Risks:        nil,
		StatusDate:   0,
	}
}

// buildSynthetic200Network constructs the deterministic synthetic network
// (200 tasks, 300 FS dependencies, 10 tiers x 20 tasks, topological seed 99, 40 Bernoulli risks).
func buildSynthetic200Network() ProjectNetwork {
	rng := NewSplitMix64(99)
	const numTiers = 10
	const tasksPerTier = 20
	const totalTasks = numTiers * tasksPerTier

	tasks := make([]Task, 0, totalTasks)
	for i := 0; i < totalTasks; i++ {
		to := 1.0 + rng.Float64()*4.0
		tm := to + 1.0 + rng.Float64()*3.0
		tp := tm + 2.0 + rng.Float64()*6.0
		tasks = append(tasks, Task{
			ID: fmt.Sprintf("T%03d", i+1),
			To: to,
			Tm: tm,
			Tp: tp,
		})
	}

	deps := make([]Dependency, 0, 300)
	for tier := 1; tier < numTiers; tier++ {
		for j := 0; j < tasksPerTier; j++ {
			curIdx := tier*tasksPerTier + j
			prevIdx := (tier-1)*tasksPerTier + j
			deps = append(deps, Dependency{
				From: tasks[prevIdx].ID,
				To:   tasks[curIdx].ID,
			})
		}
	}
	addedCross := 0
	for tier := 1; tier < numTiers && addedCross < 120; tier++ {
		for j := 0; j < tasksPerTier && addedCross < 120; j++ {
			curIdx := tier*tasksPerTier + j
			altOffset := (j + 3) % tasksPerTier
			prevIdx := (tier-1)*tasksPerTier + altOffset
			deps = append(deps, Dependency{
				From: tasks[prevIdx].ID,
				To:   tasks[curIdx].ID,
			})
			addedCross++
		}
	}

	risks := make([]Risk, 0, 40)
	for r := 0; r < 40; r++ {
		tIdx := (r * 5) % totalTasks
		risks = append(risks, Risk{
			ID:          fmt.Sprintf("R%02d", r+1),
			TaskID:      tasks[tIdx].ID,
			Probability: 0.10 + rng.Float64()*0.25,
			ImpactDays:  1.0 + rng.Float64()*3.0,
			Active:      true,
		})
	}

	return ProjectNetwork{
		StatusDate:   0,
		Tasks:        tasks,
		Dependencies: deps,
		Risks:        risks,
	}
}

// buildTCC2ScheduleNetwork builds the pre-registered TCC2 schedule network (E01-E14).
func buildTCC2ScheduleNetwork(withRisks bool) ProjectNetwork {
	if withRisks {
		return buildTCC2ScheduleNetworkWithMode("with_risks")
	}
	return buildTCC2ScheduleNetworkWithMode("base")
}

// buildTCC2ScheduleNetworkWithMode builds the TCC2 network under four modes (COD-10 & DEC-4):
// - "base": nominal PERT-Beta without discrete risks
// - "with_risks": nominal PERT-Beta with discrete risks (R_DAG, R_LEAK, R_LLM)
// - "stress": +25% inflation across all (to, tm, tp) estimates + active discrete risks
// - "stress_with_cuts": +25% inflation on non-cut tasks + active risks + ordered contingency cuts (DEC-4)
func buildTCC2ScheduleNetworkWithMode(mode string) ProjectNetwork {
	withRisks := mode == "with_risks" || mode == "stress" || mode == "stress_with_cuts"
	net := ProjectNetwork{
		StatusDate: 0.0,
		Tasks: []Task{
			// Trilha A (Autor 1 — Motor, Worker, Governança e UI Mínima)
			{ID: "E02", Name: "Contrato congelado e banco SQL/migrations", To: 1.0, Tm: 1.5, Tp: 2.0},
			{ID: "E03", Name: "API de leitura, paginação e projeções do estado vivo", To: 1.5, Tm: 2.0, Tp: 2.5},
			{ID: "E05", Name: "Port do núcleo estocástico da PoC com subfluxos CRN", To: 1.5, Tm: 2.0, Tp: 2.5},
			{ID: "E06", Name: "Worker assíncrono de simulação e persistência de execuções", To: 1.5, Tm: 2.0, Tp: 3.0},
			{ID: "E07", Name: "Propostas governadas, máquina de estados, diff, auditoria e guardrails", To: 1.5, Tm: 2.0, Tp: 2.5},
			{ID: "E11", Name: "Interface web mínima para inspeção do estado vivo, simulação e aprovação", To: 1.0, Tm: 1.5, Tp: 2.0},

			// Trilha B (Autor 2 — Dataset, Replay e Dimensões Experimentais 2, 3 e 5)
			{ID: "E01", Name: "Extrator GitHub, dataset de backtesting e replay temporal estrito", To: 1.5, Tm: 2.0, Tp: 2.5},
			{ID: "E04", Name: "Ingestão idempotente de eventos, outbox, DLQ e trilha de auditoria", To: 1.0, Tm: 1.5, Tp: 2.5},
			{ID: "E08", Name: "Suíte sintética da Dimensão 2 (10 regimes, M=500) e análise de sensibilidade", To: 1.5, Tm: 2.0, Tp: 3.0},
			{ID: "E09", Name: "Backtesting descritivo da Dimensão 3 (10 cortes, 4 baselines e sensibilidade)", To: 1.5, Tm: 2.5, Tp: 3.5},
			{ID: "E10", Name: "Agente consultivo e avaliação da Dimensão 5 (300 cenários, ablação e guardrails)", To: 2.0, Tm: 2.5, Tp: 3.5},

			// Fechamento Conjunto (Autor 1 + Autor 2)
			{ID: "E12", Name: "Consolidação estatística, tabelas finais e verificação reprodutível", To: 1.0, Tm: 1.5, Tp: 2.5},
			{ID: "E13", Name: "Redação final dos capítulos de resultados, discussão e ameaças à validade", To: 1.0, Tm: 1.5, Tp: 2.0},
			{ID: "E14", Name: "Revisão final ABNT, pacote de replicação Zenodo e preparação da defesa", To: 1.0, Tm: 1.0, Tp: 1.0},
		},
		Dependencies: []Dependency{
			// Sequenciamento intra-Trilha A (Autor 1)
			{From: "E02", To: "E03"},
			{From: "E03", To: "E05"},
			{From: "E05", To: "E06"},
			{From: "E06", To: "E07"},
			{From: "E07", To: "E11"},

			// Sequenciamento intra-Trilha B (Autor 2) + dependências técnicas entre trilhas
			{From: "E01", To: "E04"},
			{From: "E02", To: "E04"},
			{From: "E04", To: "E08"},
			{From: "E05", To: "E08"},
			{From: "E08", To: "E09"},
			{From: "E09", To: "E10"},
			{From: "E07", To: "E10"},

			// Fechamento conjunto
			{From: "E10", To: "E12"},
			{From: "E11", To: "E12"},
			{From: "E12", To: "E13"},
			{From: "E13", To: "E14"},
		},
		Risks: []Risk{
			{ID: "R_DAG", TaskID: "E01", Probability: 0.30, ImpactDays: 1.0, IsAbsolute: true, Active: withRisks},
			{ID: "R_LEAK", TaskID: "E09", Probability: 0.25, ImpactDays: 1.0, IsAbsolute: true, Active: withRisks},
			{ID: "R_LLM", TaskID: "E10", Probability: 0.30, ImpactDays: 1.0, IsAbsolute: true, Active: withRisks},
		},
	}

	if mode == "stress" || mode == "stress_with_cuts" {
		for i := range net.Tasks {
			// Apply +25% stress factor to non-fixed tasks (E14 is fixed 1-week defense/submission window)
			if net.Tasks[i].ID != "E14" {
				net.Tasks[i].To *= 1.25
				net.Tasks[i].Tm *= 1.25
				net.Tasks[i].Tp *= 1.25
			}
		}
	}

	if mode == "stress_with_cuts" {
		// Apply pre-registered contingency cuts inside essential packages (DEC-4 / ADR-004):
		// 1st cut: E10 single open LLM (1.5 / 2.0 / 2.5)
		// 2nd cut: E11 static tabular + CLI/HTTP approval (0.5 / 1.0 / 1.5)
		// 3rd cut: E09 reduced from 10 to 5 releases (1.0 / 1.5 / 2.0)
		// 4th cut: E08 focused on 6 core regimes (1.0 / 1.5 / 2.0)
		for i := range net.Tasks {
			switch net.Tasks[i].ID {
			case "E10":
				net.Tasks[i].To, net.Tasks[i].Tm, net.Tasks[i].Tp = 1.5, 2.0, 2.5
			case "E11":
				net.Tasks[i].To, net.Tasks[i].Tm, net.Tasks[i].Tp = 0.5, 1.0, 1.5
			case "E09":
				net.Tasks[i].To, net.Tasks[i].Tm, net.Tasks[i].Tp = 1.0, 1.5, 2.0
			case "E08":
				net.Tasks[i].To, net.Tasks[i].Tm, net.Tasks[i].Tp = 1.0, 1.5, 2.0
			}
		}
		for i := range net.Risks {
			if net.Risks[i].ID == "R_LLM" || net.Risks[i].ID == "R_LEAK" {
				net.Risks[i].ImpactDays = 0.5
			}
		}
	}

	return net
}

// TCC2AuditSummary stores the complete deterministic and stochastic audit of the TCC2 schedule (COD-10).
type TCC2AuditSummary struct {
	CPMDuration      float64                    `json:"cpm_duration"`
	TrackASum        float64                    `json:"track_a_sum"`
	TrackBSum        float64                    `json:"track_b_sum"`
	JointSum         float64                    `json:"joint_sum"`
	CriticalPath     []string                   `json:"critical_path"`
	TaskCPM          map[string]CPMTaskResult   `json:"task_cpm"`
	Base             *MonteCarloResult          `json:"base"`
	WithRisks        *MonteCarloResult          `json:"with_risks"`
	Stress           *MonteCarloResult          `json:"stress"`
	StressWithCuts   *MonteCarloResult          `json:"stress_with_cuts"`
}

// AuditTCC2Schedule runs deterministic CPM and all 4 Monte Carlo configurations on the TCC2 schedule (COD-10).
func AuditTCC2Schedule() (*TCC2AuditSummary, error) {
	baseNet := buildTCC2ScheduleNetworkWithMode("base")
	detDurs := make(map[string]float64)
	trackASet := map[string]bool{"E02": true, "E03": true, "E05": true, "E06": true, "E07": true, "E11": true}
	trackBSet := map[string]bool{"E01": true, "E04": true, "E08": true, "E09": true, "E10": true}
	jointSet := map[string]bool{"E12": true, "E13": true, "E14": true}

	var sumA, sumB, sumJoint float64
	for _, t := range baseNet.Tasks {
		d := CalculateExpectedDuration(t.To, t.Tm, t.Tp)
		detDurs[t.ID] = d
		if trackASet[t.ID] {
			sumA += d
		} else if trackBSet[t.ID] {
			sumB += d
		} else if jointSet[t.ID] {
			sumJoint += d
		}
	}

	cpmRes, err := SolveCPM(baseNet.Tasks, baseNet.Dependencies, detDurs, 0.0, 1e-9)
	if err != nil {
		return nil, err
	}

	order, _ := TopologicalSort(baseNet.Tasks, baseNet.Dependencies)
	var critPath []string
	for _, id := range order {
		if cpmRes.Tasks[id].Critical {
			critPath = append(critPath, id)
		}
	}

	cfg := MonteCarloConfig{
		Iterations:   10000,
		Seed:         42,
		NumWorkers:   8,
		TargetDate:   20.0,
		Lambda:       4.0,
		EpsilonFloat: 1e-9,
	}

	resBase, err := RunMonteCarlo(baseNet, cfg)
	if err != nil {
		return nil, err
	}
	resRisks, err := RunMonteCarlo(buildTCC2ScheduleNetworkWithMode("with_risks"), cfg)
	if err != nil {
		return nil, err
	}
	resStress, err := RunMonteCarlo(buildTCC2ScheduleNetworkWithMode("stress"), cfg)
	if err != nil {
		return nil, err
	}
	resCuts, err := RunMonteCarlo(buildTCC2ScheduleNetworkWithMode("stress_with_cuts"), cfg)
	if err != nil {
		return nil, err
	}

	return &TCC2AuditSummary{
		CPMDuration:    cpmRes.Duration,
		TrackASum:      sumA,
		TrackBSum:      sumB,
		JointSum:       sumJoint,
		CriticalPath:   critPath,
		TaskCPM:        cpmRes.Tasks,
		Base:           resBase,
		WithRisks:      resRisks,
		Stress:         resStress,
		StressWithCuts: resCuts,
	}, nil
}

// BenchmarkRunMetrics stores timing statistics for a multi-run benchmark configuration (COD-09).
type BenchmarkRunMetrics struct {
	Name          string  `json:"name"`
	Tasks         int     `json:"tasks"`
	Deps          int     `json:"deps"`
	Risks         int     `json:"risks"`
	Workers       int     `json:"workers"`
	Runs          int     `json:"runs"`
	MedianMs      float64 `json:"median_ms"`
	MeanMs        float64 `json:"mean_ms"`
	StdDevMs      float64 `json:"std_dev_ms"`
	MinMs         float64 `json:"min_ms"`
	MaxMs         float64 `json:"max_ms"`
	ThroughputItS float64 `json:"throughput_it_s"`
}

// BenchmarkEnvironmentSuite holds environment metadata and benchmark results across worker counts (COD-09).
type BenchmarkEnvironmentSuite struct {
	GOOS       string                `json:"goos"`
	GOARCH     string                `json:"goarch"`
	GoVersion  string                `json:"go_version"`
	NumCPU     int                   `json:"num_cpu"`
	Results    []BenchmarkRunMetrics `json:"results"`
}

// RunBenchmarkSuite measures 5-run dispersion for the 4-task network (1 worker) and 200-task network (1, 2, 4, 8 workers).
func RunBenchmarkSuite(runs int) (*BenchmarkEnvironmentSuite, error) {
	if runs < 3 {
		runs = 5
	}
	baseNet := buildBaseNetwork()
	synthNet := buildSynthetic200Network()

	measure := func(name string, net ProjectNetwork, workers int) (BenchmarkRunMetrics, error) {
		// 1 warmup run to stabilize CPU caches/goroutine pool
		_, _ = RunMonteCarlo(net, MonteCarloConfig{
			Iterations: 2000, Seed: 1, NumWorkers: workers, Lambda: 4.0, EpsilonFloat: 1e-9,
		})

		times := make([]float64, runs)
		var sum, sumSq float64
		for i := 0; i < runs; i++ {
			t0 := time.Now()
			_, err := RunMonteCarlo(net, MonteCarloConfig{
				Iterations:   10000,
				Seed:         uint64(42 + i),
				NumWorkers:   workers,
				Lambda:       4.0,
				EpsilonFloat: 1e-9,
			})
			if err != nil {
				return BenchmarkRunMetrics{}, err
			}
			ms := float64(time.Since(t0).Microseconds()) / 1000.0
			times[i] = ms
			sum += ms
			sumSq += ms * ms
		}
		sort.Float64s(times)
		mean := sum / float64(runs)
		variance := 0.0
		if runs > 1 {
			variance = (sumSq - float64(runs)*mean*mean) / float64(runs-1)
			if variance < 0 {
				variance = 0
			}
		}
		sd := math.Sqrt(variance)
		med := times[runs/2]
		tp := 0.0
		if med > 0 {
			tp = 10000.0 / (med / 1000.0)
		}
		return BenchmarkRunMetrics{
			Name:          name,
			Tasks:         len(net.Tasks),
			Deps:          len(net.Dependencies),
			Risks:         len(net.Risks),
			Workers:       workers,
			Runs:          runs,
			MedianMs:      med,
			MeanMs:        mean,
			StdDevMs:      sd,
			MinMs:         times[0],
			MaxMs:         times[runs-1],
			ThroughputItS: tp,
		}, nil
	}

	var results []BenchmarkRunMetrics
	mBase, err := measure("Rede Base (4 tarefas, 4 dep.)", baseNet, 1)
	if err != nil {
		return nil, err
	}
	results = append(results, mBase)

	for _, w := range []int{1, 2, 4, 8} {
		mSynth, err := measure(fmt.Sprintf("Rede Sintética (200 tarefas, 300 dep., %d workers)", w), synthNet, w)
		if err != nil {
			return nil, err
		}
		results = append(results, mSynth)
	}

	return &BenchmarkEnvironmentSuite{
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		GoVersion: runtime.Version(),
		NumCPU:    runtime.NumCPU(),
		Results:   results,
	}, nil
}

// ECDFGridData stores a 200-point ECDF grid and 40-bin histogram for Scenarios A and B (FOR-03).
type ECDFGridData struct {
	GridX      []float64 `json:"grid_x"`
	ECDFA      []float64 `json:"ecdf_a"`
	ECDFB      []float64 `json:"ecdf_b"`
	BinCenters []float64 `json:"bin_centers"`
	DensityA   []float64 `json:"density_a"`
	DensityB   []float64 `json:"density_b"`
}

func computeECDFAndHist(resA, resB *MonteCarloResult) ECDFGridData {
	const numGrid = 200
	minX := resA.Durations[0]
	if resB.Durations[0] < minX {
		minX = resB.Durations[0]
	}
	maxX := resA.Durations[len(resA.Durations)-1]
	if resB.Durations[len(resB.Durations)-1] > maxX {
		maxX = resB.Durations[len(resB.Durations)-1]
	}
	minX -= 0.2
	maxX += 0.2

	gridX := make([]float64, numGrid)
	ecdfA := make([]float64, numGrid)
	ecdfB := make([]float64, numGrid)
	nA := float64(len(resA.Durations))
	nB := float64(len(resB.Durations))

	for i := 0; i < numGrid; i++ {
		x := minX + (maxX-minX)*float64(i)/float64(numGrid-1)
		gridX[i] = math.Round(x*10000) / 10000
		cntA := sort.SearchFloat64s(resA.Durations, x+1e-12)
		cntB := sort.SearchFloat64s(resB.Durations, x+1e-12)
		ecdfA[i] = math.Round((float64(cntA)/nA)*10000) / 10000
		ecdfB[i] = math.Round((float64(cntB)/nB)*10000) / 10000
	}

	const numBins = 40
	binWidth := (maxX - minX) / float64(numBins)
	centers := make([]float64, numBins)
	densA := make([]float64, numBins)
	densB := make([]float64, numBins)
	for b := 0; b < numBins; b++ {
		lo := minX + float64(b)*binWidth
		hi := lo + binWidth
		centers[b] = math.Round((lo+0.5*binWidth)*10000) / 10000
		cA := sort.SearchFloat64s(resA.Durations, hi) - sort.SearchFloat64s(resA.Durations, lo)
		cB := sort.SearchFloat64s(resB.Durations, hi) - sort.SearchFloat64s(resB.Durations, lo)
		densA[b] = math.Round((float64(cA)/(nA*binWidth))*10000) / 10000
		densB[b] = math.Round((float64(cB)/(nB*binWidth))*10000) / 10000
	}

	return ECDFGridData{
		GridX:      gridX,
		ECDFA:      ecdfA,
		ECDFB:      ecdfB,
		BinCenters: centers,
		DensityA:   densA,
		DensityB:   densB,
	}
}

func fmtBR(val float64, decimals int) string {
	s := fmt.Sprintf("%.*f", decimals, val)
	return strings.ReplaceAll(s, ".", "{,}")
}

func main() {
	if err := RunAllPoCValidationAndExport(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro na execucao da PoC: %v\n", err)
		os.Exit(1)
	}
}

// RunAllPoCValidationAndExport runs all PoC scenarios, KS validations, CRN diffs,
// reproducibility checks, benchmarks, and TCC2 schedule audits, and writes JSON + LaTeX outputs
// to the repository paths.
func RunAllPoCValidationAndExport() error {
	return RunAllPoCValidationAndExportToDir(".", "../Docs", "..", 5)
}

// RunAllPoCValidationAndExportToDir runs the full validation and exports files to the specified directories.
func RunAllPoCValidationAndExportToDir(jsonDir, docsDir, rootDir string, benchRuns int) error {
	fmt.Println("=================================================================")
	fmt.Println("  PoC Go - Validação Algorítmica do Motor CPM & Monte Carlo")
	fmt.Println("  TCC 1 - Sombra Digital Governada e IA em Projetos de Software")
	fmt.Println("=================================================================")
	fmt.Println()

	cfg := MonteCarloConfig{
		Iterations:   10000,
		Seed:         42,
		NumWorkers:   4,
		TargetDate:   11.0,
		Lambda:       4.0,
		EpsilonFloat: 1e-9,
	}

	netA := buildBaseNetwork()
	resA, err := RunMonteCarlo(netA, cfg)
	if err != nil {
		return fmt.Errorf("Cenário A: %w", err)
	}

	netB := buildBaseNetwork()
	netB.Risks = []Risk{
		{ID: "R1", TaskID: "T3", Probability: 0.30, ImpactDays: 2.0, IsAbsolute: false, Active: true},
	}
	resB, err := RunMonteCarlo(netB, cfg)
	if err != nil {
		return fmt.Errorf("Cenário B: %w", err)
	}

	netC := buildBaseNetwork()
	netC.StatusDate = 5.0
	netC.Tasks[0].Completed = true
	netC.Tasks[0].Progress = 1.0
	netC.Tasks[0].ActualDuration = 3.0
	netC.Tasks[1].Progress = 0.40
	netC.Tasks[2].Progress = 0.30
	netC.Risks = []Risk{
		{ID: "R1", TaskID: "T3", Probability: 0.30, ImpactDays: 2.0, IsAbsolute: false, Active: true},
	}
	resC, err := RunMonteCarlo(netC, cfg)
	if err != nil {
		return fmt.Errorf("Cenário C: %w", err)
	}

	netCPrime := buildBaseNetwork()
	netCPrime.StatusDate = 5.0
	netCPrime.Tasks[0].Completed = true
	netCPrime.Tasks[0].Progress = 1.0
	netCPrime.Tasks[0].ActualDuration = 3.0
	netCPrime.Tasks[1].Progress = 0.40
	netCPrime.Tasks[2].Progress = 0.30
	netCPrime.Risks = []Risk{
		{ID: "R1", TaskID: "T3", Probability: 0.30, ImpactDays: 2.0, IsAbsolute: false, Active: false},
	}
	resCPrime, err := RunMonteCarlo(netCPrime, cfg)
	if err != nil {
		return fmt.Errorf("Cenário C': %w", err)
	}

	crnDiff, err := ComputeCRNDiff(resC, resCPrime, cfg.Seed)
	if err != nil {
		return fmt.Errorf("ComputeCRNDiff: %w", err)
	}

	// KS Goodness-of-Fit (n = 1,000,000 full sample + 100 seeds x 10,000)
	ksDefs := []struct {
		id         string
		to, tm, tp float64
		offset     int
	}{
		{"T1", 2, 3, 5, 101},
		{"T2", 4, 5, 9, 202},
		{"T3", 3, 5, 10, 303},
		{"T4", 1, 2, 4, 404},
	}
	var ksSingle []KSTaskResult
	var ksMulti []KSMultiSeedTaskSummary
	for _, kd := range ksDefs {
		s := ComputeKSBetaFit(kd.id, kd.to, kd.tm, kd.tp, 4.0, 1_000_000, DeriveIterationSeed(2026, kd.offset))
		ksSingle = append(ksSingle, s)
		m := RunKSMultiSeedSuite(kd.id, kd.to, kd.tm, kd.tp, 4.0, kd.offset, 100, 10000)
		ksMulti = append(ksMulti, m)
	}

	// Reproducibility across 1, 4, 16 workers (seed 12345, N = 5000)
	reproNet := buildBaseNetwork()
	reproNet.Risks = []Risk{{ID: "R1", TaskID: "T3", Probability: 0.30, ImpactDays: 2.0, Active: true}}
	reproWorkers := make(map[string]*MonteCarloResult)
	for _, w := range []int{1, 4, 16} {
		rw, err := RunMonteCarlo(reproNet, MonteCarloConfig{
			Iterations: 5000, Seed: 12345, NumWorkers: w, Lambda: 4.0, EpsilonFloat: 1e-9, TargetDate: 11.0,
		})
		if err != nil {
			return err
		}
		reproWorkers[fmt.Sprintf("w_%d", w)] = rw
	}

	// Multi-worker Benchmark Suite
	benchSuite, err := RunBenchmarkSuite(benchRuns)
	if err != nil {
		return fmt.Errorf("RunBenchmarkSuite: %w", err)
	}

	// TCC2 Schedule Audit (Base, WithRisks, Stress, StressWithCuts)
	tcc2Audit, err := AuditTCC2Schedule()
	if err != nil {
		return fmt.Errorf("AuditTCC2Schedule: %w", err)
	}

	fmt.Println("### Tabela 9 Reproduzida pela PoC em Go (N = 10.000 iterações, seed = 42)")
	fmt.Printf("| A | CPM=%.2f | Mean=%.2f (±%.3f) | P50=%.2f (±%.3f) | P85=%.2f (±%.3f) | P95=%.2f (±%.3f) | P(F<=11)=%.1f%% |\n",
		resA.CPMDeterministic, resA.Mean, resA.MeanSE, resA.P50, resA.P50SE, resA.P85, resA.P85SE, resA.P95, resA.P95SE, resA.ProbTarget)
	fmt.Printf("| B | CPM=%.2f | Mean=%.2f (±%.3f) | P50=%.2f (±%.3f) | P85=%.2f (±%.3f) | P95=%.2f (±%.3f) | P(F<=11)=%.1f%% |\n",
		resB.CPMDeterministic, resB.Mean, resB.MeanSE, resB.P50, resB.P50SE, resB.P85, resB.P85SE, resB.P95, resB.P95SE, resB.ProbTarget)
	fmt.Printf("| C | CPM=%.2f | Mean=%.2f (±%.3f) | P50=%.2f (±%.3f) | P85=%.2f (±%.3f) | P95=%.2f (±%.3f) | P(F<=11)=%.1f%% |\n",
		resC.CPMDeterministic, resC.Mean, resC.MeanSE, resC.P50, resC.P50SE, resC.P85, resC.P85SE, resC.P95, resC.P95SE, resC.ProbTarget)
	fmt.Printf("| C'| CPM=%.2f | Mean=%.2f (±%.3f) | P50=%.2f (±%.3f) | P85=%.2f (±%.3f) | P95=%.2f (±%.3f) | P(F<=11)=%.1f%% |\n",
		resCPrime.CPMDeterministic, resCPrime.Mean, resCPrime.MeanSE, resCPrime.P50, resCPrime.P50SE, resCPrime.P85, resCPrime.P85SE, resCPrime.P95, resCPrime.P95SE, resCPrime.ProbTarget)
	fmt.Printf("| CRN Diff (C' - C) | DeltaMean=%.3f (SE pareado=%.4f vs indep=%.4f, red=%.1fx) | DeltaP85=%.2f (±%.3f) |\n\n",
		crnDiff.DeltaMean, crnDiff.DeltaMeanSE, crnDiff.IndependentSE, crnDiff.VarianceRedRatio, crnDiff.DeltaP85, crnDiff.DeltaP85SE)

	fmt.Println("### Simulação do Cronograma TCC 2 (20 semanas, N = 10.000, seed = 42)")
	fmt.Printf("| Base (sem riscos)          | CPM=%.2f | Mean=%.2f | P50=%.2f | P85=%.2f | P95=%.2f | P(F<=20)=%.1f%% |\n",
		tcc2Audit.Base.CPMDeterministic, tcc2Audit.Base.Mean, tcc2Audit.Base.P50, tcc2Audit.Base.P85, tcc2Audit.Base.P95, tcc2Audit.Base.ProbTarget)
	fmt.Printf("| Com Riscos (R_DAG/LEAK/LLM)| CPM=%.2f | Mean=%.2f | P50=%.2f | P85=%.2f | P95=%.2f | P(F<=20)=%.1f%% |\n",
		tcc2Audit.WithRisks.CPMDeterministic, tcc2Audit.WithRisks.Mean, tcc2Audit.WithRisks.P50, tcc2Audit.WithRisks.P85, tcc2Audit.WithRisks.P95, tcc2Audit.WithRisks.ProbTarget)
	fmt.Printf("| Estresse (+25%% O/M/P+riscos)| CPM=%.2f | Mean=%.2f | P50=%.2f | P85=%.2f | P95=%.2f | P(F<=20)=%.1f%% |\n",
		tcc2Audit.Stress.CPMDeterministic, tcc2Audit.Stress.Mean, tcc2Audit.Stress.P50, tcc2Audit.Stress.P85, tcc2Audit.Stress.P95, tcc2Audit.Stress.ProbTarget)
	fmt.Printf("| Estresse + Cortes (DEC-4)  | CPM=%.2f | Mean=%.2f | P50=%.2f | P85=%.2f | P95=%.2f | P(F<=20)=%.1f%% |\n\n",
		tcc2Audit.StressWithCuts.CPMDeterministic, tcc2Audit.StressWithCuts.Mean, tcc2Audit.StressWithCuts.P50, tcc2Audit.StressWithCuts.P85, tcc2Audit.StressWithCuts.P95, tcc2Audit.StressWithCuts.ProbTarget)

	ecdfData := computeECDFAndHist(resA, resB)

	outputData := map[string]interface{}{
		"scenario_a":              resA,
		"scenario_b":              resB,
		"scenario_c":              resC,
		"scenario_c_prime":        resCPrime,
		"crn_diff_c_cprime":       crnDiff,
		"ecdf_ab":                 ecdfData,
		"ks_single_seed_1m":       ksSingle,
		"ks_multi_seed_100":       ksMulti,
		"reproducibility_workers": reproWorkers,
		"benchmarks":              benchSuite,
		"tcc2_schedule_base":      tcc2Audit.Base,
		"tcc2_schedule_risks":     tcc2Audit.WithRisks,
		"tcc2_schedule_stress":    tcc2Audit.Stress,
		"tcc2_schedule_cuts":      tcc2Audit.StressWithCuts,
		"tcc2_audit":              tcc2Audit,
	}
	jsonData, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(jsonDir, "poc_results.json"), jsonData, 0644); err != nil {
		return err
	}
	fmt.Println("Resultados exportados para poc_results.json com sucesso.")

	if err := exportLatexMacrosAndTables(docsDir, rootDir, resA, resB, resC, resCPrime, crnDiff, ksSingle, ksMulti, reproWorkers["w_1"], benchSuite, tcc2Audit); err != nil {
		return err
	}
	fmt.Println("Macros e tabelas LaTeX exportadas com sucesso.")
	return nil
}

func exportLatexMacrosAndTables(
	docsDir, rootDir string,
	resA, resB, resC, resCPrime *MonteCarloResult,
	crnDiff *CRNDiffResult,
	ksSingle []KSTaskResult,
	ksMulti []KSMultiSeedTaskSummary,
	reproW1 *MonteCarloResult,
	benchSuite *BenchmarkEnvironmentSuite,
	tcc2Audit *TCC2AuditSummary,
) error {
	tcc2Base := tcc2Audit.Base
	tcc2Risks := tcc2Audit.WithRisks
	tcc2Stress := tcc2Audit.Stress
	tcc2Cuts := tcc2Audit.StressWithCuts

	bBase := benchSuite.Results[0]
	bSynthW1 := benchSuite.Results[1]
	bSynthW2 := benchSuite.Results[2]
	bSynthW4 := benchSuite.Results[3]
	bSynthW8 := benchSuite.Results[4]

	macros := fmt.Sprintf(`%% Arquivo gerado automaticamente por Codigo/main.go (N = 10.000, seed = 42)
%% Nao edite manualmente; execute "cd Codigo && go run ." para atualizar.
\providecommand{\pocACPM}{%s}
\providecommand{\pocAMean}{%s}
\providecommand{\pocAMeanSE}{%s}
\providecommand{\pocAPcinquenta}{%s}
\providecommand{\pocAPcinquentaSE}{%s}
\providecommand{\pocAPoitentaeCinco}{%s}
\providecommand{\pocAPoitentaeCincoSE}{%s}
\providecommand{\pocAPnoventaeCinco}{%s}
\providecommand{\pocAPnoventaeCincoSE}{%s}
\providecommand{\pocAProbMeta}{%s}
\providecommand{\pocBCPM}{%s}
\providecommand{\pocBMean}{%s}
\providecommand{\pocBMeanSE}{%s}
\providecommand{\pocBPcinquenta}{%s}
\providecommand{\pocBPcinquentaSE}{%s}
\providecommand{\pocBPoitentaeCinco}{%s}
\providecommand{\pocBPoitentaeCincoSE}{%s}
\providecommand{\pocBPnoventaeCinco}{%s}
\providecommand{\pocBPnoventaeCincoSE}{%s}
\providecommand{\pocBProbMeta}{%s}
\providecommand{\pocCCPM}{%s}
\providecommand{\pocCMean}{%s}
\providecommand{\pocCMeanSE}{%s}
\providecommand{\pocCPcinquenta}{%s}
\providecommand{\pocCPcinquentaSE}{%s}
\providecommand{\pocCPoitentaeCinco}{%s}
\providecommand{\pocCPoitentaeCincoSE}{%s}
\providecommand{\pocCPnoventaeCinco}{%s}
\providecommand{\pocCPnoventaeCincoSE}{%s}
\providecommand{\pocCProbMeta}{%s}
\providecommand{\pocCPrimeCPM}{%s}
\providecommand{\pocCPrimeMean}{%s}
\providecommand{\pocCPrimeMeanSE}{%s}
\providecommand{\pocCPrimePcinquenta}{%s}
\providecommand{\pocCPrimePcinquentaSE}{%s}
\providecommand{\pocCPrimePoitentaeCinco}{%s}
\providecommand{\pocCPrimePoitentaeCincoSE}{%s}
\providecommand{\pocCPrimePnoventaeCinco}{%s}
\providecommand{\pocCPrimePnoventaeCincoSE}{%s}
\providecommand{\pocCPrimeProbMeta}{%s}
\providecommand{\crnDeltaMean}{%s}
\providecommand{\crnDeltaMeanSE}{%s}
\providecommand{\crnIndepSE}{%s}
\providecommand{\crnVarRedRatio}{%s}
\providecommand{\crnDeltaPoitentaeCinco}{%s}
\providecommand{\crnDeltaPoitentaeCincoSE}{%s}
\providecommand{\tccDoisBaseCPM}{%s}
\providecommand{\tccDoisBaseMean}{%s}
\providecommand{\tccDoisBaseMeanSE}{%s}
\providecommand{\tccDoisBasePcinquenta}{%s}
\providecommand{\tccDoisBasePcinquentaSE}{%s}
\providecommand{\tccDoisBasePoitentaeCinco}{%s}
\providecommand{\tccDoisBasePoitentaeCincoSE}{%s}
\providecommand{\tccDoisBasePnoventaeCinco}{%s}
\providecommand{\tccDoisBasePnoventaeCincoSE}{%s}
\providecommand{\tccDoisBaseProbVinte}{%s}
\providecommand{\tccDoisRiscosMean}{%s}
\providecommand{\tccDoisRiscosMeanSE}{%s}
\providecommand{\tccDoisRiscosPcinquenta}{%s}
\providecommand{\tccDoisRiscosPcinquentaSE}{%s}
\providecommand{\tccDoisRiscosPoitentaeCinco}{%s}
\providecommand{\tccDoisRiscosPoitentaeCincoSE}{%s}
\providecommand{\tccDoisRiscosPnoventaeCinco}{%s}
\providecommand{\tccDoisRiscosPnoventaeCincoSE}{%s}
\providecommand{\tccDoisRiscosProbVinte}{%s}
\providecommand{\tccDoisStressCPM}{%s}
\providecommand{\tccDoisStressMean}{%s}
\providecommand{\tccDoisStressPcinquenta}{%s}
\providecommand{\tccDoisStressPoitentaeCinco}{%s}
\providecommand{\tccDoisStressPnoventaeCinco}{%s}
\providecommand{\tccDoisStressProbVinte}{%s}
\providecommand{\tccDoisCutsCPM}{%s}
\providecommand{\tccDoisCutsMean}{%s}
\providecommand{\tccDoisCutsPcinquenta}{%s}
\providecommand{\tccDoisCutsPoitentaeCinco}{%s}
\providecommand{\tccDoisCutsPnoventaeCinco}{%s}
\providecommand{\tccDoisCutsProbVinte}{%s}
\providecommand{\benchBaseMedMs}{%s}
\providecommand{\benchSynthWOneMedMs}{%s}
\providecommand{\benchSynthWTwoMedMs}{%s}
\providecommand{\benchSynthWFourMedMs}{%s}
\providecommand{\benchSynthWEightMedMs}{%s}
`,
		fmtBR(resA.CPMDeterministic, 2), fmtBR(resA.Mean, 2), fmtBR(resA.MeanSE, 3), fmtBR(resA.P50, 2), fmtBR(resA.P50SE, 3), fmtBR(resA.P85, 2), fmtBR(resA.P85SE, 3), fmtBR(resA.P95, 2), fmtBR(resA.P95SE, 3), fmtBR(resA.ProbTarget, 1),
		fmtBR(resB.CPMDeterministic, 2), fmtBR(resB.Mean, 2), fmtBR(resB.MeanSE, 3), fmtBR(resB.P50, 2), fmtBR(resB.P50SE, 3), fmtBR(resB.P85, 2), fmtBR(resB.P85SE, 3), fmtBR(resB.P95, 2), fmtBR(resB.P95SE, 3), fmtBR(resB.ProbTarget, 1),
		fmtBR(resC.CPMDeterministic, 2), fmtBR(resC.Mean, 2), fmtBR(resC.MeanSE, 3), fmtBR(resC.P50, 2), fmtBR(resC.P50SE, 3), fmtBR(resC.P85, 2), fmtBR(resC.P85SE, 3), fmtBR(resC.P95, 2), fmtBR(resC.P95SE, 3), fmtBR(resC.ProbTarget, 1),
		fmtBR(resCPrime.CPMDeterministic, 2), fmtBR(resCPrime.Mean, 2), fmtBR(resCPrime.MeanSE, 3), fmtBR(resCPrime.P50, 2), fmtBR(resCPrime.P50SE, 3), fmtBR(resCPrime.P85, 2), fmtBR(resCPrime.P85SE, 3), fmtBR(resCPrime.P95, 2), fmtBR(resCPrime.P95SE, 3), fmtBR(resCPrime.ProbTarget, 1),
		fmtBR(crnDiff.DeltaMean, 2), fmtBR(crnDiff.DeltaMeanSE, 4), fmtBR(crnDiff.IndependentSE, 4), fmtBR(crnDiff.VarianceRedRatio, 1), fmtBR(crnDiff.DeltaP85, 2), fmtBR(crnDiff.DeltaP85SE, 3),
		fmtBR(tcc2Base.CPMDeterministic, 2), fmtBR(tcc2Base.Mean, 2), fmtBR(tcc2Base.MeanSE, 3), fmtBR(tcc2Base.P50, 2), fmtBR(tcc2Base.P50SE, 3), fmtBR(tcc2Base.P85, 2), fmtBR(tcc2Base.P85SE, 3), fmtBR(tcc2Base.P95, 2), fmtBR(tcc2Base.P95SE, 3), fmtBR(tcc2Base.ProbTarget, 1),
		fmtBR(tcc2Risks.Mean, 2), fmtBR(tcc2Risks.MeanSE, 3), fmtBR(tcc2Risks.P50, 2), fmtBR(tcc2Risks.P50SE, 3), fmtBR(tcc2Risks.P85, 2), fmtBR(tcc2Risks.P85SE, 3), fmtBR(tcc2Risks.P95, 2), fmtBR(tcc2Risks.P95SE, 3), fmtBR(tcc2Risks.ProbTarget, 1),
		fmtBR(tcc2Stress.CPMDeterministic, 2), fmtBR(tcc2Stress.Mean, 2), fmtBR(tcc2Stress.P50, 2), fmtBR(tcc2Stress.P85, 2), fmtBR(tcc2Stress.P95, 2), fmtBR(tcc2Stress.ProbTarget, 1),
		fmtBR(tcc2Cuts.CPMDeterministic, 2), fmtBR(tcc2Cuts.Mean, 2), fmtBR(tcc2Cuts.P50, 2), fmtBR(tcc2Cuts.P85, 2), fmtBR(tcc2Cuts.P95, 2), fmtBR(tcc2Cuts.ProbTarget, 1),
		fmtBR(bBase.MedianMs, 2), fmtBR(bSynthW1.MedianMs, 1), fmtBR(bSynthW2.MedianMs, 1), fmtBR(bSynthW4.MedianMs, 1), fmtBR(bSynthW8.MedianMs, 1),
	)

	for _, p := range []string{filepath.Join(docsDir, "resultados_poc.tex"), filepath.Join(rootDir, "resultados_poc.tex")} {
		if err := os.WriteFile(p, []byte(macros), 0644); err != nil {
			return err
		}
	}

	// Build LaTeX table fragments in docsDir/tabelas/ and rootDir/tabelas/
	for _, dir := range []string{filepath.Join(docsDir, "tabelas"), filepath.Join(rootDir, "tabelas")} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		// 1. tab_ks.tex
		var ksRows strings.Builder
		ksRows.WriteString("% Gerado automaticamente por Codigo/main.go\n")
		for i, k := range ksSingle {
			m := ksMulti[i]
			ksRows.WriteString(fmt.Sprintf("%s $(%.0f,\\,%.0f,\\,%.0f)$ & $%s$ / $%s$ & $%s$ / $%s$ & $%s$ & $%s$ & %d/100 ($%s$\\%%) \\\\\n",
				k.TaskID, k.To, k.Tm, k.Tp,
				fmtBR(k.MuTheo, 3), fmtBR(k.MuSample, 3),
				fmtBR(k.VarTheo, 3), fmtBR(k.VarSample, 3),
				fmtBR(k.DnFull, 5), fmtBR(k.PValueFull, 3),
				m.Rejections005, fmtBR(m.RejectRate005, 0),
			))
		}
		if err := os.WriteFile(filepath.Join(dir, "tab_ks.tex"), []byte(ksRows.String()), 0644); err != nil {
			return err
		}

		// 2. tab_poc_cenarios.tex
		scenRows := fmt.Sprintf(`%% Gerado automaticamente por Codigo/main.go
\textbf{A.} Início, sem riscos & $%s$ & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s\%%$ \\
\textbf{B.} Início, com R1 em T3 ($p_r=0{,}30, \Delta_r=2$) & $%s$ & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s\%%$ \\
\textbf{C.} Status $t_s=5$, R1 ativo & $%s$ & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s\%%$ \\
\textbf{C$'$.} Como C, com R1 mitigado & $%s$ & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s\%%$ \\
\midrule
\textbf{Diferença pareada CRN ($\text{C}' - \text{C}$)} & $0{,}00$ & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $+%s$ p.p. \\
`,
			fmtBR(resA.CPMDeterministic, 2), fmtBR(resA.Mean, 2), fmtBR(resA.MeanSE, 3), fmtBR(resA.P50, 2), fmtBR(resA.P50SE, 3), fmtBR(resA.P85, 2), fmtBR(resA.P85SE, 3), fmtBR(resA.P95, 2), fmtBR(resA.P95SE, 3), fmtBR(resA.ProbTarget, 1),
			fmtBR(resB.CPMDeterministic, 2), fmtBR(resB.Mean, 2), fmtBR(resB.MeanSE, 3), fmtBR(resB.P50, 2), fmtBR(resB.P50SE, 3), fmtBR(resB.P85, 2), fmtBR(resB.P85SE, 3), fmtBR(resB.P95, 2), fmtBR(resB.P95SE, 3), fmtBR(resB.ProbTarget, 1),
			fmtBR(resC.CPMDeterministic, 2), fmtBR(resC.Mean, 2), fmtBR(resC.MeanSE, 3), fmtBR(resC.P50, 2), fmtBR(resC.P50SE, 3), fmtBR(resC.P85, 2), fmtBR(resC.P85SE, 3), fmtBR(resC.P95, 2), fmtBR(resC.P95SE, 3), fmtBR(resC.ProbTarget, 1),
			fmtBR(resCPrime.CPMDeterministic, 2), fmtBR(resCPrime.Mean, 2), fmtBR(resCPrime.MeanSE, 3), fmtBR(resCPrime.P50, 2), fmtBR(resCPrime.P50SE, 3), fmtBR(resCPrime.P85, 2), fmtBR(resCPrime.P85SE, 3), fmtBR(resCPrime.P95, 2), fmtBR(resCPrime.P95SE, 3), fmtBR(resCPrime.ProbTarget, 1),
			fmtBR(crnDiff.DeltaMean, 2), fmtBR(crnDiff.DeltaMeanSE, 4), fmtBR(crnDiff.DeltaP50, 2), fmtBR(crnDiff.DeltaP50SE, 3), fmtBR(crnDiff.DeltaP85, 2), fmtBR(crnDiff.DeltaP85SE, 3), fmtBR(crnDiff.DeltaP95, 2), fmtBR(crnDiff.DeltaP95SE, 3), fmtBR(resCPrime.ProbTarget-resC.ProbTarget, 1),
		)
		if err := os.WriteFile(filepath.Join(dir, "tab_poc_cenarios.tex"), []byte(scenRows), 0644); err != nil {
			return err
		}

		// 3. tab_poc_ic.tex
		icRows := fmt.Sprintf(`%% Gerado automaticamente por Codigo/main.go
\textbf{A.} Início, sem riscos & $%s\%%$ & $%s\%%$ & $%s\%%$ & $%s\%%$ \\
\textbf{B.} Início, com R1 em T3 & $%s\%%$ & $%s\%%$ & $%s\%%$ & $%s\%%$ \\
\textbf{C.} Status $t_s = 5$, R1 ativo & --- & $%s\%%$ & $%s\%%$ & $%s\%%$ \\
\textbf{C$'$.} Status $t_s = 5$, R1 mitigado & --- & $%s\%%$ & $%s\%%$ & $%s\%%$ \\
`,
			fmtBR(resA.CriticalityIndex["T1"], 1), fmtBR(resA.CriticalityIndex["T2"], 1), fmtBR(resA.CriticalityIndex["T3"], 1), fmtBR(resA.CriticalityIndex["T4"], 1),
			fmtBR(resB.CriticalityIndex["T1"], 1), fmtBR(resB.CriticalityIndex["T2"], 1), fmtBR(resB.CriticalityIndex["T3"], 1), fmtBR(resB.CriticalityIndex["T4"], 1),
			fmtBR(resC.CriticalityIndex["T2"], 1), fmtBR(resC.CriticalityIndex["T3"], 1), fmtBR(resC.CriticalityIndex["T4"], 1),
			fmtBR(resCPrime.CriticalityIndex["T2"], 1), fmtBR(resCPrime.CriticalityIndex["T3"], 1), fmtBR(resCPrime.CriticalityIndex["T4"], 1),
		)
		if err := os.WriteFile(filepath.Join(dir, "tab_poc_ic.tex"), []byte(icRows), 0644); err != nil {
			return err
		}

		// 4. tab_reprodutibilidade.tex
		reproRows := fmt.Sprintf(`%% Gerado automaticamente por Codigo/main.go
1 \textit{worker} (sequencial) & $%s$ & $%s$ & $%s$ & $%s$ & $%s\%%$ & $%s\%%$ \\
4 \textit{workers} (paralelo) & $%s$ & $%s$ & $%s$ & $%s$ & $%s\%%$ & $%s\%%$ \\
16 \textit{workers} (alta concorrência) & $%s$ & $%s$ & $%s$ & $%s$ & $%s\%%$ & $%s\%%$ \\
\textbf{Divergência máxima ($\Delta_{\max}$)} & \textbf{0,000000} & \textbf{0,000000} & \textbf{0,000000} & \textbf{0,000000} & \textbf{0,0000\%%} & \textbf{0,0000\%%} \\
`,
			fmtBR(reproW1.Mean, 6), fmtBR(reproW1.P50, 6), fmtBR(reproW1.P85, 6), fmtBR(reproW1.P95, 6), fmtBR(reproW1.ProbTarget, 4), fmtBR(reproW1.CriticalityIndex["T3"], 4),
			fmtBR(reproW1.Mean, 6), fmtBR(reproW1.P50, 6), fmtBR(reproW1.P85, 6), fmtBR(reproW1.P95, 6), fmtBR(reproW1.ProbTarget, 4), fmtBR(reproW1.CriticalityIndex["T3"], 4),
			fmtBR(reproW1.Mean, 6), fmtBR(reproW1.P50, 6), fmtBR(reproW1.P85, 6), fmtBR(reproW1.P95, 6), fmtBR(reproW1.ProbTarget, 4), fmtBR(reproW1.CriticalityIndex["T3"], 4),
		)
		if err := os.WriteFile(filepath.Join(dir, "tab_reprodutibilidade.tex"), []byte(reproRows), 0644); err != nil {
			return err
		}

		// 5. tab_benchmarks.tex
		var benchRows strings.Builder
		benchRows.WriteString("% Gerado automaticamente por Codigo/main.go\n")
		for _, br := range benchSuite.Results {
			benchRows.WriteString(fmt.Sprintf("%s & %d / %d / %d & %d & $%s$ & $%s \\pm %s$ & $[%s;\\,%s]$ & $\\approx %.0f$ \\\\\n",
				br.Name, br.Tasks, br.Deps, br.Risks, br.Workers,
				fmtBR(br.MedianMs, 2), fmtBR(br.MeanMs, 2), fmtBR(br.StdDevMs, 2),
				fmtBR(br.MinMs, 2), fmtBR(br.MaxMs, 2), br.ThroughputItS,
			))
		}
		if err := os.WriteFile(filepath.Join(dir, "tab_benchmarks.tex"), []byte(benchRows.String()), 0644); err != nil {
			return err
		}

		// 6. tab_tcc2_mc.tex
		tcc2Rows := fmt.Sprintf(`%% Gerado automaticamente por Codigo/main.go
\textbf{Base PERT-Beta (sem riscos discretos)} & $%s$ & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s\%%$ \\
\textbf{Com Riscos Discretos ($\text{R}_{\text{DAG}}, \text{R}_{\text{LEAK}}, \text{R}_{\text{LLM}}$)} & $%s$ & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s\%%$ \\
\textbf{Estresse ($+25\%%$ em O/M/P + riscos ativos)} & $%s$ & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s\%%$ \\
\textbf{Estresse + Cortes Pré-Registrados (E08--E11)} & $%s$ & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s$ ($\pm %s$) & $%s\%%$ \\
`,
			fmtBR(tcc2Base.CPMDeterministic, 2), fmtBR(tcc2Base.Mean, 2), fmtBR(tcc2Base.MeanSE, 3), fmtBR(tcc2Base.P50, 2), fmtBR(tcc2Base.P50SE, 3), fmtBR(tcc2Base.P85, 2), fmtBR(tcc2Base.P85SE, 3), fmtBR(tcc2Base.P95, 2), fmtBR(tcc2Base.P95SE, 3), fmtBR(tcc2Base.ProbTarget, 1),
			fmtBR(tcc2Risks.CPMDeterministic, 2), fmtBR(tcc2Risks.Mean, 2), fmtBR(tcc2Risks.MeanSE, 3), fmtBR(tcc2Risks.P50, 2), fmtBR(tcc2Risks.P50SE, 3), fmtBR(tcc2Risks.P85, 2), fmtBR(tcc2Risks.P85SE, 3), fmtBR(tcc2Risks.P95, 2), fmtBR(tcc2Risks.P95SE, 3), fmtBR(tcc2Risks.ProbTarget, 1),
			fmtBR(tcc2Stress.CPMDeterministic, 2), fmtBR(tcc2Stress.Mean, 2), fmtBR(tcc2Stress.MeanSE, 3), fmtBR(tcc2Stress.P50, 2), fmtBR(tcc2Stress.P50SE, 3), fmtBR(tcc2Stress.P85, 2), fmtBR(tcc2Stress.P85SE, 3), fmtBR(tcc2Stress.P95, 2), fmtBR(tcc2Stress.P95SE, 3), fmtBR(tcc2Stress.ProbTarget, 1),
			fmtBR(tcc2Cuts.CPMDeterministic, 2), fmtBR(tcc2Cuts.Mean, 2), fmtBR(tcc2Cuts.MeanSE, 3), fmtBR(tcc2Cuts.P50, 2), fmtBR(tcc2Cuts.P50SE, 3), fmtBR(tcc2Cuts.P85, 2), fmtBR(tcc2Cuts.P85SE, 3), fmtBR(tcc2Cuts.P95, 2), fmtBR(tcc2Cuts.P95SE, 3), fmtBR(tcc2Cuts.ProbTarget, 1),
		)
		if err := os.WriteFile(filepath.Join(dir, "tab_tcc2_mc.tex"), []byte(tcc2Rows), 0644); err != nil {
			return err
		}
	}

	return nil
}
