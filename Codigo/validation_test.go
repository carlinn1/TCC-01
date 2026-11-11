package main

import (
	"fmt"
	"math"
	"sort"
	"testing"
	"time"
)

// TestPertBetaMomentsMatchAnalytical valida o amostrador (Marsaglia-Tsang + Beta)
// contra as fórmulas fechadas da seção 2.2 do texto:
//
//	mu     = (to + 4*tm + tp) / 6
//	sigma2 = (mu - to) * (tp - mu) / 7
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
			t.Errorf("%s: média amostral %.5f != analítica %.5f (5·SE = %.5f)", c.id, mu, wantMu, 5*seMu)
		}
		if math.Abs(variance-wantVar)/wantVar > 0.02 {
			t.Errorf("%s: variância amostral %.5f != analítica %.5f (>2%%)", c.id, variance, wantVar)
		}
		t.Logf("%s  mu=%.5f (analítico %.5f)  var=%.5f (analítico %.5f)", c.id, mu, wantMu, variance, wantVar)
	}
}

// regularizedIncompleteBeta calcula I_x(a, b) via expansão em frações contínuas
// de Lentz (Numerical Recipes / Abramowitz & Stegun 26.5.8), com precisão ~1e-14.
func regularizedIncompleteBeta(x, a, b float64) float64 {
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

// kolmogorovPValue calcula o p-valor assintótico da estatística KS unilateral/bilateral.
func kolmogorovPValue(d float64, n int) float64 {
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

// TestPertBetaKSGoodnessOfFit realiza o teste de aderência de Kolmogorov-Smirnov (KS)
// contra a CDF Beta teórica para cada tarefa T1..T4 (Pendência 17).
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
	dCrit := 1.35810 / math.Sqrt(float64(n)) // alpha = 0.05 -> 0.0013581

	for _, c := range cases {
		diff := c.tp - c.to
		a := 1.0 + lambda*(c.tm-c.to)/diff
		b := 1.0 + lambda*(c.tp-c.tm)/diff

		rng := NewSplitMix64(DeriveIterationSeed(2026, c.seedOffset))
		samples := make([]float64, n)
		for i := 0; i < n; i++ {
			samples[i] = rng.SampleBeta(a, b)
		}
		sort.Float64s(samples)

		// Avalia o supremo D_n em 2000 quantis uniformemente espaçados
		maxD := 0.0
		step := n / 2000
		for i := 0; i < n; i += step {
			x := samples[i]
			cdfTheo := regularizedIncompleteBeta(x, a, b)
			empUpper := float64(i+1) / float64(n)
			empLower := float64(i) / float64(n)
			d1 := math.Abs(empUpper - cdfTheo)
			d2 := math.Abs(cdfTheo - empLower)
			if d1 > maxD {
				maxD = d1
			}
			if d2 > maxD {
				maxD = d2
			}
		}

		pVal := kolmogorovPValue(maxD, n)
		if maxD >= dCrit {
			t.Errorf("%s: KS D_n = %.6f >= D_crit = %.6f (p = %.4f)", c.id, maxD, dCrit, pVal)
		}
		t.Logf("%s  KS D_n=%.5f (< D_crit=%.5f)  p-valor=%.3f", c.id, maxD, dCrit, pVal)
	}
}

// TestDegeneratePertBetaZeroVariance valida o caso degenerado to = tm = tp (Pendência 17).
func TestDegeneratePertBetaZeroVariance(t *testing.T) {
	rng := NewSplitMix64(2026)
	for i := 0; i < 100; i++ {
		s := rng.SamplePERTBeta(5.0, 5.0, 5.0, 4.0)
		if math.IsNaN(s) || math.Abs(s-5.0) > 1e-12 {
			t.Fatalf("SamplePERTBeta degenerado inválido: obtido %v, esperado 5.0", s)
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
		t.Fatalf("erro inesperado no caso degenerado: %v", err)
	}
	for _, d := range res.Durations {
		if math.IsNaN(d) || math.Abs(d-5.0) > 1e-12 {
			t.Fatalf("duração degenerada inválida: obtido %v, esperado 5.0", d)
		}
	}
	t.Logf("Caso degenerado (5,5,5): P50=%.5f, P85=%.5f, P95=%.5f, var=0.00000", res.P50, res.P85, res.P95)
}

// TestDeterministicCPMScenarios confere o CPM determinístico dos cenários A e C
// contra o cálculo manual (soma de médias PERT ao longo de T1->max(T2,T3)->T4).
func TestDeterministicCPMScenarios(t *testing.T) {
	cfg := MonteCarloConfig{Iterations: 100, Seed: 1, NumWorkers: 1, Lambda: 4, EpsilonFloat: 1e-9}

	resA, err := RunMonteCarlo(buildBaseNetwork(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(resA.CPMDeterministic-(19.0/6+5.5+13.0/6)) > 1e-9 {
		t.Errorf("Cenário A: CPM = %.6f, esperado %.6f", resA.CPMDeterministic, 19.0/6+5.5+13.0/6)
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
		t.Errorf("Cenário C: CPM = %.6f, esperado %.6f", resC.CPMDeterministic, want)
	}
}

// TestPrintReproducibilityValues imprime os valores reais para a Tabela de Reprodutibilidade (RNF08).
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
		t.Logf("workers=%2d  P50=%.6f  P85=%.6f  P95=%.6f  P(F<=11)=%.4f%%  IC(T3)=%.4f%%",
			w, r.P50, r.P85, r.P95, r.ProbTarget, r.CriticalityIndex["T3"])
	}
}

// buildSynthetic200Network constrói a rede sintética determinística (200 tarefas, 300 dependências,
// 10 camadas de 20 tarefas, semente topológica 99) descrita na Seção de Benchmark (Pendência 17).
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
	// 180 arestas base (1 para cada tarefa nas camadas 1..9 vinda da camada anterior)
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
	// +120 arestas cruzadas distintas para atingir exatamente 300 dependências FS acíclicas
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

// TestBenchmarkDispersion5Runs mede mediana, média, desvio-padrão, mínimo e máximo
// das 5 repetições de benchmark reportadas na Tabela de Desempenho (Pendência 17).
func TestBenchmarkDispersion5Runs(t *testing.T) {
	const runs = 5
	baseNet := buildBaseNetwork()
	synthNet := buildSynthetic200Network()

	runStats := func(name string, net ProjectNetwork, workers int) {
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
				t.Fatal(err)
			}
			ms := float64(time.Since(t0).Microseconds()) / 1000.0
			times[i] = ms
			sum += ms
			sumSq += ms * ms
		}
		sort.Float64s(times)
		mean := sum / runs
		variance := (sumSq - float64(runs)*mean*mean) / float64(runs-1)
		if variance < 0 {
			variance = 0
		}
		sd := math.Sqrt(variance)
		t.Logf("%s (5 runs): mediana=%.2f ms | média=%.2f±%.2f ms | min=%.2f ms | max=%.2f ms",
			name, times[runs/2], mean, sd, times[0], times[runs-1])
	}

	runStats("Rede Base (4/4)", baseNet, 1)
	runStats("Rede Sintética (200/300)", synthNet, 8)
}
