package main

import (
	"math"
	"testing"
)

// TestPertBetaMomentsMatchAnalytical valida o amostrador (Marsaglia-Tsang + Beta)
// contra as fórmulas fechadas da seção 2.2 do texto:
//
//	mu     = (to + 4*tm + tp) / 6
//	sigma2 = (mu - to) * (tp - mu) / 7
//
// Sem este teste, a PoC só "reproduz a si mesma"; com ele há uma referência
// analítica independente do código do motor.
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

		// erro-padrão da média e tolerância de 5 sigmas (falso alarme ~ 6e-7)
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

// TestDeterministicCPMScenarios confere o CPM determinístico dos cenários A e C
// contra o cálculo manual (soma de médias PERT ao longo de T1->max(T2,T3)->T4).
func TestDeterministicCPMScenarios(t *testing.T) {
	cfg := MonteCarloConfig{Iterations: 100, Seed: 1, NumWorkers: 1, Lambda: 4, EpsilonFloat: 1e-9}

	// A: 19/6 + 5.5 + 13/6 = 10.8333...
	resA, err := RunMonteCarlo(buildBaseNetwork(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(resA.CPMDeterministic-(19.0/6+5.5+13.0/6)) > 1e-9 {
		t.Errorf("Cenário A: CPM = %.6f, esperado %.6f", resA.CPMDeterministic, 19.0/6+5.5+13.0/6)
	}

	// C: ts=5; T1 concluída; T3 restante = 0.7*5.5 = 3.85 (maior que T2: 0.6*5.5 = 3.3); T4 = 13/6
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

// TestPrintReproducibilityValues NÃO substitui TestReproducibilityAcrossWorkers:
// apenas imprime (go test -v -run TestPrintReproducibilityValues) os valores reais
// para preencher a Tabela 12 do texto com saída verdadeira do comando.
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
