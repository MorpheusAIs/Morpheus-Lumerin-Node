package lib

import "testing"

func TestUnique(t *testing.T) {
	s := NewSet()
	s.Add("a")
	s.Add("b")
	s.Add("c")
	s.Add("a")
	s.Add("b")
	s.Add("c")
	if s.Len() != 3 {
		t.Errorf("Expected 3, got %d", s.Len())
	}
}

func TestRemove(t *testing.T) {
	s := NewSet()
	s.Add("a")
	s.Add("b")
	s.Add("c")
	s.Remove("b")
	if s.Contains("b") {
		t.Errorf("Expected false, got true")
	}
}

func TestContains(t *testing.T) {
	s := NewSet()
	s.Add("a")
	s.Add("b")
	s.Add("c")
	if !s.Contains("b") {
		t.Errorf("Expected true, got false")
	}
}

func TestLen(t *testing.T) {
	s := NewSet()
	s.Add("a")
	s.Add("b")
	s.Add("c")
	if s.Len() != 3 {
		t.Errorf("Expected 3, got %d", s.Len())
	}
}

func TestToSlice(t *testing.T) {
	s := NewSet()
	s.Add("a")
	s.Add("b")
	s.Add("c")
	if len(s.ToSlice()) != 3 {
		t.Errorf("Expected 3, got %d", len(s.ToSlice()))
	}
}
