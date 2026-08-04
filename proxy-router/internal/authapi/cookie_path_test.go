package authapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/gin-gonic/gin"
)

func TestGetPathToCookieFileDoesNotDoubleJoinAbsolute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	abs := filepath.Join(t.TempDir(), ".cookie")

	authCfg := system.NewAuthConfig("./proxy.conf", abs, "", nil)
	ctrl := NewAuthController(authCfg, "production", nil)

	r := gin.New()
	r.GET("/auth/cookie/path", ctrl.GetPathToCookieFile)

	req := httptest.NewRequest(http.MethodGet, "/auth/cookie/path", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}

	var body struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Path != filepath.Clean(abs) {
		t.Fatalf("got %q want %q", body.Path, filepath.Clean(abs))
	}
}
