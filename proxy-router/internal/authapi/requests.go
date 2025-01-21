package authapi

import "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"

type AddUserReq struct {
	Username string   `json:"username",validate:"required"`
	Password string   `json:"password",validate:"required"`
	Perms    []string `json:"perms",validate:"required"`
}

type RemoveUserReq struct {
	Username string `json:"username",validate:"required"`
}

type AuthRes struct {
	Result bool `json:"result"`
}

type RequestAgentUserReq struct {
	Username   string                `json:"username",validate:"required"`
	Password   string                `json:"password",validate:"required"`
	Perms      []string              `json:"perms",validate:"required"`
	Allowances map[string]lib.BigInt `json:"allowances",validate:"required"`
}

type RequestAllowanceReq struct {
	Username  string     `json:"username",validate:"required"`
	Token     string     `json:"token",validate:"required"`
	Allowance lib.BigInt `json:"allowance"`
}

type ConfirmAgentReq struct {
	Username string `json:"username",validate:"required"`
	Confirm  bool   `json:"confirm",validate:"required"`
}

type RevokeAllowanceReq struct {
	Username string `json:"username",validate:"required"`
	Token    string `json:"token",validate:"required"`
}

type ConfirmAllowanceReq struct {
	Username string `json:"username" validate:"required"`
	Token    string `json:"token" validate:"required"`
	Confirm  bool   `json:"confirm" validate:"required"`
}

type AllowanceRequest struct {
	Username  string     `json:"username"`
	Token     string     `json:"token"`
	Allowance lib.BigInt `json:"allowance"`
}

type AllowanceRequestsRes struct {
	Requests []AllowanceRequest `json:"requests"`
}

type AgentUsersReq struct {
	Username    string                `json:"username"`
	Perms       []string              `json:"perms"`
	IsConfirmed bool                  `json:"is_confirmed"`
	Allowances  map[string]lib.BigInt `json:"allowances"`
}

type AgentUsersRes struct {
	Agents []string `json:"agents"`
}

type AgentTx struct {
	TxHash string `json:"tx_hash"`
	Username string `json:"username"`
}

type AgentTxsRes struct {
	Txs []AgentTx `json:"txs"`
}
