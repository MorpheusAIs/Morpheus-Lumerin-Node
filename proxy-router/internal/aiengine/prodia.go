package aiengine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	c "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

const HEADER_PRODIA_KEY = "X-Prodia-Key"
const PRODIA_DEFAULT_BASE_URL = "https://api.prodia.com/v1"

type ProdiaGenerationResult struct {
	Job      string `json:"job"`
	Status   string `json:"status"`
	ImageUrl string `json:"imageUrl" binding:"omitempty"`
}

func waitJobResult(ctx context.Context, apiURL, apiKey, jobID string) (*ProdiaGenerationResult, error) {
	url := fmt.Sprintf("%s/job/%s", apiURL, jobID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		err = lib.WrapError(ErrJobCheckRequest, err)
		return nil, err
	}

	req.Header.Add(c.HEADER_ACCEPT, c.CONTENT_TYPE_JSON)
	req.Header.Add(HEADER_PRODIA_KEY, apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = lib.WrapError(ErrJobCheckRequest, err)
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		err = lib.WrapError(ErrJobCheckRequest, err)
		return nil, err
	}

	var result ProdiaGenerationResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = lib.WrapError(ErrJobCheckRequest, err)
		return nil, err
	}

	if result.Status == "succeeded" {
		return &result, nil
	}

	if result.Status == "failed" {
		return nil, ErrJobFailed
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(1 * time.Second):
	}

	return waitJobResult(ctx, apiURL, apiKey, jobID)
}
