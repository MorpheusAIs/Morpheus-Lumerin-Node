package system

import (
	"math/big"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newDelegationTestConfig seeds admin (*) plus three scoped users:
// agent (add_user, remove_user, agent_requests, get_balance), ops (send_mor)
// and junior (get_balance).
func newDelegationTestConfig(t *testing.T) *HTTPAuthConfig {
	t.Helper()
	dir := t.TempDir()
	cfg := NewAuthConfig(filepath.Join(dir, "proxy.conf"), filepath.Join(dir, ".cookie"), "admin:admin-pw", storages.NewAuthStorage(storages.NewTestStorage()))
	require.NoError(t, cfg.EnsureConfigFilesExist())
	require.NoError(t, cfg.AddUser("agent", "agent-pw", []string{"add_user", "remove_user", "agent_requests", "get_balance"}))
	require.NoError(t, cfg.AddUser("ops", "ops-pw", []string{"send_mor"}))
	require.NoError(t, cfg.AddUser("junior", "junior-pw", []string{"get_balance"}))
	return cfg
}

func TestValidateUserSpec(t *testing.T) {
	assert.NoError(t, ValidateUserSpec("agent-1.bot_x@y", []string{"get_balance", "*"}))
	for _, u := range []string{"", "a b", "a:b", "a$b", "a\nb", "a\tb", "a b", strings.Repeat("a", 129)} {
		assert.ErrorIs(t, ValidateUserSpec(u, nil), ErrInvalidUserSpec, "username %q", u)
	}
	for _, p := range []string{"", "a b", "a,b", "a:b", "a\nb"} {
		assert.ErrorIs(t, ValidateUserSpec("ok", []string{p}), ErrInvalidUserSpec, "perm %q", p)
	}
}

// A username carrying a newline used to inject proxy.conf lines that gave the
// attacker "*" with a password of their choosing on the next restart.
func TestAddUser_RejectsProxyConfInjection(t *testing.T) {
	cfg := newDelegationTestConfig(t)
	payload := "evil:0011$2233\nrpcwhitelist=evil:*\nrpcauth=zz"
	assert.ErrorIs(t, cfg.AddUser(payload, "pw", []string{"get_balance"}), ErrInvalidUserSpec)
	assert.ErrorIs(t, cfg.RequestAgentUser(payload, "pw", []string{"get_balance"}, nil), ErrInvalidUserSpec)
	assert.ErrorIs(t, cfg.AddUser("ok", "pw", []string{"get_balance,send_mor"}), ErrInvalidUserSpec)

	reloaded := NewAuthConfig(cfg.FilePath, cfg.CookieFilePath, "", nil)
	require.NoError(t, reloaded.ReadConfig())
	_, exists := reloaded.AuthEntries["evil"]
	assert.False(t, exists)
}

func TestHasFullAccess(t *testing.T) {
	cfg := newDelegationTestConfig(t)
	assert.True(t, cfg.HasFullAccess("admin"))
	assert.False(t, cfg.HasFullAccess("agent"))
	assert.False(t, cfg.HasFullAccess("nobody"))

	require.NoError(t, cfg.AddUser("default", "pw", nil)) // no rpcwhitelist line
	assert.False(t, cfg.HasFullAccess("default"))
	cfg.WhitelistDefault = true
	assert.True(t, cfg.HasFullAccess("default"), "rpcwhitelistdefault=1 makes an unlisted user unrestricted")
}

func TestAuthorizeUserChange(t *testing.T) {
	cfg := newDelegationTestConfig(t)

	t.Run("full-access caller may do anything", func(t *testing.T) {
		assert.NoError(t, cfg.AuthorizeUserChange("admin", "admin", []string{"*"}))
		assert.NoError(t, cfg.AuthorizeUserChange("admin", "ops", []string{"*"}))
	})

	t.Run("scoped caller cannot touch admin", func(t *testing.T) {
		assert.ErrorIs(t, cfg.AuthorizeUserChange("agent", "admin", []string{"add_user"}), ErrUserChangeForbidden)
		assert.ErrorIs(t, cfg.AuthorizeUserChange("agent", "admin", nil), ErrUserChangeForbidden)
	})

	t.Run("scoped caller cannot grant * or perms it lacks", func(t *testing.T) {
		assert.ErrorIs(t, cfg.AuthorizeUserChange("agent", "new", []string{"*"}), ErrUserChangeForbidden)
		assert.ErrorIs(t, cfg.AuthorizeUserChange("agent", "new", []string{"send_mor"}), ErrUserChangeForbidden)
		assert.ErrorIs(t, cfg.AuthorizeUserChange("agent", "new", []string{"get_balance", "send_eth"}), ErrUserChangeForbidden)
	})

	t.Run("scoped caller may grant a subset of its own perms", func(t *testing.T) {
		assert.NoError(t, cfg.AuthorizeUserChange("agent", "new", []string{"get_balance", "add_user"}))
		assert.NoError(t, cfg.AuthorizeUserChange("agent", "new", nil), "no perms = no access while rpcwhitelistdefault=0")
	})

	t.Run("scoped caller cannot overwrite a user holding perms it lacks", func(t *testing.T) {
		assert.ErrorIs(t, cfg.AuthorizeUserChange("agent", "ops", []string{"get_balance"}), ErrUserChangeForbidden)
	})

	t.Run("scoped caller may overwrite users it dominates, including itself", func(t *testing.T) {
		assert.NoError(t, cfg.AuthorizeUserChange("agent", "junior", []string{"get_balance"}))
		assert.NoError(t, cfg.AuthorizeUserChange("agent", "agent", []string{"add_user"}))
	})

	t.Run("unknown caller is refused", func(t *testing.T) {
		assert.ErrorIs(t, cfg.AuthorizeUserChange("", "new", nil), ErrUserChangeForbidden)
		assert.ErrorIs(t, cfg.AuthorizeUserChange("ghost", "new", nil), ErrUserChangeForbidden)
	})
}

func TestAuthorizeUserChange_WhitelistDefaultNeedsExplicitPerms(t *testing.T) {
	cfg := newDelegationTestConfig(t)
	cfg.WhitelistDefault = true
	assert.ErrorIs(t, cfg.AuthorizeUserChange("agent", "new", nil), ErrUserChangeForbidden)
	assert.ErrorIs(t, cfg.AuthorizeUserChange("agent", "new", []string{}), ErrUserChangeForbidden)
	assert.NoError(t, cfg.AuthorizeUserChange("agent", "new", []string{"get_balance"}))
	assert.NoError(t, cfg.AuthorizeUserChange("admin", "new", nil))
}

func TestAuthorizeUserChange_AllowancesNeedFullAccess(t *testing.T) {
	cfg := newDelegationTestConfig(t)
	// A pending request with a cap, filed through the unauthenticated endpoint.
	require.NoError(t, cfg.RequestAgentUser("rich", "pw", []string{"get_balance"}, map[string]string{"eth": "1000"}))
	assert.ErrorIs(t, cfg.AuthorizeUserChange("agent", "rich", []string{"get_balance"}), ErrUserChangeForbidden)
	assert.NoError(t, cfg.AuthorizeUserChange("admin", "rich", []string{"get_balance"}))

	// A zero cap is not a cap.
	require.NoError(t, cfg.RequestAgentUser("poor", "pw", []string{"get_balance"}, map[string]string{"eth": "0"}))
	assert.NoError(t, cfg.AuthorizeUserChange("agent", "poor", []string{"get_balance"}))
}

func TestAuthorizeUserRemoval(t *testing.T) {
	cfg := newDelegationTestConfig(t)
	assert.ErrorIs(t, cfg.AuthorizeUserRemoval("admin", "admin"), ErrUserChangeForbidden, "admin is never removable")
	assert.ErrorIs(t, cfg.AuthorizeUserRemoval("agent", "admin"), ErrUserChangeForbidden)
	assert.NoError(t, cfg.AuthorizeUserRemoval("admin", "ops"))
	assert.ErrorIs(t, cfg.AuthorizeUserRemoval("agent", "ops"), ErrUserChangeForbidden, "ops holds send_mor, agent does not")
	assert.NoError(t, cfg.AuthorizeUserRemoval("agent", "junior"))
	assert.NoError(t, cfg.AuthorizeUserRemoval("agent", "missing"))

	// Removal grants nothing, so the rpcwhitelistdefault=1 empty-perms rule
	// must not apply; a user with no whitelist is then full-access, though.
	cfg.WhitelistDefault = true
	require.NoError(t, cfg.AddUser("unlisted", "pw", nil))
	assert.NoError(t, cfg.AuthorizeUserRemoval("agent", "junior"))
	assert.ErrorIs(t, cfg.AuthorizeUserRemoval("agent", "unlisted"), ErrUserChangeForbidden)
}

func TestIsAllowanceEnough_RequiresConfirmedAgent(t *testing.T) {
	cfg := newDelegationTestConfig(t)
	require.NoError(t, cfg.RequestAgentUser("rich", "pw", []string{"send_eth"}, map[string]string{"eth": "1000"}))

	// Pending: the cap was chosen by the filer and must not be spendable.
	_, err := cfg.IsAllowanceEnough("rich", "eth", big.NewInt(1))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not confirmed")

	require.NoError(t, cfg.ConfirmAgentUserAs("admin", "rich"))
	ok, err := cfg.IsAllowanceEnough("rich", "eth", big.NewInt(1))
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestConfirmAgentUserAs_RequiresDominance(t *testing.T) {
	cfg := newDelegationTestConfig(t)
	require.NoError(t, cfg.RequestAgentUser("bot", "bot-pw", []string{"*"}, nil))

	err := cfg.ConfirmAgentUserAs("agent", "bot")
	assert.ErrorIs(t, err, ErrUserChangeForbidden)
	assert.False(t, cfg.ValidatePassword("bot", "bot-pw"), "a refused confirmation must not create the user")

	require.NoError(t, cfg.ConfirmAgentUserAs("admin", "bot"))
	assert.True(t, cfg.ValidatePassword("bot", "bot-pw"))
	assert.True(t, cfg.HasFullAccess("bot"))

	require.NoError(t, cfg.RequestAgentUser("helper", "helper-pw", []string{"get_balance"}, nil))
	require.NoError(t, cfg.ConfirmAgentUserAs("agent", "helper"))
	assert.True(t, cfg.IsMethodAllowed("helper", "get_balance"))
	assert.False(t, cfg.IsMethodAllowed("helper", "send_mor"))
}

func TestConfirmAgentUserAs_AllowanceAndReconfirmNeedFullAccess(t *testing.T) {
	cfg := newDelegationTestConfig(t)
	require.NoError(t, cfg.RequestAgentUser("rich", "pw", []string{"get_balance"}, map[string]string{"eth": "5"}))
	assert.ErrorIs(t, cfg.ConfirmAgentUserAs("agent", "rich"), ErrUserChangeForbidden)
	require.NoError(t, cfg.ConfirmAgentUserAs("admin", "rich"))

	require.NoError(t, cfg.RequestAgentUser("plain", "pw", []string{"get_balance"}, nil))
	require.NoError(t, cfg.ConfirmAgentUserAs("agent", "plain"))
	// Admin rotates the password; a scoped re-confirm would restore the old one.
	require.NoError(t, cfg.AddUser("plain", "rotated", []string{"get_balance"}))
	assert.ErrorIs(t, cfg.ConfirmAgentUserAs("agent", "plain"), ErrUserChangeForbidden)
	assert.True(t, cfg.ValidatePassword("plain", "rotated"))
}

func TestDeclineAgentUserAs(t *testing.T) {
	cfg := newDelegationTestConfig(t)
	require.NoError(t, cfg.RequestAgentUser("pending-star", "pw", []string{"*"}, nil))
	assert.ErrorIs(t, cfg.DeclineAgentUserAs("agent", "pending-star"), ErrUserChangeForbidden)

	require.NoError(t, cfg.RequestAgentUser("pending-ok", "pw", []string{"get_balance"}, nil))
	require.NoError(t, cfg.DeclineAgentUserAs("agent", "pending-ok"))

	require.NoError(t, cfg.RequestAgentUser("confirmed", "pw", []string{"get_balance"}, nil))
	require.NoError(t, cfg.ConfirmAgentUserAs("admin", "confirmed"))
	assert.ErrorIs(t, cfg.DeclineAgentUserAs("agent", "confirmed"), ErrUserChangeForbidden, "declining a confirmed agent blocks its spending")
	require.NoError(t, cfg.DeclineAgentUserAs("admin", "confirmed"))
}
