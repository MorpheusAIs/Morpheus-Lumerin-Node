package config

import (
	"fmt"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/rating"
)

const (
	DefaultRatingConfigPath = "rating-config.json"
	RatingConfigDefault     = `
	{
		"$schema": "./internal/rating/rating-config-schema.json",
		"algorithm": "default",
		"providerAllowlist": [],
		"params": {
			"weights": {
				"tps": 0.24,
				"ttft": 0.08,
				"duration": 0.24,
				"success": 0.32,
				"stake": 0.12
			}
		}
	}
	`
)

func LoadRating(path string, content string, log lib.ILogger) (*rating.Rating, error) {
	log = log.Named("RATING_LOADER")

	filePath := DefaultRatingConfigPath
	if path != "" {
		filePath = path
	}

	var config string
	if content != "" {
		wrote, err := syncContentFile(filePath, content, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to write rating config content to file: %s", err)
		}
		if wrote {
			log.Infof("rating config written from RATING_CONFIG_CONTENT: %s", filePath)
		}
		config = content
	} else {
		loaded, err := lib.ReadJSONFile(filePath)
		if err != nil {
			log.Warnf("failed to load rating config file, using defaults")
			config = RatingConfigDefault
		} else {
			log.Infof("rating config loaded from file: %s", filePath)
			config = loaded
		}
	}

	return rating.NewRatingFromConfig([]byte(config), log)
}
