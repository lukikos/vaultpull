package audit

import "sort"

// CouplingResult describes how often two secret paths are synced together.
type CouplingResult struct {
	PathA     string  `json:"path_a"`
	PathB     string  `json:"path_b"`
	CoOccurs  int     `json:"co_occurs"`
	Support   float64 `json:"support"`   // fraction of total syncs where both appear
}

// DetectCoupling analyses audit entries and returns pairs of paths that are
// frequently synced together, ordered by co-occurrence count descending.
// minSupport is the minimum fraction (0–1) required to include a pair.
func DetectCoupling(entries []Entry, minSupport float64) []CouplingResult {
	if len(entries) == 0 {
		return nil
	}

	// Group paths by timestamp (same-second syncs are treated as one batch).
	batches := map[string]map[string]struct{}{}
	for _, e := range entries {
		if e.Action != "sync" {
			continue
		}
		key := e.Timestamp.UTC().Format("2006-01-02T15:04:05")
		if batches[key] == nil {
			batches[key] = map[string]struct{}{}
		}
		batches[key][e.Path] = struct{}{}
	}

	totalBatches := len(batches)
	if totalBatches == 0 {
		return nil
	}

	coCount := map[[2]string]int{}
	for _, paths := range batches {
		list := make([]string, 0, len(paths))
		for p := range paths {
			list = append(list, p)
		}
		sort.Strings(list)
		for i := 0; i < len(list); i++ {
			for j := i + 1; j < len(list); j++ {
				coCount[[2]string{list[i], list[j]}]++
			}
		}
	}

	var results []CouplingResult
	for pair, count := range coCount {
		support := float64(count) / float64(totalBatches)
		if support < minSupport {
			continue
		}
		results = append(results, CouplingResult{
			PathA:    pair[0],
			PathB:    pair[1],
			CoOccurs: count,
			Support:  support,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].CoOccurs != results[j].CoOccurs {
			return results[i].CoOccurs > results[j].CoOccurs
		}
		return results[i].PathA < results[j].PathA
	})
	return results
}
