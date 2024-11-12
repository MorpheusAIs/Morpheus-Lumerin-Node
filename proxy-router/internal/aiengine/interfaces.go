package aiengine

type ProdiaGenerationResult struct {
	Job      string `json:"job"`
	Status   string `json:"status"`
	ImageUrl string `json:"imageUrl" binding:"omitempty"`
}

type ProdiaGenerationRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	ApiUrl string `json:"apiUrl"`
	ApiKey string `json:"apiKey"`
}

type ProdiaImageGenerationCallback func(completion *ProdiaGenerationResult) error

type LocalModel struct {
	Id      string
	Name    string
	Model   string
	ApiType string
}
