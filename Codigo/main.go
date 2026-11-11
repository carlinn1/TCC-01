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
	fmt.Println("| Cenário | CPM determinístico (F) | Média (MCSE) | P50 (MCSE) | P85 (MCSE) | P95 (MCSE) | P(F <= 11) | Tempo (ms) |")
	fmt.Println("| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: |")
	fmt.Printf("| **A.** Início do projeto, sem riscos | %.2f | %.2f (±%.3f) | %.2f (±%.3f) | %.2f (±%.3f) | %.2f (±%.3f) | %.1f%% | %.1f ms |\n",
		resA.CPMDeterministic, resA.Mean, resA.MeanSE, resA.P50, resA.P50SE, resA.P85, resA.P85SE, resA.P95, resA.P95SE, resA.ProbTarget, resA.DurationMs)
	fmt.Printf("| **B.** Início, com risco R1 em T3 ($p_r = 0{,}30$, $\\Delta_r = 2$) | %.2f | %.2f (±%.3f) | %.2f (±%.3f) | %.2f (±%.3f) | %.2f (±%.3f) | %.1f%% | %.1f ms |\n",
		resB.CPMDeterministic, resB.Mean, resB.MeanSE, resB.P50, resB.P50SE, resB.P85, resB.P85SE, resB.P95, resB.P95SE, resB.ProbTarget, resB.DurationMs)
	fmt.Printf("| **C.** Data de status $t_s = 5$, com R1 ativo | %.2f | %.2f (±%.3f) | %.2f (±%.3f) | %.2f (±%.3f) | %.2f (±%.3f) | %.1f%% | %.1f ms |\n",
		resC.CPMDeterministic, resC.Mean, resC.MeanSE, resC.P50, resC.P50SE, resC.P85, resC.P85SE, resC.P95, resC.P95SE, resC.ProbTarget, resC.DurationMs)
	fmt.Printf("| **C′.** Como C, com R1 mitigado | %.2f | %.2f (±%.3f) | %.2f (±%.3f) | %.2f (±%.3f) | %.2f (±%.3f) | %.1f%% | %.1f ms |\n",
		resCPrime.CPMDeterministic, resCPrime.Mean, resCPrime.MeanSE, resCPrime.P50, resCPrime.P50SE, resCPrime.P85, resCPrime.P85SE, resCPrime.P95, resCPrime.P95SE, resCPrime.ProbTarget, resCPrime.DurationMs)
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

	// Simulação do cronograma pré-registrado do TCC 2 (duas trilhas de execução: Autor 1 e Autor 2)
	cfgTCC2 := MonteCarloConfig{
		Iterations:   10000,
		Seed:         42,
		NumWorkers:   8,
		TargetDate:   20.0,
		Lambda:       4.0,
		EpsilonFloat: 1e-9,
	}
	netTCC2Base := buildTCC2ScheduleNetwork(false)
	resTCC2Base, err := RunMonteCarlo(netTCC2Base, cfgTCC2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro no Cronograma TCC2 (sem riscos): %v\n", err)
		return
	}
	netTCC2Risks := buildTCC2ScheduleNetwork(true)
	resTCC2Risks, err := RunMonteCarlo(netTCC2Risks, cfgTCC2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro no Cronograma TCC2 (com riscos): %v\n", err)
		return
	}

	fmt.Println("### Simulação de Monte Carlo do Cronograma do TCC 2 (20 semanas úteis, N = 10.000, seed = 42)")
	fmt.Printf("| **Base PERT-Beta (sem riscos discretos)** | CPM=%.2f | Média=%.2f | P50=%.2f (±%.3f) | P85=%.2f (±%.3f) | P95=%.2f (±%.3f) | P(F<=20)=%.1f%% |\n",
		resTCC2Base.CPMDeterministic, resTCC2Base.Mean, resTCC2Base.P50, resTCC2Base.P50SE, resTCC2Base.P85, resTCC2Base.P85SE, resTCC2Base.P95, resTCC2Base.P95SE, resTCC2Base.ProbTarget)
	fmt.Printf("| **Com Riscos Discretos (R_DAG, R_LEAK, R_LLM)** | CPM=%.2f | Média=%.2f | P50=%.2f (±%.3f) | P85=%.2f (±%.3f) | P95=%.2f (±%.3f) | P(F<=20)=%.1f%% |\n",
		resTCC2Risks.CPMDeterministic, resTCC2Risks.Mean, resTCC2Risks.P50, resTCC2Risks.P50SE, resTCC2Risks.P85, resTCC2Risks.P85SE, resTCC2Risks.P95, resTCC2Risks.P95SE, resTCC2Risks.ProbTarget)
	fmt.Println()

	outputData := map[string]interface{}{
		"scenario_a":        resA,
		"scenario_b":        resB,
		"scenario_c":        resC,
		"scenario_c_prime":  resCPrime,
		"tcc2_schedule_base":  resTCC2Base,
		"tcc2_schedule_risks": resTCC2Risks,
	}
	jsonData, _ := json.MarshalIndent(outputData, "", "  ")
	_ = os.WriteFile("poc_results.json", jsonData, 0644)
	fmt.Println("Resultados exportados para poc_results.json com sucesso.")

	if err := exportLatexMacros("../Docs/resultados_poc.tex", resA, resB, resC, resCPrime, resTCC2Base, resTCC2Risks); err == nil {
		_ = exportLatexMacros("../resultados_poc.tex", resA, resB, resC, resCPrime, resTCC2Base, resTCC2Risks)
		fmt.Println("Macros LaTeX exportadas para ../Docs/resultados_poc.tex com sucesso.")
	}
}

