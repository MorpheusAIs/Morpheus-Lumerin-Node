package aiengine

type LocalModel struct {
	Id             string
	Name           string
	Model          string
	ApiType        string
	ApiUrl         string
	Slots          int
	CapacityPolicy string
}

type LocalAgent struct {
	Id              string
	Name            string
	Command         string
	Args            []string
	ConcurrentSlots int
	CapacityPolicy  string
}

type AgentTool struct {
	Name        string        `json:"name"`
	Description string `json:"description"`
	InputSchema ToolInputSchema `json:"inputSchema"`

}

type ToolInputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}
