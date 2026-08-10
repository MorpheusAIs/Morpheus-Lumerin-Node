package keychain

import (
	"errors"
	"testing"
)

// This test previously could not pass under any circumstances: the Set call was
// commented out, so it read back a key that had never been written and then
// asserted the result equalled "testval". Its deferred cleanup also began with
// a bare `return`, making the delete unreachable.
//
// It now performs a real round-trip, and skips when there is no usable keyring
// backend — which is the normal situation on a headless CI runner, where the
// OS keychain daemon is not running.
func TestKeychainRoundTrip(t *testing.T) {
	kc := NewKeychain()

	const (
		key = "keychain_test_roundtrip"
		val = "testval"
	)

	if err := kc.Insert(key, val); err != nil {
		t.Skipf("no usable keyring backend in this environment: %v", err)
	}
	t.Cleanup(func() {
		if err := kc.DeleteIfExists(key); err != nil {
			t.Errorf("cleanup: failed to delete %q: %v", key, err)
		}
	})

	got, err := kc.Get(key)
	if err != nil {
		t.Fatalf("Get(%q) failed: %v", key, err)
	}
	if got != val {
		t.Errorf("Get(%q) = %q, want %q", key, got, val)
	}
}

func TestKeychainUpsertOverwrites(t *testing.T) {
	kc := NewKeychain()

	const key = "keychain_test_upsert"

	if err := kc.Insert(key, "first"); err != nil {
		t.Skipf("no usable keyring backend in this environment: %v", err)
	}
	t.Cleanup(func() { _ = kc.DeleteIfExists(key) })

	if err := kc.Upsert(key, "second"); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	got, err := kc.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != "second" {
		t.Errorf("after Upsert, Get = %q, want %q", got, "second")
	}
}

// DeleteIfExists must swallow "not found" but still surface real errors —
// callers rely on it being safe to call unconditionally during cleanup.
func TestDeleteIfExistsOnMissingKeyIsNotAnError(t *testing.T) {
	kc := NewKeychain()

	// Probe for a backend first so a missing keyring reads as a skip rather
	// than a spurious pass.
	if _, err := kc.Get("keychain_test_probe"); err != nil && !errors.Is(err, ErrKeyNotFound) {
		t.Skipf("no usable keyring backend in this environment: %v", err)
	}

	if err := kc.DeleteIfExists("keychain_test_definitely_absent"); err != nil {
		t.Errorf("DeleteIfExists on a missing key returned %v, want nil", err)
	}
}
