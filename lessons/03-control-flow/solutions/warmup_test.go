package main

import (
	"reflect"
	"testing"
)

func TestWarmupClassify(t *testing.T) {
	cases := []struct {
		name string
		n    int
		want string
	}{
		{"positive-small", 7, "positive"},
		{"negative-small", -3, "negative"},
		{"zero", 0, "zero"},
		{"positive-large", 1_000_000, "positive"},
		{"negative-large", -1_000_000, "negative"},
		{"positive-one", 1, "positive"},
		{"negative-one", -1, "negative"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := WarmupClassify(tc.n); got != tc.want {
				t.Errorf("WarmupClassify(%d) = %q, want %q", tc.n, got, tc.want)
			}
		})
	}
}

func TestWarmupFizzBuzz(t *testing.T) {
	cases := []struct {
		name string
		n    int
		want []string
	}{
		{"five", 5, []string{"1", "2", "Fizz", "4", "Buzz"}},
		{"fifteen", 15, []string{
			"1", "2", "Fizz", "4", "Buzz",
			"Fizz", "7", "8", "Fizz", "Buzz",
			"11", "Fizz", "13", "14", "FizzBuzz",
		}},
		{"one", 1, []string{"1"}},
		{"three", 3, []string{"1", "2", "Fizz"}},
		{"zero-empty-slice", 0, []string{}},
		{"negative-empty-slice", -5, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := WarmupFizzBuzz(tc.n)
			if got == nil {
				t.Fatalf("WarmupFizzBuzz(%d) returned a nil slice; expected an empty slice", tc.n)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("WarmupFizzBuzz(%d) = %v, want %v", tc.n, got, tc.want)
			}
		})
	}
}
