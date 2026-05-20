package counter

import "testing"

func TestCounter(t *testing.T) {
	cases := []struct {
		name     string
		incCount int
		want     int
	}{
		{"zero", 0, 0},
		{"one", 1, 1},
		{"three", 3, 3},
		{"hundred", 100, 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c Counter
			for i := 0; i < tc.incCount; i++ {
				c.Inc()
			}
			if got := c.Value(); got != tc.want {
				t.Errorf("after %d Inc() calls, Value() = %d, want %d", tc.incCount, got, tc.want)
			}
		})
	}
}

func TestCountersDoNotInterfere(t *testing.T) {
	var a, b Counter
	a.Inc()
	a.Inc()
	a.Inc()
	b.Inc()
	if got := a.Value(); got != 3 {
		t.Errorf("a.Value() = %d, want 3", got)
	}
	if got := b.Value(); got != 1 {
		t.Errorf("b.Value() = %d, want 1", got)
	}
}
