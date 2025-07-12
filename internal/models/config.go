package models

import "time"

// Config represents the complete application configuration
type Config struct {
	App      AppConfig      `yaml:"app"`
	Target   TargetConfig   `yaml:"target"`
	Crawler  CrawlerConfig  `yaml:"crawler"`
	Selectors SelectorConfig `yaml:"selectors"`
	Storage  StorageConfig  `yaml:"storage"`
}

// AppConfig contains application-level settings
type AppConfig struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	LogLevel string `yaml:"log_level"`
}

// TargetConfig defines the target website configuration
type TargetConfig struct {
	BaseURL         string   `yaml:"base_url"`
	StartURLs       []string `yaml:"start_urls"`
	AllowedDomains  []string `yaml:"allowed_domains"`
	ExcludePatterns []string `yaml:"exclude_patterns"`
}

// CrawlerConfig contains crawler behavior settings
type CrawlerConfig struct {
	ParallelJobs     int           `yaml:"parallel_jobs"`
	RequestDelay     time.Duration `yaml:"request_delay"`
	Timeout          time.Duration `yaml:"timeout"`
	MaxDepth         int           `yaml:"max_depth"`
	UserAgent        string        `yaml:"user_agent"`
	RespectRobotsTxt bool          `yaml:"respect_robots_txt"`
}

// SelectorConfig defines HTML selectors for content extraction
type SelectorConfig struct {
	Article ArticleSelectors `yaml:"article"`
	Links   LinkSelectors    `yaml:"links"`
}

// ArticleSelectors contains selectors for extracting article content
type ArticleSelectors struct {
	Title         string `yaml:"title"`
	Content       string `yaml:"content"`
	PublishedDate string `yaml:"published_date"`
	Author        string `yaml:"author"`
}

// LinkSelectors contains selectors for finding links
type LinkSelectors struct {
	InternalLinks string `yaml:"internal_links"`
	Pagination    string `yaml:"pagination"`
}

// StorageConfig defines how and where to store collected data
type StorageConfig struct {
	OutputFormat     string `yaml:"output_format"`
	OutputFile       string `yaml:"output_file"`
	BackupEnabled    bool   `yaml:"backup_enabled"`
	BackupDirectory  string `yaml:"backup_directory"`
	MaxBackupFiles   int    `yaml:"max_backup_files"`
}