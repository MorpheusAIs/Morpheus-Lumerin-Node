package config

import (
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

func LoadRating(path string, log lib.ILogger) (*rating.Rating, error) {
	log = log.Named("RATING_LOADER")

	filePath := DefaultRatingConfigPath
	if path != "" {
		filePath = path
	}

	config, err := lib.ReadJSONFile(filePath)
	if err != nil {
		log.Warnf("failed to load rating config file, using defaults")
		config = RatingConfigDefault
	} else {
		log.Infof("rating config loaded from file: %s", filePath)
	}

	return rating.NewRatingFromConfig([]byte(config), log)
}
