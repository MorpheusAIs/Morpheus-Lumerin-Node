package system

import (
	"errors"
	"fmt"
	"unicode"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
)

// AdminUsername is the operator account seeded from the cookie file. It holds
// every permission and is never subject to agent allowances.
const AdminUsername = "admin"

// ErrUserChangeForbidden is returned when a caller tries to create, change,
// confirm or remove a user it does not dominate. Handlers map it to HTTP 403.
var ErrUserChangeForbidden = errors.New("forbidden user change")

// ErrInvalidUserSpec is returned for a username or permission list that cannot
// be stored safely. Handlers map it to HTTP 400.
var ErrInvalidUserSpec = errors.New("invalid user specification")

const maxUsernameLen = 128

// ValidateUserSpec rejects usernames and perms that would corrupt proxy.conf.
// The file is line based ("rpcauth=user:salt$hash", "rpcwhitelist=user:a,b"),
// so a newline, ':' or '$' inside a username, or a newline, ',' or ':' inside
// a perm, injects entries that take effect on the next restart: a username of
// "x:s$h\nrpcwhitelist=x:*" hands x every permission with a password of the
// attacker's choosing. gin does not enforce the validate: tags on the request
// structs, so this is the check.
func ValidateUserSpec(username string, perms []string) error {
	if username == "" || len(username) > maxUsernameLen {
		return fmt.Errorf("%w: username must be 1-%d characters", ErrInvalidUserSpec, maxUsernameLen)
	}
	for _, r := range username {
		if r <= ' ' || r == 0x7f || unicode.IsSpace(r) || r == ':' || r == '$' {
			return fmt.Errorf("%w: username contains forbidden character %q", ErrInvalidUserSpec, r)
		}
	}
	for _, p := range perms {
		if p == "" {
			return fmt.Errorf("%w: empty permission", ErrInvalidUserSpec)
		}
		for _, r := range p {
			if r <= ' ' || r == 0x7f || unicode.IsSpace(r) || r == ',' || r == ':' {
				return fmt.Errorf("%w: permission %q contains forbidden character %q", ErrInvalidUserSpec, p, r)
			}
		}
	}
	return nil
}

// HasFullAccess reports whether username may call every method: its whitelist
// contains "*", or it has no whitelist and rpcwhitelistdefault is on.
func (cfg *HTTPAuthConfig) HasFullAccess(username string) bool {
	if _, ok := cfg.AuthEntries[username]; !ok {
		return false
	}
	methods, found := cfg.Whitelists[username]
	if !found || len(methods) == 0 {
		return cfg.WhitelistDefault
	}
	for _, m := range methods {
		if m == "*" {
			return true
		}
	}
	return false
}

// requireKnownCaller refuses an empty or unregistered caller. CheckAuth
// guarantees a registered one on the HTTP path; this covers every other path.
func (cfg *HTTPAuthConfig) requireKnownCaller(caller string) error {
	if _, known := cfg.AuthEntries[caller]; caller == "" || !known {
		return fmt.Errorf("%w: unknown caller", ErrUserChangeForbidden)
	}
	return nil
}

// dominatesPerms reports whether caller may grant perms. A full-access caller
// dominates everything; any other caller must itself hold every perm, and "*"
// is never delegable. Returns the first offending perm.
func (cfg *HTTPAuthConfig) dominatesPerms(caller string, perms []string) (string, bool) {
	if cfg.HasFullAccess(caller) {
		return "", true
	}
	for _, p := range perms {
		if p == "*" || !cfg.IsMethodAllowed(caller, p) {
			return p, false
		}
	}
	return "", true
}

// agentRecord loads the AuthStorage record for username, tolerating a nil
// store (unit tests and embedders that run without agent support).
func (cfg *HTTPAuthConfig) agentRecord(username string) (*storages.AgentUser, error) {
	if cfg.AuthStorage == nil {
		return nil, nil
	}
	return cfg.AuthStorage.GetAgentUser(username)
}

// hasSpendingAllowance reports whether an agent record carries a non-zero cap
// for any token. Allowances are money, so only a full-access caller may create,
// approve or alter a user that has one.
func hasSpendingAllowance(rec *storages.AgentUser) bool {
	if rec == nil {
		return false
	}
	for _, a := range rec.Allowances {
		if a.Int.Sign() > 0 {
			return true
		}
	}
	return false
}

// authorizeTarget decides whether a scoped caller may touch target at all:
// never admin, never a user whose current perms the caller lacks, never a user
// that carries spending allowances (pending or confirmed).
func (cfg *HTTPAuthConfig) authorizeTarget(caller, target string) error {
	if target == AdminUsername {
		return fmt.Errorf("%w: only a full-access user may change %q", ErrUserChangeForbidden, AdminUsername)
	}
	if _, exists := cfg.AuthEntries[target]; exists {
		if cfg.HasFullAccess(target) {
			return fmt.Errorf("%w: %q cannot change full-access user %q", ErrUserChangeForbidden, caller, target)
		}
		if p, ok := cfg.dominatesPerms(caller, cfg.Whitelists[target]); !ok {
			return fmt.Errorf("%w: %q cannot change user %q, which holds permission %q", ErrUserChangeForbidden, caller, target, p)
		}
	}
	rec, err := cfg.agentRecord(target)
	if err != nil {
		return fmt.Errorf("error reading agent user: %w", err)
	}
	if hasSpendingAllowance(rec) {
		return fmt.Errorf("%w: %q cannot change %q, which carries spending allowances", ErrUserChangeForbidden, caller, target)
	}
	return nil
}

