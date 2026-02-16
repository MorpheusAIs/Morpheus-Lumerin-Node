package system

type FD struct {
	ID   string
	Path string
}

type SetEthNodeURLReq struct {
	URLs []string `json:"urls" binding:"required" validate:"required,url"`
}

type ConfigResponse struct {
	Version       string
	Commit        string
	DerivedConfig interface{}
	Config        interface{}
}

type HealthCheckResponse struct {
	Status     string            `json:"status"`
	Version    string            `json:"version"`
	Uptime     string            `json:"uptime"`
	Components map[string]string `json:"components,omitempty"`
}

type StatusRes struct {
	Status string `json:"status"`
}

func OkRes() StatusRes {
	return StatusRes{Status: "ok"}
}