// buildTCC2ScheduleNetwork constrói a rede pré-registrada do cronograma do TCC 2 (E01-E14)
// condicional à alocação estática em 2 trilhas de execução (Autor 1: Trilha A; Autor 2: Trilha B)
// e fechamento conjunto (E12-E14).
func buildTCC2ScheduleNetwork(withRisks bool) ProjectNetwork {
	return ProjectNetwork{
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
			{ID: "E08", Name: "Suíte sintética da Dimensão 2 (10 regimes, M=200) e análise de sensibilidade", To: 1.5, Tm: 2.0, Tp: 3.0},
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
			{From: "E02", To: "E04"}, // E04 depende do schema/banco de E02
			{From: "E04", To: "E08"},
			{From: "E05", To: "E08"}, // E08 usa o motor estocástico de E05
			{From: "E08", To: "E09"},
			{From: "E09", To: "E10"},
			{From: "E07", To: "E10"}, // E10 depende dos contratos de propostas/guardrails de E07

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
}

func exportLatexMacros(path string, resA, resB, resC, resCPrime, tcc2Base, tcc2Risks *MonteCarloResult) error {
	content := fmt.Sprintf(`%% Arquivo gerado automaticamente por Codigo/main.go (N = 10.000, seed = 42)
%% Nao edite manualmente; execute "cd Codigo && go run ." para atualizar.
\providecommand{\pocACPM}{%.2f}
\providecommand{\pocAPcinquenta}{%.2f}
\providecommand{\pocAPoitentaeCinco}{%.2f}
\providecommand{\pocAPnoventaeCinco}{%.2f}
\providecommand{\pocAProbMeta}{%.1f}
\providecommand{\pocBCPM}{%.2f}
\providecommand{\pocBPcinquenta}{%.2f}
\providecommand{\pocBPoitentaeCinco}{%.2f}
\providecommand{\pocBPnoventaeCinco}{%.2f}
\providecommand{\pocBProbMeta}{%.1f}
\providecommand{\pocCCPM}{%.2f}
\providecommand{\pocCPcinquenta}{%.2f}
\providecommand{\pocCPoitentaeCinco}{%.2f}
\providecommand{\pocCPnoventaeCinco}{%.2f}
\providecommand{\pocCProbMeta}{%.1f}
\providecommand{\pocCPrimeCPM}{%.2f}
\providecommand{\pocCPrimePcinquenta}{%.2f}
\providecommand{\pocCPrimePoitentaeCinco}{%.2f}
\providecommand{\pocCPrimePnoventaeCinco}{%.2f}
\providecommand{\pocCPrimeProbMeta}{%.1f}
\providecommand{\tccDoisBaseCPM}{%.2f}
\providecommand{\tccDoisBasePcinquenta}{%.2f}
\providecommand{\tccDoisBasePoitentaeCinco}{%.2f}
\providecommand{\tccDoisBasePnoventaeCinco}{%.2f}
\providecommand{\tccDoisBaseProbVinte}{%.1f}
\providecommand{\tccDoisRiscosPcinquenta}{%.2f}
\providecommand{\tccDoisRiscosPoitentaeCinco}{%.2f}
\providecommand{\tccDoisRiscosPnoventaeCinco}{%.2f}
\providecommand{\tccDoisRiscosProbVinte}{%.1f}
`,
		resA.CPMDeterministic, resA.P50, resA.P85, resA.P95, resA.ProbTarget,
		resB.CPMDeterministic, resB.P50, resB.P85, resB.P95, resB.ProbTarget,
		resC.CPMDeterministic, resC.P50, resC.P85, resC.P95, resC.ProbTarget,
		resCPrime.CPMDeterministic, resCPrime.P50, resCPrime.P85, resCPrime.P95, resCPrime.ProbTarget,
		tcc2Base.CPMDeterministic, tcc2Base.P50, tcc2Base.P85, tcc2Base.P95, tcc2Base.ProbTarget,
		tcc2Risks.P50, tcc2Risks.P85, tcc2Risks.P95, tcc2Risks.ProbTarget,
	)
	return os.WriteFile(path, []byte(content), 0644)
}

