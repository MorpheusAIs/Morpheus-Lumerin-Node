package proxyapi

import "testing"

func TestSessionExpiredByServerTime(t *testing.T) {
	if sessionExpiredByServerTime(100, 99999) {
		t.Fatalf("early")
	}
	if !sessionExpiredByServerTime(100, 100001) {
		t.Fatalf("late")
	}
	if sessionExpiredByServerTime(100, 50) {
		t.Fatalf("early2")
	}
}
