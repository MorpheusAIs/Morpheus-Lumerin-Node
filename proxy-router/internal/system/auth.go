package system

import (
	"bufio"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/gin-gonic/gin"
)

// HTTPAuthEntry holds the parsed username, salt, and hash from rpcauth lines.
type HTTPAuthEntry struct {
	Username string
	Salt     string
	Hash     string
}

// HTTPAuthConfig is our main struct to manage the config file and (optionally) a cookie file.
type HTTPAuthConfig struct {
	FilePath         string
	CookieFilePath   string
	CookieContent    string
	AuthEntries      map[string]*HTTPAuthEntry // keyed by username
	Whitelists       map[string][]string       // keyed by username
	WhitelistDefault bool                      // true => methods allowed if not listed

	AuthStorage *storages.AuthStorage
}

type AgentTx struct {
	TxHash   string
	Username string
}

// NewAuthConfig initializes an empty RPCConfig struct, pointing to config + cookie paths.
func NewAuthConfig(configFilePath, cookieFilePath string, cookieContent string, authStorage *storages.AuthStorage) *HTTPAuthConfig {
	return &HTTPAuthConfig{
		FilePath:         configFilePath,
		CookieFilePath:   cookieFilePath,
		CookieContent:    cookieContent,
		AuthEntries:      make(map[string]*HTTPAuthEntry),
		Whitelists:       make(map[string][]string),
		WhitelistDefault: false,
		AuthStorage:      authStorage,
	}
}

func (cfg *HTTPAuthConfig) ReadCookieFile() (string, string, error) {
	data, err := os.ReadFile(cfg.CookieFilePath)
	if err != nil {
		return "", "", err
	}
	line := strings.TrimSpace(string(data))
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid cookie format")
	}
	return parts[0], parts[1], nil
}

