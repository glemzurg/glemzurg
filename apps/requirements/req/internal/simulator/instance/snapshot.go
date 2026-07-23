package instance

import (
	"sort"

	"github.com/glemzurg/glemzurg/apps/requirements/req/internal/identity"
)

// Snapshot is a read-only export of the run world for reporting (not live pointers).
type Snapshot struct {
	InstanceCount int
	LinkCount     int
	Instances     []SnapshotInstance
}

// SnapshotInstance is one instance's exported identity and attribute strings.
type SnapshotInstance struct {
	ID         ID
	ClassKey   identity.Key
	Attributes map[string]string
}

// Snapshot builds a deterministic report of instances and link counts.
// Instances are ordered by ID. Attribute values use Object.Inspect().
func (s *State) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := Snapshot{
		InstanceCount: len(s.instances),
		LinkCount:     s.links.Count(),
		Instances:     make([]SnapshotInstance, 0, len(s.instances)),
	}

	for _, inst := range s.instances {
		attrs := make(map[string]string)
		for _, name := range inst.AttributeNames() {
			val := inst.GetAttribute(name)
			if val != nil {
				attrs[name] = val.Inspect()
			}
		}
		out.Instances = append(out.Instances, SnapshotInstance{
			ID:         inst.ID,
			ClassKey:   inst.ClassKey,
			Attributes: attrs,
		})
	}

	sort.Slice(out.Instances, func(i, j int) bool {
		return out.Instances[i].ID < out.Instances[j].ID
	})

	return out
}
