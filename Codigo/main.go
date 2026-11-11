package main

import (
	"encoding/json"
	"fmt"
	"os"
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

func main() {
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

	// Cenário A: Início do projeto, sem riscos
	netA := buildBaseNetwork()
	resA, err := RunMonteCarlo(netA, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro no Cenário A: %v\n", err)
		return
	}

	// Cenário B: Início do projeto, com risco R1 em T3 (pr=0.30, Delta=2)
	netB := buildBaseNetwork()
	netB.Risks = []Risk{
		{ID: "R1", TaskID: "T3", Probability: 0.30, ImpactDays: 2.0, IsAbsolute: false, Active: true},
	}
	resB, err := RunMonteCarlo(netB, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro no Cenário B: %v\n", err)
		return
	}

	// Cenário C: Status date ts=5, T1 concluída (3d), T2 c=0.40, T3 c=0.30, R1 ativo
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
		fmt.Fprintf(os.Stderr, "Erro no Cenário C: %v\n", err)
		return
	}

	// Cenário C': Como C, com risco R1 mitigado (Active = false)
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
		fmt.Fprintf(os.Stderr, "Erro no Cenário C': %v\n", err)
		return
	}

	fmt.Println("### Tabela 9 Reproduzida pela PoC em Go (N = 10.000 iterações, seed = 42)")
	fmt.Println()
	fmt.Println("| Cenário | CPM determinístico (F) | P50 | P85 | P95 | P(F <= 11) | Tempo (ms) |")
	fmt.Println("| :--- | :---: | :---: | :---: | :---: | :---: | :---: |")
	fmt.Printf("| **A.** Início do projeto, sem riscos | %.2f | %.2f | %.2f | %.2f | %.1f%% | %.1f ms |\n",
		resA.CPMDeterministic, resA.P50, resA.P85, resA.P95, resA.ProbTarget, resA.DurationMs)
	fmt.Printf("| **B.** Início, com risco R1 em T3 ($p_r = 0{,}30$, $\\Delta_r = 2$) | %.2f | %.2f | %.2f | %.2f | %.1f%% | %.1f ms |\n",
		resB.CPMDeterministic, resB.P50, resB.P85, resB.P95, resB.ProbTarget, resB.DurationMs)
	fmt.Printf("| **C.** Data de status $t_s = 5$, com R1 ativo | %.2f | %.2f | %.2f | %.2f | %.1f%% | %.1f ms |\n",
		resC.CPMDeterministic, resC.P50, resC.P85, resC.P95, resC.ProbTarget, resC.DurationMs)
	fmt.Printf("| **C′.** Como C, com R1 mitigado | %.2f | %.2f | %.2f | %.2f | %.1f%% | %.1f ms |\n",
		resCPrime.CPMDeterministic, resCPrime.P50, resCPrime.P85, resCPrime.P95, resCPrime.ProbTarget, resCPrime.DurationMs)
	fmt.Println()

	fmt.Println("### Índice de Criticidade por Tarefa (IC %)")
	fmt.Println("| Cenário | T1 | T2 | T3 | T4 |")
	fmt.Println("| :--- | :---: | :---: | :---: | :---: |")
	fmt.Printf("| **A.** Início, sem riscos | %.1f%% | %.1f%% | %.1f%% | %.1f%% |\n",
		resA.CriticalityIndex["T1"], resA.CriticalityIndex["T2"], resA.CriticalityIndex["T3"], resA.CriticalityIndex["T4"])
	fmt.Printf("| **B.** Início, com R1 em T3 | %.1f%% | %.1f%% | %.1f%% | %.1f%% |\n",
		resB.CriticalityIndex["T1"], resB.CriticalityIndex["T2"], resB.CriticalityIndex["T3"], resB.CriticalityIndex["T4"])
	fmt.Printf("| **C.** Status $t_s = 5$, R1 ativo | — | %.1f%% | %.1f%% | %.1f%% |\n",
		resC.CriticalityIndex["T2"], resC.CriticalityIndex["T3"], resC.CriticalityIndex["T4"])
	fmt.Printf("| **C′.** Status $t_s = 5$, R1 mitigado | — | %.1f%% | %.1f%% | %.1f%% |\n",
		resCPrime.CriticalityIndex["T2"], resCPrime.CriticalityIndex["T3"], resCPrime.CriticalityIndex["T4"])
	fmt.Println()

	outputData := map[string]interface{}{
		"scenario_a":       resA,
		"scenario_b":       resB,
		"scenario_c":       resC,
		"scenario_c_prime": resCPrime,
	}
	jsonData, _ := json.MarshalIndent(outputData, "", "  ")
	_ = os.WriteFile("poc_results.json", jsonData, 0644)
	fmt.Println("Resultados exportados para poc_results.json com sucesso.")
}

