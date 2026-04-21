package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiff_AllAdded(t *testing.T) {
	incoming := map[string]string{"FOO": "bar", "BAZ": "qux"}
	changes := Diff(nil, incoming)

	assert.Len(t, changes, 2)
	for _, c := range changes {
		assert.Equal(t, "added", c.Action)
	}
}

func TestDiff_UpdatedKey(t *testing.T) {
	prev := map[string]string{"FOO": "old"}
	incoming := map[string]string{"FOO": "new"}
	changes := Diff(prev, incoming)

	assert.Len(t, changes, 1)
	assert.Equal(t, "FOO", changes[0].Key)
	assert.Equal(t, "updated", changes[0].Action)
}

func TestDiff_UnchangedKey(t *testing.T) {
	prev := map[string]string{"FOO": "same"}
	incoming := map[string]string{"FOO": "same"}
	changes := Diff(prev, incoming)

	assert.Len(t, changes, 1)
	assert.Equal(t, "unchanged", changes[0].Action)
}

func TestDiff_MixedActions(t *testing.T) {
	prev := map[string]string{"KEEP": "v", "CHANGE": "old"}
	incoming := map[string]string{"KEEP": "v", "CHANGE": "new", "NEW": "val"}
	changes := Diff(prev, incoming)

	summary := SummarizeDiff(changes)
	assert.Equal(t, 1, summary.Added)
	assert.Equal(t, 1, summary.Updated)
	assert.Equal(t, 1, summary.Unchanged)
}

func TestDiff_SortedByKey(t *testing.T) {
	incoming := map[string]string{"Z": "1", "A": "2", "M": "3"}
	changes := Diff(nil, incoming)

	assert.Equal(t, "A", changes[0].Key)
	assert.Equal(t, "M", changes[1].Key)
	assert.Equal(t, "Z", changes[2].Key)
}

func TestSummarizeDiff_Empty(t *testing.T) {
	s := SummarizeDiff(nil)
	assert.Equal(t, DiffSummary{}, s)
}
