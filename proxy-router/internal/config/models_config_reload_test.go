package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type stubValidator struct{}

func (stubValidator) Struct(s interface{}) error { return nil }

type stubChain struct{}

func (stubChain) ModelExists(ctx context.Context, ID common.Hash) (bool, error) { return true, nil }

type stubConn struct{}

func (stubConn) TryConnect(ctx context.Context, url string) error { return nil }

const modelA = "0x1111111111111111111111111111111111111111111111111111111111111111"
const modelB = "0x2222222222222222222222222222222222222222222222222222222222222222"

func writeModels(t *testing.T, path string, ids ...string) {
	t.Helper()
	body := `{"models":[`
	for i, id := range ids {
		if i > 0 {
			body += ","
		}
		body += `{"modelId":"` + id + `","modelName":"m-` + id[2:6] + `","apiType":"openai","apiUrl":"http://127.0.0.1:1"}`
	}
	body += `]}`
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func newTestLoader(t *testing.T, path string) *ModelConfigLoader {
	t.Helper()
	l := NewModelConfigLoader(path, "", stubValidator{}, stubChain{}, stubConn{}, lib.NewTestLogger())
	if err := l.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return l
}

// A model added to the file becomes visible after Reload without a restart,
// and one removed from the file disappears.
func TestReloadAddsAndRemovesModels(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models-config.json")
	writeModels(t, path, modelA)
	l := newTestLoader(t, path)

	if got := l.ModelConfigFromID(modelB).ModelName; got != "" {
		t.Fatalf("model B visible before reload: %q", got)
	}

	writeModels(t, path, modelA, modelB)
	added, removed, err := l.Reload()
	if err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if len(added) != 1 || added[0] != modelB || len(removed) != 0 {
		t.Fatalf("added=%v removed=%v, want added=[B] removed=[]", added, removed)
	}
	if got := l.ModelConfigFromID(modelB).ModelName; got == "" {
		t.Fatal("model B not visible after reload")
	}
	if ids, _ := l.GetAll(); len(ids) != 2 {
		t.Fatalf("GetAll after reload: %d models, want 2", len(ids))
	}

	writeModels(t, path, modelB)
	added, removed, err = l.Reload()
	if err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if len(added) != 0 || len(removed) != 1 || removed[0] != modelA {
		t.Fatalf("added=%v removed=%v, want added=[] removed=[A]", added, removed)
	}
	if got := l.ModelConfigFromID(modelA).ModelName; got != "" {
		t.Fatalf("model A still visible after removal: %q", got)
	}
}

// A file that no longer parses must not empty a serving node: Reload returns
// the error and the previous table stays in force.
func TestReloadKeepsPreviousTableOnBadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models-config.json")
	writeModels(t, path, modelA)
	l := newTestLoader(t, path)

	if err := os.WriteFile(path, []byte(`{"models":[{"modelId":`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Reload(); err == nil {
		t.Fatal("Reload accepted a broken file")
	}
	if got := l.ModelConfigFromID(modelA).ModelName; got == "" {
		t.Fatal("previous table lost after a failed reload")
	}
}
