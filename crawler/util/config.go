package util

import (
	"os"

	"github.com/goccy/go-yaml"
)

/**
 * Config represents the full YAML configuration file.
 */
type Config struct {
	SearchEngine SearchEngineConfig `yaml:"search-engine"`
}

/**
 * SearchEngineConfig groups all search engine settings.
 */
type SearchEngineConfig struct {
	Version  string         `yaml:"version"`
	Crawler  CrawlerConfig  `yaml:"crawler"`
	Database DatabaseConfig `yaml:"database"`
}

/**
 * CrawlerConfig contains crawler runtime parameters.
 */
type CrawlerConfig struct {
	StartURL   string `yaml:"start-url"`
	CustomURL  string `yaml:"custom-url"`
	CrawlLimit int    `yaml:"crawl-limit"`
	MinWorkers int    `yaml:"min-workers"`
	MaxWorkers int    `yaml:"max-workers"`
}

/**
 * DatabaseConfig contains database configuration.
 */
type DatabaseConfig struct {
	StoragePath string `yaml:"storage-path"`
}

/**
 * LoadConfig loads and parses a YAML configuration file.
 *
 * @param path path to config.yml
 * @return parsed Config struct
 */
func LoadConfig(path string) (*Config, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}