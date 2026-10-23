package main

import (
	"fmt"
	"math"
	"sort"
)

// TopologicalSort sorts task IDs topologically using Kahn's algorithm.
// Returns an error if the graph contains a cycle.
func TopologicalSort(tasks []Task, deps []Dependency) ([]string, error) {
	inDegree := make(map[string]int)
	adj := make(map[string][]string)

	for _, t := range tasks {
		inDegree[t.ID] = 0
		adj[t.ID] = make([]string, 0)
	}

	for _, d := range deps {
		if _, ok := inDegree[d.From]; !ok {
			return nil, fmt.Errorf("dependency from unknown task: %s", d.From)
		}
		if _, ok := inDegree[d.To]; !ok {
			return nil, fmt.Errorf("dependency to unknown task: %s", d.To)
		}
		adj[d.From] = append(adj[d.From], d.To)
		inDegree[d.To]++
	}

	queue := make([]string, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	sorted := make([]string, 0, len(tasks))
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		sorted = append(sorted, u)

		for _, v := range adj[u] {
			inDegree[v]--
			if inDegree[v] == 0 {
				queue = append(queue, v)
				sort.Strings(queue)
			}
		}
	}

	if len(sorted) != len(tasks) {
		return nil, fmt.Errorf("cycle detected in project dependencies graph")
	}

	return sorted, nil
}

// CalculateExpectedDuration returns (to + 4*tm + tp) / 6
func CalculateExpectedDuration(to, tm, tp float64) float64 {
	return (to + 4.0*tm + tp) / 6.0
}

// CompiledGraph represents the static topology indexed by integers for allocation-free CPM.
type CompiledGraph struct {
	NumTasks int
	Order    []int   // Topological order of task indices
	Preds    [][]int // Predecessor indices for each task
	Succs    [][]int // Active successor indices for each task
	IsActive []bool  // True if task is not completed (progress < 1.0)
	TaskMap  map[string]int
	TaskIDs  []string
}

// CompileGraph compiles the DAG into integer slices for ultra-fast simulation.
func CompileGraph(tasks []Task, deps []Dependency) (*CompiledGraph, error) {
	orderStr, err := TopologicalSort(tasks, deps)
	if err != nil {
		return nil, err
	}

	taskMap := make(map[string]int)
	taskIDs := make([]string, len(tasks))
	for i, t := range tasks {
		taskMap[t.ID] = i
		taskIDs[i] = t.ID
	}

	order := make([]int, len(orderStr))
	for i, id := range orderStr {
		order[i] = taskMap[id]
	}

	preds := make([][]int, len(tasks))
	succs := make([][]int, len(tasks))
	for i := range tasks {
		preds[i] = make([]int, 0)
		succs[i] = make([]int, 0)
	}

	for _, d := range deps {
		u := taskMap[d.From]
		v := taskMap[d.To]
		succs[u] = append(succs[u], v)
		preds[v] = append(preds[v], u)
	}

	isActive := make([]bool, len(tasks))
	for i, t := range tasks {
		isActive[i] = !t.Completed && t.Progress < 1.0
	}

	// Filter successors to only active successors
	activeSuccs := make([][]int, len(tasks))
	for i := range tasks {
		activeSuccs[i] = make([]int, 0, len(succs[i]))
		for _, v := range succs[i] {
			if isActive[v] {
				activeSuccs[i] = append(activeSuccs[i], v)
			}
		}
	}

	return &CompiledGraph{
		NumTasks: len(tasks),
		Order:    order,
		Preds:    preds,
		Succs:    activeSuccs,
		IsActive: isActive,
		TaskMap:  taskMap,
		TaskIDs:  taskIDs,
	}, nil
}

// FastCPM executes CPM using caller-provided buffers (es, ef, ls, lf, critical),
// so a call performs no heap allocation.
func FastCPM(g *CompiledGraph, durations []float64, es, ef, ls, lf []float64, critical []bool, eps float64) float64 {
	// 1. Forward pass
	for _, u := range g.Order {
		if !g.IsActive[u] {
			es[u] = 0
			ef[u] = 0
			continue
		}

		maxEF := 0.0
		for _, p := range g.Preds[u] {
			if ef[p] > maxEF {
				maxEF = ef[p]
			}
		}
		es[u] = maxEF
		ef[u] = maxEF + durations[u]
	}

	// 2. Project duration
	tCPM := 0.0
	for _, u := range g.Order {
		if g.IsActive[u] && ef[u] > tCPM {
			tCPM = ef[u]
		}
	}

	// 3. Backward pass
	for i := len(g.Order) - 1; i >= 0; i-- {
		u := g.Order[i]
		if !g.IsActive[u] {
			ls[u] = 0
			lf[u] = 0
			continue
		}

		if len(g.Succs[u]) == 0 {
			lf[u] = tCPM
		} else {
			minLS := math.MaxFloat64
			for _, s := range g.Succs[u] {
				if ls[s] < minLS {
					minLS = ls[s]
				}
			}
			lf[u] = minLS
		}
		ls[u] = lf[u] - durations[u]
	}

	// 4. Critical flags (buffer fornecido pelo chamador; reiniciado a cada chamada)
	for _, u := range g.Order {
		critical[u] = false
		if g.IsActive[u] {
			tf := lf[u] - ef[u]
			if math.Abs(tf) < eps {
				critical[u] = true
			}
		}
	}

	return tCPM
}

// SolveCPM provides the high-level API with map output for reports and validation.
func SolveCPM(tasks []Task, deps []Dependency, durations map[string]float64, statusDate float64, eps float64) (*CPMResult, error) {
	g, err := CompileGraph(tasks, deps)
	if err != nil {
		return nil, err
	}

	durSlice := make([]float64, g.NumTasks)
	for id, d := range durations {
		durSlice[g.TaskMap[id]] = d
	}

	es := make([]float64, g.NumTasks)
	ef := make([]float64, g.NumTasks)
	ls := make([]float64, g.NumTasks)
	lf := make([]float64, g.NumTasks)

	crit := make([]bool, g.NumTasks)
	tCPM := FastCPM(g, durSlice, es, ef, ls, lf, crit, eps)

	taskResults := make(map[string]CPMTaskResult)
	for i, id := range g.TaskIDs {
		tf := lf[i] - ef[i]
		taskResults[id] = CPMTaskResult{
			ES:       es[i],
			EF:       ef[i],
			LS:       ls[i],
			LF:       lf[i],
			TF:       tf,
			Critical: crit[i],
		}
	}

	return &CPMResult{
		Duration:   tCPM,
		FinishDate: statusDate + tCPM,
		Tasks:      taskResults,
	}, nil
}
