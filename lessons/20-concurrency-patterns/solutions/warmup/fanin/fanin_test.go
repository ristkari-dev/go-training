package fanin

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestFanInTwo(t *testing.T) {
	a := make(chan int, 3)
	b := make(chan int, 3)
	a <- 1
	a <- 2
	a <- 3
	close(a)
	b <- 4
	b <- 5
	b <- 6
	close(b)

	out := FanIn[int](a, b)
	var got []int
	for v := range out {
		got = append(got, v)
	}
	sort.Ints(got)

	want := []int{1, 2, 3, 4, 5, 6}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FanIn = %v, want %v", got, want)
	}
}

func TestFanInThree(t *testing.T) {
	a := make(chan string, 1)
	b := make(chan string, 1)
	c := make(chan string, 1)
	a <- "alpha"
	b <- "beta"
	c <- "gamma"
	close(a)
	close(b)
	close(c)

	out := FanIn[string](a, b, c)
	var got []string
	for v := range out {
		got = append(got, v)
	}
	sort.Strings(got)

	want := []string{"alpha", "beta", "gamma"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FanIn = %v, want %v", got, want)
	}
}

func TestFanInEmpty(t *testing.T) {
	out := FanIn[int]()
	for range out {
		t.Errorf("expected no values from empty FanIn")
	}
}

func TestFanInClosesWithAllInputs(t *testing.T) {
	a := make(chan int)
	b := make(chan int)
	out := FanIn[int](a, b)

	close(a)
	close(b)

	select {
	case _, ok := <-out:
		if ok {
			t.Errorf("expected closed channel after both inputs closed")
		}
	case <-time.After(time.Second):
		t.Fatal("output didn't close within 1s — closer goroutine didn't fire")
	}
}
