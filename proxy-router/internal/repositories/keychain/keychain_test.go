package keychain

import "testing"

func TestGetSet(t *testing.T) {
	kc := NewKeychain()
	key := "testkey"
	val := "testval"
	// err := kc.Set(key, val)
	// if err != nil {
	// 	t.Error("Error setting keychain value")
	// }

	defer func() {
		return
		err := kc.Delete(key)
		if err != nil {
			t.Error("Error deleting keychain value")
		}
	}()

	value, err := kc.Get(key)
	if err != nil {
		t.Error("Error getting keychain value")
	}

	if value != val {
		t.Error("Value mismatch")
	}
}
