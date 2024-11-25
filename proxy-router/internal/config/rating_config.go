package config

import (
	"fmt"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/rating"
)

const (
	DefaultRatingConfigPath = "rating-config.json"
)

func LoadScorer(path string, log lib.ILogger) (*rating.Rating, error) {
	filePath := DefaultRatingConfigPath
	if path != "" {
		filePath = path
	}

	modelsConfig, err := lib.ReadJSONFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read models config file: %s", err)
	}

	log.Infof("rating config loaded from file: %s", filePath)

	return rating.NewRatingFromConfig([]byte(modelsConfig))
}
