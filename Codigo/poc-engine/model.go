package main

import (
	"fmt"
	"math"
)

// Task represents a work unit with 3-point estimates and progress.
type Task struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	To             float64 `json:"to"`              // Optimistic duration
	Tm             float64 `json:"tm"`              // Most likely duration
	Tp             float64 `json:"tp"`              // Pessimistic duration
	Progress       float64 `json:"progress"`        // c in [0, 1]
	ActualDuration float64 `json:"actual_duration"` // Used when Progress == 1
	Completed      bool    `json:"completed"`
}

// Dependency represents a Finish-to-Start (FS) precedence constraint: From -> To.
type Dependency struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Risk represents an event risk with Bernoulli occurrence and delay impact.
type Risk struct {
	ID          string  `json:"id"`
	TaskID      string  `json:"task_id"`     // Required: ID of the target task in ProjectNetwork.Tasks
	Probability float64 `json:"probability"` // p_r in [0, 1]
	ImpactDays  float64 `json:"impact_days"` // Delta_r in business days (>= 0)
	IsAbsolute  bool    `json:"is_absolute"` // If true, s_{r,i} = 1 regardless of progress
	Active      bool    `json:"active"`      // If false, risk is mitigated/inactive
}

// ProjectNetwork contains the entire DAG topology and risk registry.
type ProjectNetwork struct {
	Tasks        []Task       `json:"tasks"`
	Dependencies []Dependency `json:"dependencies"`
	Risks        []Risk       `json:"risks"`
	StatusDate   float64      `json:"status_date"` // t_s (in business days from project start, >= 0)
}

// CPMTaskResult stores the deterministic schedule metrics for a single task.
type CPMTaskResult struct {
	ES       float64 `json:"es"`
	EF       float64 `json:"ef"`
	LS       float64 `json:"ls"`
	LF       float64 `json:"lf"`
	TF       float64 `json:"tf"`
	Critical bool    `json:"critical"`
}

// CPMResult contains the deterministic CPM forward/backward pass results.
type CPMResult struct {
	Duration   float64                  `json:"duration"`    // Remaining duration T_CPM from status date
	FinishDate float64                  `json:"finish_date"` // F = t_s + T_CPM
	Tasks      map[string]CPMTaskResult `json:"tasks"`
}

// MonteCarloConfig holds configuration for the simulation.
type MonteCarloConfig struct {
	Iterations     int     `json:"iterations"`
	Seed           uint64  `json:"seed"`
	NumWorkers     int     `json:"num_workers"`
	TargetDate     float64 `json:"target_date"`     // Threshold for P(F <= target)
	Lambda         float64 `json:"lambda"`          // PERT-Beta shape parameter (default 4.0, must be > 0)
	EpsilonFloat   float64 `json:"epsilon"`         // Numerical tolerance for float comparison (default 1e-9)
	CorrelationRho float64 `json:"correlation_rho"` // Latent Gaussian copula correlation in [0, 1)
}

// MonteCarloResult holds the aggregated statistical distribution of project outcomes.
type MonteCarloResult struct {
	CPMDeterministic float64            `json:"cpm_deterministic"`
	Mean             float64            `json:"mean"`
	MeanSE           float64            `json:"mean_se"`
	P50              float64            `json:"p50"`
	P50SE            float64            `json:"p50_se"`
	P85              float64            `json:"p85"`
	P85SE            float64            `json:"p85_se"`
	P95              float64            `json:"p95"`
	P95SE            float64            `json:"p95_se"`
	ProbTarget       float64            `json:"prob_target"`
	CriticalityIndex map[string]float64 `json:"criticality_index"`
	Durations        []float64          `json:"-"` // Sorted finish dates
	RawFinishes      []float64          `json:"-"` // Unsorted finish dates in index order k = 0..N-1
	DurationMs       float64            `json:"duration_ms"`
	Iterations       int                `json:"iterations"`
	NumWorkers       int                `json:"num_workers"`
}

// CRNDiffResult stores paired Common Random Numbers (CRN) comparison metrics between two scenarios.
type CRNDiffResult struct {
	DeltaMean        float64 `json:"delta_mean"`
	DeltaMeanSE      float64 `json:"delta_mean_se"`
	IndependentSE    float64 `json:"independent_se"`
	VarianceRedRatio float64 `json:"variance_reduction_ratio"`
	DeltaP50         float64 `json:"delta_p50"`
	DeltaP50SE       float64 `json:"delta_p50_se"`
	DeltaP85         float64 `json:"delta_p85"`
	DeltaP85SE       float64 `json:"delta_p85_se"`
	DeltaP95         float64 `json:"delta_p95"`
	DeltaP95SE       float64 `json:"delta_p95_se"`
	PairedMeanDiff   float64 `json:"paired_mean_diff"`
	PairedP50Diff    float64 `json:"paired_p50_diff"`
	PairedP85Diff    float64 `json:"paired_p85_diff"`
	PairedP95Diff    float64 `json:"paired_p95_diff"`
	PairedP85DiffSE  float64 `json:"paired_p85_diff_se"`
}

