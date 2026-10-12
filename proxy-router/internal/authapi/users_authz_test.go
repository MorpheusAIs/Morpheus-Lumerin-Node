package authapi

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authzFixture struct {
	cfg        *system.HTTPAuthConfig
	router     *gin.Engine
	cookiePath string
}

// newAuthzFixture seeds admin (*), agent (add_user, remove_user,
// agent_requests) and ops (send_mor) behind the real auth routes.
func newAuthzFixture(t *testing.T) *authzFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	cookiePath := filepath.Join(dir, ".cookie")
	cfg := system.NewAuthConfig(filepath.Join(dir, "proxy.conf"), cookiePath, "admin:admin-pw", storages.NewAuthStorage(storages.NewTestStorage()))
	require.NoError(t, cfg.EnsureConfigFilesExist())
	require.NoError(t, cfg.AddUser("agent", "agent-pw", []string{"add_user", "remove_user", "agent_requests"}))
	require.NoError(t, cfg.AddUser("ops", "ops-pw", []string{"send_mor"}))

	r := gin.New()
	NewAuthController(cfg, "test", lib.NewTestLogger()).RegisterRoutes(r)
	return &authzFixture{cfg: cfg, router: r, cookiePath: cookiePath}
}

func (f *authzFixture) do(method, path, user, pass, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if user != "" {
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(user+":"+pass)))
	}
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

// Regression for the bounty report: a user holding only add_user could POST
// username "admin" and replace the operator credential (and the .cookie).
func TestAddUser_ScopedCallerCannotOverwriteAdmin(t *testing.T) {
	f := newAuthzFixture(t)
	w := f.do(http.MethodPost, "/auth/users", "agent", "agent-pw", `{"username":"admin","password":"PWNED","perms":["*"]}`)
	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	assert.True(t, f.cfg.ValidatePassword("admin", "admin-pw"))
	assert.False(t, f.cfg.ValidatePassword("admin", "PWNED"))
	cookie, err := os.ReadFile(f.cookiePath)
	require.NoError(t, err)
	assert.Equal(t, "admin:admin-pw", string(cookie))
}

func TestAddUser_ScopedCallerCannotEscalatePerms(t *testing.T) {
	f := newAuthzFixture(t)
	for _, body := range []string{
		`{"username":"backdoor","password":"x","perms":["*"]}`,
		`{"username":"backdoor","password":"x","perms":["send_mor"]}`,
		`{"username":"ops","password":"x","perms":["add_user"]}`, // ops holds send_mor, which agent lacks
	} {
		w := f.do(http.MethodPost, "/auth/users", "agent", "agent-pw", body)
		assert.Equal(t, http.StatusForbidden, w.Code, body)
	}
	assert.False(t, f.cfg.ValidatePassword("backdoor", "x"))
	assert.True(t, f.cfg.ValidatePassword("ops", "ops-pw"))
}

func TestAddUser_ScopedCallerMayDelegateSubset(t *testing.T) {
	f := newAuthzFixture(t)
	w := f.do(http.MethodPost, "/auth/users", "agent", "agent-pw", `{"username":"sub","password":"sub-pw","perms":["add_user"]}`)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.True(t, f.cfg.ValidatePassword("sub", "sub-pw"))
	assert.True(t, f.cfg.IsMethodAllowed("sub", "add_user"))
	assert.False(t, f.cfg.IsMethodAllowed("sub", "remove_user"))
}

func TestAddUser_AdminKeepsFullControl(t *testing.T) {
	f := newAuthzFixture(t)
	w := f.do(http.MethodPost, "/auth/users", "admin", "admin-pw", `{"username":"admin","password":"rotated","perms":["*"]}`)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.True(t, f.cfg.ValidatePassword("admin", "rotated"))
	cookie, err := os.ReadFile(f.cookiePath)
	require.NoError(t, err)
	assert.Equal(t, "admin:rotated\n", string(cookie))

	w = f.do(http.MethodPost, "/auth/users", "admin", "rotated", `{"username":"ops","password":"new","perms":["*"]}`)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.True(t, f.cfg.HasFullAccess("ops"))
}

// proxy.conf is line-oriented; a newline in a username injected entries that
// took effect on the next restart.
func TestAddUser_RejectsProxyConfInjection(t *testing.T) {
	f := newAuthzFixture(t)
	w := f.do(http.MethodPost, "/auth/users", "admin", "admin-pw", `{"username":"evil:a$b\nrpcwhitelist=evil:*","password":"x","perms":["get_balance"]}`)
	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())

	w = f.do(http.MethodPost, "/auth/users/request", "", "", `{"username":"evil\nrpcwhitelist=evil:*","password":"x","perms":["get_balance"],"allowances":{}}`)
	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())

	w = f.do(http.MethodPost, "/auth/users", "admin", "admin-pw", `{"username":"ok","password":"x","perms":["get_balance,send_mor"]}`)
	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}