// authorizeGrant decides whether a scoped caller may hand out perms: only a
// subset of its own, and never an empty list while rpcwhitelistdefault=1,
// because a user without a whitelist is then unrestricted (a hidden "*").
func (cfg *HTTPAuthConfig) authorizeGrant(caller string, perms []string) error {
	if cfg.WhitelistDefault && len(perms) == 0 {
		return fmt.Errorf("%w: rpcwhitelistdefault=1 would leave the user unrestricted; list explicit permissions", ErrUserChangeForbidden)
	}
	if p, ok := cfg.dominatesPerms(caller, perms); !ok {
		return fmt.Errorf("%w: %q cannot grant permission %q", ErrUserChangeForbidden, caller, p)
	}
	return nil
}

// AuthorizeUserChange decides whether caller may create or overwrite target
// with requestedPerms. A full-access caller may do anything; any other caller
// is bound by authorizeTarget and authorizeGrant.
//
// This is what makes add_user and agent_requests delegable. Without it either
// permission was a one-request escalation to admin: POST /auth/users with
// username "admin" replaced the operator's credential, and admin bypasses
// agent allowances, so send_eth/send_mor then moved funds with no cap.
func (cfg *HTTPAuthConfig) AuthorizeUserChange(caller, target string, requestedPerms []string) error {
	if err := cfg.requireKnownCaller(caller); err != nil {
		return err
	}
	if cfg.HasFullAccess(caller) {
		return nil
	}
	if err := cfg.authorizeTarget(caller, target); err != nil {
		return err
	}
	return cfg.authorizeGrant(caller, requestedPerms)
}

// AuthorizeUserRemoval decides whether caller may remove target. The admin
// account can never be removed through the API, since that locks everyone out
// until proxy.conf is edited by hand. Other users follow authorizeTarget.
func (cfg *HTTPAuthConfig) AuthorizeUserRemoval(caller, target string) error {
	if target == AdminUsername {
		return fmt.Errorf("%w: %q cannot be removed", ErrUserChangeForbidden, AdminUsername)
	}
	if err := cfg.requireKnownCaller(caller); err != nil {
		return err
	}
	if cfg.HasFullAccess(caller) {
		return nil
	}
	return cfg.authorizeTarget(caller, target)
}

// ConfirmAgentUserAs confirms a pending agent request on behalf of caller.
// Requests are filed unauthenticated with perms and allowances of the filer's
// choosing, so a scoped caller is bound by AuthorizeUserChange (which refuses
// a request carrying allowances) and may not re-confirm an already confirmed
// agent, since re-confirming resets its password to the one in the request.
func (cfg *HTTPAuthConfig) ConfirmAgentUserAs(caller, username string) error {
	request, err := cfg.AuthStorage.GetAgentUser(username)
	if err != nil {
		return fmt.Errorf("error reading agent user: %w", err)
	}
	if request == nil {
		return fmt.Errorf("auth request not found")
	}
	if request.IsConfirmed && !cfg.HasFullAccess(caller) {
		return fmt.Errorf("%w: %q is already confirmed; only a full-access user may re-confirm it", ErrUserChangeForbidden, username)
	}
	if err := cfg.AuthorizeUserChange(caller, request.Username, request.Perms); err != nil {
		return err
	}
	return cfg.confirmAgentRequest(request)
}

// DeclineAgentUserAs declines (deletes) an agent request on behalf of caller.
// Declining a confirmed agent removes its allowance record and so blocks its
// spending, which only a full-access caller may do; a pending request may be
// declined by any caller that dominates its perms.
func (cfg *HTTPAuthConfig) DeclineAgentUserAs(caller, username string) error {
	if !cfg.HasFullAccess(caller) {
		if err := cfg.requireKnownCaller(caller); err != nil {
			return err
		}
		rec, err := cfg.AuthStorage.GetAgentUser(username)
		if err != nil {
			return fmt.Errorf("error reading agent user: %w", err)
		}
		if rec != nil {
			if rec.IsConfirmed {
				return fmt.Errorf("%w: only a full-access user may remove confirmed agent %q", ErrUserChangeForbidden, username)
			}
			if p, ok := cfg.dominatesPerms(caller, rec.Perms); !ok {
				return fmt.Errorf("%w: %q cannot decline a request holding permission %q", ErrUserChangeForbidden, caller, p)
			}
		}
	}
	return cfg.DeclineAgentUser(username)
}
