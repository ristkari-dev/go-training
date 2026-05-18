package clock

import "testing"

func TestNow(t *testing.T) {
	got := Now()

	if len(got) != 19 {
		t.Fatalf("Now() = %q (len %d), want a 19-char string", got, len(got))
	}
	checks := []struct {
		pos int
		ch  byte
	}{
		{4, '-'},
		{7, '-'},
		{10, ' '},
		{13, ':'},
		{16, ':'},
	}
	for _, c := range checks {
		if got[c.pos] != c.ch {
			t.Errorf("Now()[%d] = %q, want %q (full output: %q)", c.pos, got[c.pos], c.ch, got)
		}
	}
}
