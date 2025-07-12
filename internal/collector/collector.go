package collector

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/yourname/collycrawler/internal/models"
)

// Collector wraps colly.Collector with our configuration
type Collector struct {
	*colly.Collector
	config *models.Config
	stats  *models.CrawlStats
}

// NewCollector creates a new configured Colly collector
func NewCollector(config *models.Config) (*Collector, error) {
	// Create base colly collector
	c := colly.NewCollector(
		colly.UserAgent(config.Crawler.UserAgent),
		colly.Async(true),
	)

	// Set allowed domains
	c.AllowedDomains = config.Target.AllowedDomains

	// Set max depth if specified
	if config.Crawler.MaxDepth > 0 {
		c.MaxDepth = config.Crawler.MaxDepth
	}

	// Configure rate limiting
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: config.Crawler.ParallelJobs,
		Delay:       config.Crawler.RequestDelay,
	})

	// Set timeout
	c.SetRequestTimeout(config.Crawler.Timeout)

	// Respect robots.txt if configured
	if config.Crawler.RespectRobotsTxt {
		c.CheckHead = true
	}

	// Add URL filters for excluded patterns
	for _, pattern := range config.Target.ExcludePatterns {
		regexPattern := convertGlobToRegex(pattern)
		if compiledRegex, err := regexp.Compile(regexPattern); err == nil {
			c.DisallowedURLFilters = append(c.DisallowedURLFilters, compiledRegex)
		}
	}

	// Initialize stats
	stats := &models.CrawlStats{
		StartTime: time.Now(),
	}

	collector := &Collector{
		Collector: c,
		config:    config,
		stats:     stats,
	}

	// Set up middleware
	collector.setupMiddleware()

	return collector, nil
}

// setupMiddleware configures common middleware for logging and error handling
func (c *Collector) setupMiddleware() {
	// Request logging middleware
	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting: %s", r.URL.String())
		c.stats.TotalURLsVisited++
	})

	// Response logging middleware
	c.OnResponse(func(r *colly.Response) {
		log.Printf("Response %d: %s", r.StatusCode, r.Request.URL.String())
	})

	// Error handling middleware
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error visiting %s: %v", r.Request.URL.String(), err)
		c.stats.ErrorsCount++
	})

	// HTML validation middleware
	c.OnHTML("html", func(e *colly.HTMLElement) {
		// Basic HTML validation - ensure we have a proper HTML document
		if e.DOM.Find("head").Length() == 0 && e.DOM.Find("body").Length() == 0 {
			log.Printf("Warning: Invalid HTML structure at %s", e.Request.URL.String())
		}
	})

	// Add debug logging if log level is debug
	if strings.ToLower(c.config.App.LogLevel) == "debug" {
		c.Collector.OnRequest(func(r *colly.Request) {
			log.Printf("DEBUG: Request - %s %s", r.Method, r.URL.String())
		})
		c.Collector.OnResponse(func(r *colly.Response) {
			log.Printf("DEBUG: Response - %d %s", r.StatusCode, r.Request.URL.String())
		})
	}
}

// Start begins the crawling process with the configured start URLs
func (c *Collector) Start() error {
	log.Printf("Starting crawler for %s", c.config.App.Name)
	log.Printf("Target domains: %v", c.config.Target.AllowedDomains)
	log.Printf("Parallel jobs: %d", c.config.Crawler.ParallelJobs)
	log.Printf("Request delay: %v", c.config.Crawler.RequestDelay)

	// Visit all start URLs
	for _, startURL := range c.config.Target.StartURLs {
		log.Printf("Adding start URL: %s", startURL)
		c.Visit(startURL)
	}

	// Start the async collector
	c.Wait()

	// Update end time and duration
	c.stats.EndTime = time.Now()
	c.stats.Duration = c.stats.EndTime.Sub(c.stats.StartTime).String()

	log.Printf("Crawling completed in %s", c.stats.Duration)
	return nil
}

// GetStats returns the current crawling statistics
func (c *Collector) GetStats() *models.CrawlStats {
	return c.stats
}

// GetConfig returns the collector's configuration
func (c *Collector) GetConfig() *models.Config {
	return c.config
}

// SetupArticleHandler sets up the HTML handler for extracting articles
func (c *Collector) SetupArticleHandler(handler func(*colly.HTMLElement)) {
	// すべてのページでコンテンツ抽出を試行するため、bodyセレクターを使用
	c.OnHTML("body", handler)
}

// SetupLinkHandler sets up the HTML handler for following internal links
func (c *Collector) SetupLinkHandler(handler func(*colly.HTMLElement)) {
	// すべてのページでリンクを抽出するため、bodyセレクターを使用
	c.OnHTML("body", handler)
}

// IsAllowedURL checks if a URL should be crawled based on configuration
func (c *Collector) IsAllowedURL(url string) bool {
	// Check against excluded patterns
	for _, pattern := range c.config.Target.ExcludePatterns {
		if matchesPattern(url, pattern) {
			return false
		}
	}

	// Check if URL is in allowed domains
	for _, domain := range c.config.Target.AllowedDomains {
		if strings.Contains(url, domain) {
			return true
		}
	}

	return false
}

// convertGlobToRegex converts a glob pattern to a regex pattern
func convertGlobToRegex(glob string) string {
	// Simple glob to regex conversion
	// * becomes .*
	// ? becomes .
	regex := strings.ReplaceAll(glob, "*", ".*")
	regex = strings.ReplaceAll(regex, "?", ".")
	return regex
}

// matchesPattern checks if a URL matches a glob pattern
func matchesPattern(url, pattern string) bool {
	// Simple pattern matching for common cases
	if strings.Contains(pattern, "*") {
		// Handle wildcard patterns
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(url, parts[0]) && strings.HasSuffix(url, parts[1])
		}
	}
	return strings.Contains(url, pattern)
}