package authapi

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

type AddUserReq struct {
	Username string   `json:"username" validate:"required"`
	Password string   `json:"password" validate:"required"`
	Perms    []string `json:"perms" validate:"required"`
}

type RemoveUserReq struct {
	Username string `json:"username" validate:"required"`
}

type AuthRes struct {
	Result bool `json:"result"`
}

type RequestAgentUserReq struct {
	Username   string            `json:"username" validate:"required"`
	Password   string            `json:"password" validate:"required"`
	Perms      []string          `json:"perms" validate:"required"`
	Allowances map[string]string `json:"allowances" validate:"required"`
}

type RequestAllowanceReq struct {
	Username  string     `json:"username" validate:"required"`
	Token     string     `json:"token" validate:"required"`
	Allowance lib.BigInt `json:"allowance" swaggertype:"string"`
}

type ConfirmAgentReq struct {
	Username string `json:"username" validate:"required"`
	Confirm  bool   `json:"confirm" validate:"required"`
}

type RevokeAllowanceReq struct {
	Username string `json:"username" validate:"required"`
	Token    string `json:"token" validate:"required"`
}

type ConfirmAllowanceReq struct {
	Username string `json:"username" validate:"required"`
	Token    string `json:"token" validate:"required"`
	Confirm  bool   `json:"confirm" validate:"required"`
}

type AllowanceRequest struct {
	Username  string     `json:"username"`
	Token     string     `json:"token"`
	Allowance lib.BigInt `json:"allowance" swaggertype:"string"`
}

type AllowanceRequestsRes struct {
	Requests []AllowanceRequest `json:"requests"`
}

type AgentUsersRes struct {
	Agents []*AgentUser `json:"agents"`
}

type AgentUser struct {
	Username    string                `json:"username"`
	Perms       []string              `json:"perms"`
	IsConfirmed bool                  `json:"isConfirmed"`
	Allowances  map[string]lib.BigInt `json:"allowances"`
}

type AgentTxsRes struct {
	TxHashes   []string `json:"txHashes"`
	NextCursor []byte   `json:"nextCursor"`
}

type AgentTxReqURI struct {
	Username string `json:"username" uri:"username" binding:"required" validate:"required"`
}

type CursorQuery struct {
	Cursor []byte `json:"cursor" form:"cursor"`
	Limit  uint   `json:"limit" form:"limit" binding:"required" validate:"required,min=1,max=100"`
}
