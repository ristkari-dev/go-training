package fanin

import (
	"sort"
	"testing"
)

// TestFanInTwo is a SKELETON. Two inputs; send N values on each;
// assert all 2N values appear on the output (in any order).
func TestFanInTwo(t *testing.T) {
	// TODO:
	//   a := make(chan int, 3); b := make(chan int, 3)
	//   a <- 1; a <- 2; a <- 3; close(a)
	//   b <- 4; b <- 5; b <- 6; close(b)
	//   out := FanIn[int](a, b)
	//   var got []int
	//   for v := range out { got = append(got, v) }
	//   sort.Ints(got)
	//   want := []int{1, 2, 3, 4, 5, 6}
	//   compare
	_ = sort.Ints
}

// TestFanInEmpty is a SKELETON. Zero inputs; out should close immediately.
func TestFanInEmpty(t *testing.T) {
	// TODO:
	//   out := FanIn[int]()
	//   for range out { t.Errorf("expected no values") }
}

// TestFanInClosesWithAllInputs is a SKELETON. After closing all inputs,
// out should close. Use a timeout via time.After to detect hangs.
func TestFanInClosesWithAllInputs(t *testing.T) {
	// TODO:
	//   a := make(chan int); b := make(chan int)
	//   out := FanIn[int](a, b)
	//   close(a); close(b)
	//   select {
	//   case _, ok := <-out:
	//       if ok { t.Errorf("expected closed channel") }
	//   case <-time.After(time.Second):
	//       t.Fatal("output didn't close")
	//   }
}