// ReadConfig reads and parses the config file from disk.
func (cfg *HTTPAuthConfig) ReadConfig() error {
	file, err := os.Open(cfg.FilePath)
	if err != nil {
		// If file doesn't exist, that's okay—we'll create it later.
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines or comments
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		// rpcauth=user:salt$hash
		if strings.HasPrefix(line, "rpcauth=") {
			entryLine := strings.TrimPrefix(line, "rpcauth=")
			parts := strings.SplitN(entryLine, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid rpcauth line: %s", line)
			}
			user := parts[0]
			saltHash := parts[1]
			shParts := strings.SplitN(saltHash, "$", 2)
			if len(shParts) != 2 {
				return fmt.Errorf("invalid salt$hash in rpcauth line: %s", line)
			}

			salt := shParts[0]
			hashVal := shParts[1]

			cfg.AuthEntries[user] = &HTTPAuthEntry{
				Username: user,
				Salt:     salt,
				Hash:     hashVal,
			}
		}

		// rpcwhitelist=user:method1,method2
		if strings.HasPrefix(line, "rpcwhitelist=") {
			wLine := strings.TrimPrefix(line, "rpcwhitelist=")
			parts := strings.SplitN(wLine, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid rpcwhitelist line: %s", line)
			}
			user := parts[0]
			methods := strings.Split(parts[1], ",")
			for i := range methods {
				methods[i] = strings.TrimSpace(methods[i])
			}
			cfg.Whitelists[user] = methods
		}

		// rpcwhitelistdefault=0 or 1
		if strings.HasPrefix(line, "rpcwhitelistdefault=") {
			valStr := strings.TrimPrefix(line, "rpcwhitelistdefault=")
			// Attempt bool parse. If that fails, handle "0"/"1".
			b, errBool := strconv.ParseBool(valStr)
			if errBool != nil {
				// The config might use "0" or "1"
				if valStr == "0" {
					b = false
				} else if valStr == "1" {
					b = true
				} else {
					return fmt.Errorf("invalid rpcwhitelistdefault value: %s", valStr)
				}
			}
			cfg.WhitelistDefault = b
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// WriteConfig writes the current in-memory config to disk with file perms 0600.
func (cfg *HTTPAuthConfig) WriteConfig() error {
	file, err := os.OpenFile(cfg.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write rpcauth lines
	for _, entry := range cfg.AuthEntries {
		line := fmt.Sprintf("rpcauth=%s:%s$%s\n", entry.Username, entry.Salt, entry.Hash)
		if _, err := file.WriteString(line); err != nil {
			return err
		}
	}
	// Write whitelist lines
	for user, methods := range cfg.Whitelists {
		if len(methods) == 0 {
			continue
		}
		line := fmt.Sprintf("rpcwhitelist=%s:%s\n", user, strings.Join(methods, ","))
		if _, err := file.WriteString(line); err != nil {
			return err
		}
	}
	// Write whitelist default
	whDefault := "0"
	if cfg.WhitelistDefault {
		whDefault = "1"
	}
	line := fmt.Sprintf("rpcwhitelistdefault=%s\n", whDefault)
	if _, err := file.WriteString(line); err != nil {
		return err
	}

	return nil
}

// CheckFilePermissions ensures the config file is only readable by the owner (chmod 600).
func (cfg *HTTPAuthConfig) CheckFilePermissions() error {
	info, err := os.Stat(cfg.FilePath)
	if err != nil {
		return err
	}
	mode := info.Mode().Perm()
	if mode != 0600 {
		return fmt.Errorf("file permissions are not 0600, found: %o", mode)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Cookie File Management
// ----------------------------------------------------------------------------

// EnsureCookieFileExists checks if cookie file exists; if not, creates it with admin credentials.
func (cfg *HTTPAuthConfig) EnsureCookieFileExists() error {
	if _, err := os.Stat(cfg.CookieFilePath); os.IsNotExist(err) {
		// Cookie file: "admin:<password>"
		var cookieLine string
		var pass string
		if cfg.CookieContent != "" {
			cookieLine = cfg.CookieContent
			parts := strings.SplitN(cfg.CookieContent, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid cookie content: %s", cfg.CookieContent)
			}
			pass = parts[1]
		} else {
			// Generate a random password
			pass, err = generateRandomString(32)
			if err != nil {
				return fmt.Errorf("failed generating cookie password: %v", err)
			}
			cookieLine = fmt.Sprintf("admin:%s\n", pass)
		}

		// Write cookie file with perms 0600
		f, errCreate := os.OpenFile(cfg.CookieFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if errCreate != nil {
			return fmt.Errorf("could not create cookie file: %v", errCreate)
		}
		if _, errWrite := f.WriteString(cookieLine); errWrite != nil {
			_ = f.Close()
			return fmt.Errorf("failed writing cookie file: %v", errWrite)
		}
		f.Close()

		if err := cfg.AddUser("admin", pass, []string{"*"}); err != nil {
			return fmt.Errorf("failed to add admin rpcauth: %v", err)
		}
	}

	return nil
}

func (cfg *HTTPAuthConfig) AddUser(username string, plaintextPassword string, perms []string) error {
	saltBytes := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, saltBytes); err != nil {
		return err
	}
	saltHex := hex.EncodeToString(saltBytes)

	// Compute HMAC-SHA256(salt, password)
	h := hmac.New(sha256.New, saltBytes)
	h.Write([]byte(plaintextPassword))
	hashHex := hex.EncodeToString(h.Sum(nil))

	// Save to our in-memory config
	cfg.AuthEntries[username] = &HTTPAuthEntry{
		Username: username,
		Salt:     saltHex,
		Hash:     hashHex,
	}
	if perms != nil {
		cfg.Whitelists[username] = perms
	}

	if errWriteCfg := cfg.WriteConfig(); errWriteCfg != nil {
		return fmt.Errorf("failed writing config after cookie creation: %v", errWriteCfg)
	}

	return nil
}

func (cfg *HTTPAuthConfig) UpdateCookieContent(cookieLine string) error {
	if err := os.WriteFile(cfg.CookieFilePath, []byte(cookieLine), 0600); err != nil {
		return fmt.Errorf("failed updating cookie file: %v", err)
	}
	return nil
}

func (cfg *HTTPAuthConfig) RemoveUser(username string) error {
	delete(cfg.AuthEntries, username)
	delete(cfg.Whitelists, username)

	if errWriteCfg := cfg.WriteConfig(); errWriteCfg != nil {
		return fmt.Errorf("failed writing config after user removal: %v", errWriteCfg)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Utilities
// ----------------------------------------------------------------------------

// generateRandomString uses a simple approach to create alphanumeric random strings of length n.
func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		randNum, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		result[i] = letters[randNum.Int64()]
	}
	return string(result), nil
}

// ValidatePassword checks a user's plaintext password against the HMAC-based hash.
func (cfg *HTTPAuthConfig) ValidatePassword(username, password string) bool {
	entry, ok := cfg.AuthEntries[username]
	if !ok {
		return false
	}
	// Decode the hex salt
	saltBytes, err := hex.DecodeString(entry.Salt)
	if err != nil {
		return false
	}

	// Recompute HMAC-SHA256
	h := hmac.New(sha256.New, saltBytes)
	h.Write([]byte(password))
	computed := h.Sum(nil)

	storedHash, err := hex.DecodeString(entry.Hash)
	if err != nil {
		return false
	}

	return hmac.Equal(computed, storedHash)
}

func (cfg *HTTPAuthConfig) ParseBasicAuthHeader(header string) (string, string) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return "", ""
	}
	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", ""
	}
	creds := strings.SplitN(string(decoded), ":", 2)
	if len(creds) != 2 {
		return "", ""
	}
	return creds[0], creds[1]
}

// IsMethodAllowed checks if a given user is allowed to call a specific RPC method.
func (cfg *HTTPAuthConfig) IsMethodAllowed(username, method string) bool {
	_, ok := cfg.AuthEntries[username]
	if !ok {
		return false
	}

	// If user has a whitelist entry, check if the method is in that list
	methods, found := cfg.Whitelists[username]
	if found && len(methods) > 0 {
		for _, m := range methods {
			if m == "*" || m == method {
				return true
			}
		}
		return false
	}

	// If user has no explicit whitelist, fallback to WhitelistDefault
	return cfg.WhitelistDefault
}

func (cfg *HTTPAuthConfig) CheckAuth(method string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		basicAuth := ctx.GetHeader("Authorization")
		if basicAuth == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no basic auth provided"})
			return
		}

		username, password := cfg.ParseBasicAuthHeader(basicAuth)
		if username == "" || password == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid basic auth provided"})
			return
		}

		ok := cfg.ValidatePassword(username, password)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}

		result := cfg.IsMethodAllowed(username, method)
		if !result {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "method not allowed"})
			return
		}

		ctx.Set("username", username)
	}
}