func isInvalidFloat(x float64) bool {
	return math.IsNaN(x) || math.IsInf(x, 0)
}

// ValidateTasksAndDeps validates the task list and dependency references.
func ValidateTasksAndDeps(tasks []Task, deps []Dependency) error {
	if len(tasks) == 0 {
		return fmt.Errorf("project network must contain at least one task")
	}
	seen := make(map[string]bool, len(tasks))
	for _, t := range tasks {
		if t.ID == "" {
			return fmt.Errorf("task ID cannot be empty")
		}
		if seen[t.ID] {
			return fmt.Errorf("duplicate task ID: %s", t.ID)
		}
		seen[t.ID] = true

		if isInvalidFloat(t.To) || isInvalidFloat(t.Tm) || isInvalidFloat(t.Tp) || isInvalidFloat(t.Progress) || isInvalidFloat(t.ActualDuration) {
			return fmt.Errorf("task %s contains NaN or Inf values", t.ID)
		}
		if t.To < 0 || t.Tm < t.To || t.Tp < t.Tm {
			return fmt.Errorf("task %s has invalid 3-point estimates (require 0 <= To <= Tm <= Tp): got To=%v, Tm=%v, Tp=%v", t.ID, t.To, t.Tm, t.Tp)
		}
		if t.Progress < 0.0 || t.Progress > 1.0 {
			return fmt.Errorf("task %s has progress outside [0, 1]: %v", t.ID, t.Progress)
		}
		if t.ActualDuration < 0.0 {
			return fmt.Errorf("task %s has negative actual duration: %v", t.ID, t.ActualDuration)
		}
	}
	for _, d := range deps {
		if !seen[d.From] {
			return fmt.Errorf("dependency from unknown task: %s", d.From)
		}
		if !seen[d.To] {
			return fmt.Errorf("dependency to unknown task: %s", d.To)
		}
	}
	return nil
}

// ValidateNetwork validates the entire ProjectNetwork and MonteCarloConfig.
func ValidateNetwork(net ProjectNetwork, cfg MonteCarloConfig) error {
	if err := ValidateTasksAndDeps(net.Tasks, net.Dependencies); err != nil {
		return err
	}
	if isInvalidFloat(net.StatusDate) || net.StatusDate < 0.0 {
		return fmt.Errorf("invalid status_date (must be >= 0): %v", net.StatusDate)
	}
	if isInvalidFloat(cfg.Lambda) || cfg.Lambda < 0.0 {
		return fmt.Errorf("invalid lambda (must be > 0): %v", cfg.Lambda)
	}
	if isInvalidFloat(cfg.EpsilonFloat) || cfg.EpsilonFloat < 0.0 {
		return fmt.Errorf("invalid epsilon (must be > 0): %v", cfg.EpsilonFloat)
	}
	if isInvalidFloat(cfg.CorrelationRho) || cfg.CorrelationRho < 0.0 || cfg.CorrelationRho >= 1.0 {
		return fmt.Errorf("invalid correlation_rho (must be in [0, 1)): %v", cfg.CorrelationRho)
	}

	taskSet := make(map[string]bool, len(net.Tasks))
	for _, t := range net.Tasks {
		taskSet[t.ID] = true
	}

	for _, r := range net.Risks {
		if r.TaskID == "" {
			return fmt.Errorf("risk %s has empty TaskID (every risk must reference a valid TaskID)", r.ID)
		}
		if !taskSet[r.TaskID] {
			return fmt.Errorf("risk %s references unknown TaskID %q", r.ID, r.TaskID)
		}
		if isInvalidFloat(r.Probability) || r.Probability < 0.0 || r.Probability > 1.0 {
			return fmt.Errorf("risk %s has probability outside [0, 1]: %v", r.ID, r.Probability)
		}
		if isInvalidFloat(r.ImpactDays) || r.ImpactDays < 0.0 {
			return fmt.Errorf("risk %s has negative or invalid impact_days: %v", r.ID, r.ImpactDays)
		}
	}
	return nil
}


