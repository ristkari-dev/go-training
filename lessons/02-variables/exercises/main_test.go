package exercises

import "testing"

func TestTipRate(t *testing.T) {
	if TipRate != 0.14 {
		t.Errorf("TipRate = %g, want 0.14", TipRate)
	}
}

func TestTotal(t *testing.T) {
	cases := []struct {
		name    string
		a, b, c float64
		want    float64
	}{
		{"three-expenses", 4.50, 12.00, 23.50, 40.0},
		{"zero", 0, 0, 0, 0},
		{"single-large", 1000.0, 0, 0, 1000.0},
		{"negative-values", -5.0, -3.0, -2.0, -10.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Total(tc.a, tc.b, tc.c)
			if got != tc.want {
				t.Errorf("Total(%g, %g, %g) = %g, want %g", tc.a, tc.b, tc.c, got, tc.want)
			}
		})
	}
}

func TestAverage(t *testing.T) {
	cases := []struct {
		name    string
		a, b, c float64
		want    float64
	}{
		{"clean-divide", 10.0, 20.0, 30.0, 20.0},
		{"zero", 0, 0, 0, 0},
		{"three-expenses-avg", 4.50, 12.00, 23.50, 40.0 / 3.0},
		{"negative-avg", -3.0, -6.0, -9.0, -6.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Average(tc.a, tc.b, tc.c)
			if got != tc.want {
				t.Errorf("Average(%g, %g, %g) = %g, want %g", tc.a, tc.b, tc.c, got, tc.want)
			}
		})
	}
}

func TestSummary(t *testing.T) {
	got := Summary(4.50, 12.00, 23.50)
	want := "3 expenses\nTotal: €40.00\nAverage: €13.33\nTip (14%): €5.60\nTotal with tip: €45.60"
	if got != want {
		t.Errorf("Summary(4.50, 12.00, 23.50) =\n%q\nwant\n%q", got, want)
	}
}
