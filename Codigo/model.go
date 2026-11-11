package main

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
	TaskID      string  `json:"task_id"`     // If empty, applies to synthetic project end node Z
	Probability float64 `json:"probability"` // p_r in [0, 1]
	ImpactDays  float64 `json:"impact_days"` // Delta_r in business days
	IsAbsolute  bool    `json:"is_absolute"` // If true, s_{r,i} = 1 regardless of progress
	Active      bool    `json:"active"`      // If false, risk is mitigated/inactive
}

// ProjectNetwork contains the entire DAG topology and risk registry.
type ProjectNetwork struct {
	Tasks        []Task       `json:"tasks"`
	Dependencies []Dependency `json:"dependencies"`
	Risks        []Risk       `json:"risks"`
	StatusDate   float64      `json:"status_date"` // t_s (in business days from project start)
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
	Iterations   int     `json:"iterations"`
	Seed         uint64  `json:"seed"`
	NumWorkers   int     `json:"num_workers"`
	TargetDate   float64 `json:"target_date"` // Threshold for P(F <= target)
	Lambda       float64 `json:"lambda"`      // PERT-Beta shape parameter (default 4.0)
	EpsilonFloat float64 `json:"epsilon"`     // Numerical tolerance for float comparison (default 1e-9)
}

// MonteCarloResult holds the aggregated statistical distribution of project outcomes.
type MonteCarloResult struct {
	CPMDeterministic float64            `json:"cpm_deterministic"`
	P50              float64            `json:"p50"`
	P85              float64            `json:"p85"`
	P95              float64            `json:"p95"`
	ProbTarget       float64            `json:"prob_target"`
	CriticalityIndex map[string]float64 `json:"criticality_index"`
	Durations        []float64          `json:"-"`
	DurationMs       float64            `json:"duration_ms"`
	Iterations       int                `json:"iterations"`
	NumWorkers       int                `json:"num_workers"`
}

