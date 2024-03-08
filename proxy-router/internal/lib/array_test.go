package lib

import (
	"testing"
)

func TestFilter(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	f := func(i int) bool { return i%2 == 0 }
	expected := []int{2, 4}
	actual := Filter(arr, f)

	if len(actual) != len(expected) {
		t.Errorf("expected: %v, actual: %v", expected, actual)
	}

	for i := range actual {
		if actual[i] != expected[i] {
			t.Errorf("expected: %v, actual: %v", expected, actual)
		}
	}
}

func TestFilterValue(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	val := 3
	expected := []int{1, 2, 4, 5}
	actual := FilterValue(arr, val)

	if len(actual) != len(expected) {
		t.Errorf("expected: %v, actual: %v", expected, actual)
	}

	for i := range actual {
		if actual[i] != expected[i] {
			t.Errorf("expected: %v, actual: %v", expected, actual)
		}
	}
}
