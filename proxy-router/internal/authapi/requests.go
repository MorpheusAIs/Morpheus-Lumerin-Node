package authapi

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