func (cfg *HTTPAuthConfig) IsAllowanceEnough(username string, token string, amount *big.Int) (bool, error) {
	agentUser, ok := cfg.AuthStorage.GetAgentUser(username)
	if username == "admin" {
		return false, nil
	}
	if !ok {
		return false, fmt.Errorf("user not found")
	}

	token = strings.ToLower(token)
	allowance := agentUser.Allowances[token]
	if amount.Cmp(&allowance.Int) > 0 {
		return false, fmt.Errorf("not enough allowance")
	}
	return true, nil
}

func (cfg *HTTPAuthConfig) RequestAgentUser(username, password string, perms []string, allowances map[string]lib.BigInt) error {
	if username == "admin" {
		return fmt.Errorf("admin is a reserved username")
	}

	_, ok := cfg.AuthStorage.GetAgentUser(username)
	if ok {
		return fmt.Errorf("username %s already has an active request", username)
	}

	err := cfg.AuthStorage.AddAuthRequest(&storages.AgentUser{
		Username:    username,
		Password:    password,
		Perms:       perms,
		Allowances:  allowances,
		IsConfirmed: false,
	})
	if err != nil {
		return err
	}
	return nil
}

func (cfg *HTTPAuthConfig) ConfirmAgentUser(username string) error {
	request, ok := cfg.AuthStorage.GetAgentUser(username)
	if !ok {
		return fmt.Errorf("auth request not found")
	}

	err := cfg.AddUser(request.Username, request.Password, request.Perms)
	if err != nil {
		return err
	}

	request.IsConfirmed = true
	err = cfg.AuthStorage.AddAuthRequest(request)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *HTTPAuthConfig) DeclineAgentUser(username string) error {
	err := cfg.AuthStorage.DeleteAuthRequest(username)
	if err != nil {
		return err
	}
	return nil
}

func (cfg *HTTPAuthConfig) GetAgentUsers() ([]*storages.AgentUser, error) {
	requests, err := cfg.AuthStorage.GetAgentUsers()
	if err != nil {
		return nil, err
	}
	if requests == nil {
		return []*storages.AgentUser{}, nil
	}
	return requests, nil
}

func (cfg *HTTPAuthConfig) RemoveAgentUser(username string) error {
	err := cfg.RemoveUser(username)
	if err != nil {
		return err
	}

	_, ok := cfg.AuthStorage.GetAgentUser(username)
	if !ok {
		return nil
	}

	return cfg.AuthStorage.DeleteAuthRequest(username)
}

func (cfg *HTTPAuthConfig) RequestAllowance(username string, token string, allowance lib.BigInt) error {
	token = strings.ToLower(token)
	return cfg.AuthStorage.AddAllowanceRequest(&storages.AllowanceRequest{
		Username:  username,
		Token:     token,
		Allowance: allowance,
	})
}

func (cfg *HTTPAuthConfig) GetAllowanceRequests() ([]*storages.AllowanceRequest, error) {
	requests, err := cfg.AuthStorage.GetAllowanceRequests()
	if err != nil {
		return nil, err
	}
	if requests == nil {
		return []*storages.AllowanceRequest{}, nil
	}
	return requests, nil
}

func (cfg *HTTPAuthConfig) ConfirmOrDeclineAllowanceRequest(username string, token string, isConfirmed bool) error {
	return cfg.AuthStorage.ConfirmOrDeclineAllowanceRequest(username, token, isConfirmed)
}

func (cfg *HTTPAuthConfig) RevokeAllowance(username string, token string) error {
	zeroBigInt := lib.BigInt{}
	token = strings.ToLower(token)
	return cfg.AuthStorage.SetAllowance(username, token, zeroBigInt)
}

func (cfg *HTTPAuthConfig) DecreaseAllowance(username string, token string, amount lib.BigInt) error {
	agentUser, ok := cfg.AuthStorage.GetAgentUser(username)
	if !ok {
		return fmt.Errorf("allowance not found")
	}

	token = strings.ToLower(token)
	allowance, ok := agentUser.Allowances[token]
	if !ok {
		return fmt.Errorf("allowance not found")
	}

	allowance.Sub(&allowance.Int, &amount.Int)
	return cfg.AuthStorage.SetAllowance(username, token, allowance)
}

func (cfg *HTTPAuthConfig) GetAllAgentTx() ([]AgentTx, error) {
	txs, err := cfg.AuthStorage.GetAgentTxs()
	if err != nil {
		return nil, err
	}

	agentTxs := []AgentTx{}
	for txHash, username := range txs {
		agentTxs = append(agentTxs, AgentTx{TxHash: txHash, Username: username})
	}
	return agentTxs, nil
}
