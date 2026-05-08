package exercises

import "testing"

func TestWarmupZeroValues(t *testing.T) {
	i, f, s, b := WarmupZeroValues()
	if i != 0 {
		t.Errorf("int zero value: got %d, want 0", i)
	}
	if f != 0.0 {
		t.Errorf("float64 zero value: got %g, want 0", f)
	}
	if s != "" {
		t.Errorf("string zero value: got %q, want %q", s, "")
	}
	if b != false {
		t.Errorf("bool zero value: got %v, want false", b)
	}
}

func TestWarmupConvert(t *testing.T) {
	cases := []struct {
		name      string
		intVal    int
		floatVal  float64
		wantFloat float64
		wantInt   int
	}{
		{"positive", 5, 3.7, 5.0, 3},
		{"zero", 0, 0.0, 0.0, 0},
		{"negative-truncates-toward-zero", 0, -2.9, 0.0, -2},
		{"large", 1000, 999.99, 1000.0, 999},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotFloat, gotInt := WarmupConvert(tc.intVal, tc.floatVal)
			if gotFloat != tc.wantFloat {
				t.Errorf("WarmupConvert(%d, %g) float: got %g, want %g",
					tc.intVal, tc.floatVal, gotFloat, tc.wantFloat)
			}
			if gotInt != tc.wantInt {
				t.Errorf("WarmupConvert(%d, %g) int: got %d, want %d",
					tc.intVal, tc.floatVal, gotInt, tc.wantInt)
			}
		})
	}
}
