package logstats

import "testing"

// TestStore is a SKELETON. Merge some deltas, Snapshot, assert counts +
// total. Then exercise concurrent Merge under `go test -race`.
func TestStore(t *testing.T) {
	cases := []struct {
		name   string
		deltas []map[string]int
		want   map[string]int
	}{
		// TODO: e.g. {"single", []map[string]int{{"INFO":2}}, map[string]int{"INFO":2}}
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStore()
			for _, d := range tc.deltas {
				s.Merge(d)
			}
			got, _ := s.Snapshot()
			_ = got
			_ = tc.want
		})
	}
}
