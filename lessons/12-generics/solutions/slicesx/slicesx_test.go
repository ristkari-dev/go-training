package slicesx

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestFilterInts(t *testing.T) {
	even := func(x int) bool { return x%2 == 0 }
	gt100 := func(x int) bool { return x > 100 }
	cases := []struct {
		name string
		in   []int
		pred func(int) bool
		want []int
	}{
		{"evens", []int{1, 2, 3, 4, 5, 6}, even, []int{2, 4, 6}},
		{"no-matches", []int{1, 2, 3}, gt100, []int{}},
		{"empty-input", []int{}, even, []int{}},
		{"all-match", []int{2, 4, 6}, even, []int{2, 4, 6}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Filter(tc.in, tc.pred)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Filter(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestFilterStrings(t *testing.T) {
	longerThan1 := func(s string) bool { return len(s) > 1 }
	got := Filter([]string{"a", "bb", "ccc"}, longerThan1)
	want := []string{"bb", "ccc"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Filter strings = %v, want %v", got, want)
	}
}

func TestMapIntToString(t *testing.T) {
	got := Map([]int{1, 2, 3}, strconv.Itoa)
	want := []string{"1", "2", "3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map ints→strings = %v, want %v", got, want)
	}
}

func TestMapStringToString(t *testing.T) {
	got := Map([]string{"hi", "world"}, strings.ToUpper)
	want := []string{"HI", "WORLD"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map ToUpper = %v, want %v", got, want)
	}
}

func TestMapStringToInt(t *testing.T) {
	got := Map([]string{"hi", "world", ""}, func(s string) int { return len(s) })
	want := []int{2, 5, 0}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map len = %v, want %v", got, want)
	}
}

func TestMapEmpty(t *testing.T) {
	got := Map([]int{}, strconv.Itoa)
	if len(got) != 0 {
		t.Errorf("Map([]) = %v, want empty", got)
	}
}
