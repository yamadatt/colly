package config

import (
	"fmt"
	"os"

	"github.com/yourname/collycrawler/internal/models"
	"gopkg.in/yaml.v3"
)

// LoadConfig reads and parses the YAML configuration file
func LoadConfig(configPath string) (*models.Config, error) {
	// Read the configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse YAML into Config struct
	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Validate the configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig performs basic validation on the loaded configuration
func validateConfig(config *models.Config) error {
	// Validate app configuration
	if config.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}
	if config.App.Version == "" {
		return fmt.Errorf("app.version is required")
	}

	// Validate target configuration
	if config.Target.BaseURL == "" {
		return fmt.Errorf("target.base_url is required")
	}
	if len(config.Target.StartURLs) == 0 {
		return fmt.Errorf("target.start_urls must contain at least one URL")
	}
	if len(config.Target.AllowedDomains) == 0 {
		return fmt.Errorf("target.allowed_domains must contain at least one domain")
	}

	// Validate crawler configuration
	if config.Crawler.ParallelJobs <= 0 {
		return fmt.Errorf("crawler.parallel_jobs must be greater than 0")
	}
	if config.Crawler.MaxDepth < 0 {
		return fmt.Errorf("crawler.max_depth must be non-negative")
	}
	if config.Crawler.UserAgent == "" {
		return fmt.Errorf("crawler.user_agent is required")
	}

	// Validate selectors
	if config.Selectors.Article.Title == "" {
		return fmt.Errorf("selectors.article.title is required")
	}
	if config.Selectors.Article.Content == "" {
		return fmt.Errorf("selectors.article.content is required")
	}

	// Validate storage configuration
	if config.Storage.OutputFile == "" {
		return fmt.Errorf("storage.output_file is required")
	}

	return nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	return "configs/config.yaml"
}