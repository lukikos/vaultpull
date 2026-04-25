package audit

import (
	"fmt"
	"sort"
	"time"
)

// SnapshotComparison holds the result of comparing two named snapshots.
type SnapshotComparison struct {
	From      string
	To        string
	ComparedAt time.Time
	Diffs     []DiffEntry
}

// Summary returns a human-readable summary of the comparison.
func (c SnapshotComparison) Summary() string {
	s := SummarizeDiff(c.Diffs)
	return fmt.Sprintf("snapshot %q → %q: %d added, %d updated, %d removed, %d unchanged",
		c.From, c.To, s.Added, s.Updated, s.Removed, s.Unchanged)
}

// CompareSnapshots loads two snapshots by name and returns a SnapshotComparison.
func CompareSnapshots(dir, fromName, toName string) (SnapshotComparison, error) {
	fromSnap, err := LoadSnapshot(dir, fromName)
	if err != nil {
		return SnapshotComparison{}, fmt.Errorf("load snapshot %q: %w", fromName, err)
	}

	toSnap, err := LoadSnapshot(dir, toName)
	if err != nil {
		return SnapshotComparison{}, fmt.Errorf("load snapshot %q: %w", toName, err)
	}

	diffs := Diff(fromSnap, toSnap)

	// Stable ordering by key.
	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Key < diffs[j].Key
	})

	return SnapshotComparison{
		From:       fromName,
		To:         toName,
		ComparedAt: time.Now().UTC(),
		Diffs:      diffs,
	}, nil
}