func TestDeleteUser_AdminIsNeverRemovable(t *testing.T) {
	f := newAuthzFixture(t)
	for _, c := range [][2]string{{"admin", "admin-pw"}, {"agent", "agent-pw"}} {
		w := f.do(http.MethodDelete, "/auth/users", c[0], c[1], `{"username":"admin"}`)
		assert.Equal(t, http.StatusForbidden, w.Code, c[0]+": "+w.Body.String())
	}
	assert.True(t, f.cfg.ValidatePassword("admin", "admin-pw"))
}

func TestDeleteUser_ScopedCallerFollowsDominance(t *testing.T) {
	f := newAuthzFixture(t)
	w := f.do(http.MethodDelete, "/auth/users", "agent", "agent-pw", `{"username":"ops"}`)
	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	assert.True(t, f.cfg.ValidatePassword("ops", "ops-pw"))

	require.NoError(t, f.cfg.AddUser("junior", "junior-pw", []string{"add_user"}))
	w = f.do(http.MethodDelete, "/auth/users", "agent", "agent-pw", `{"username":"junior"}`)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.False(t, f.cfg.ValidatePassword("junior", "junior-pw"))
}

// Requests are filed unauthenticated, so confirming one must be bounded by
// the confirmer's own perms or agent_requests is a second route to admin.
func TestConfirmAgentRequest_RequiresDominance(t *testing.T) {
	f := newAuthzFixture(t)
	w := f.do(http.MethodPost, "/auth/users/request", "", "", `{"username":"bot","password":"bot-pw","perms":["*"],"allowances":{}}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	w = f.do(http.MethodPost, "/auth/users/confirm", "agent", "agent-pw", `{"username":"bot","confirm":true}`)
	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	assert.False(t, f.cfg.ValidatePassword("bot", "bot-pw"))

	w = f.do(http.MethodPost, "/auth/users/confirm", "admin", "admin-pw", `{"username":"bot","confirm":true}`)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.True(t, f.cfg.ValidatePassword("bot", "bot-pw"))
}

func TestConfirmAgentRequest_AllowanceNeedsFullAccess(t *testing.T) {
	f := newAuthzFixture(t)
	w := f.do(http.MethodPost, "/auth/users/request", "", "", `{"username":"rich","password":"pw","perms":[],"allowances":{"eth":"1000"}}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	w = f.do(http.MethodPost, "/auth/users/confirm", "agent", "agent-pw", `{"username":"rich","confirm":true}`)
	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())

	// Declining a confirmed agent (which removes its allowance) also needs full access.
	w = f.do(http.MethodPost, "/auth/users/confirm", "admin", "admin-pw", `{"username":"rich","confirm":true}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	w = f.do(http.MethodPost, "/auth/users/confirm", "agent", "agent-pw", `{"username":"rich","confirm":false}`)
	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
}

func TestAllowanceEndpoints_NeedFullAccess(t *testing.T) {
	f := newAuthzFixture(t)
	w := f.do(http.MethodPost, "/auth/users/request", "", "", `{"username":"spender","password":"sp-pw","perms":["request_allowance"],"allowances":{}}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	w = f.do(http.MethodPost, "/auth/users/confirm", "admin", "admin-pw", `{"username":"spender","confirm":true}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	// An agent may only ask for itself.
	w = f.do(http.MethodPost, "/auth/allowance/requests", "spender", "sp-pw", `{"username":"ops","token":"eth","allowance":"10"}`)
	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	w = f.do(http.MethodPost, "/auth/allowance/requests", "spender", "sp-pw", `{"username":"spender","token":"eth","allowance":"10"}`)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	// agent_requests alone neither approves nor revokes money.
	w = f.do(http.MethodPost, "/auth/allowance/confirm", "agent", "agent-pw", `{"username":"spender","token":"eth","confirm":true}`)
	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	w = f.do(http.MethodPost, "/auth/allowance/revoke", "agent", "agent-pw", `{"username":"spender","token":"eth"}`)
	assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())

	w = f.do(http.MethodPost, "/auth/allowance/confirm", "admin", "admin-pw", `{"username":"spender","token":"eth","confirm":true}`)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	w = f.do(http.MethodPost, "/auth/allowance/revoke", "admin", "admin-pw", `{"username":"spender","token":"eth"}`)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
}
