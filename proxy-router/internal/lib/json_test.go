package lib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeJsonKeyOrder(t *testing.T) {
	msg1 := []byte(`{"firstKey": "firstValue", "secondKey": "secondValue"}`)
	msg2 := []byte(`{"secondKey": "secondValue", "firstKey": "firstValue"}`)
	require.NotEqual(t, msg1, msg2)

	normalized1, err := NormalizeJson(msg1)
	if err != nil {
		t.Fatal(err)
	}
	normalized2, err := NormalizeJson(msg2)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, normalized1, normalized2)
}

func TestNormalizeJsonIndentation(t *testing.T) {
	msg1 := []byte(`{"firstKey":   "firstValue", ` + "\n\n\n" + `"secondKey": "secondValue"}`)
	msg2 := []byte(`{"firstKey": "firstValue", "secondKey": "secondValue"}`)
	require.NotEqual(t, msg1, msg2)

	normalized1, err := NormalizeJson(msg1)
	if err != nil {
		t.Fatal(err)
	}
	normalized2, err := NormalizeJson(msg2)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, normalized1, normalized2)
}

func TestNormalizeJsonArrayOrder(t *testing.T) {
	msg1 := []byte(`[1,2,3]`)
	msg2 := []byte(`[3,2,1]`)
	require.NotEqual(t, msg1, msg2)

	normalized1, err := NormalizeJson(msg1)
	if err != nil {
		t.Fatal(err)
	}
	normalized2, err := NormalizeJson(msg2)
	if err != nil {
		t.Fatal(err)
	}

	require.NotEqual(t, normalized1, normalized2)
}
