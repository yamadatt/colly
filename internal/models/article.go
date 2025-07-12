package models

import "time"

// Article represents a scraped article with its metadata
type Article struct {
	URL           string    `json:"url"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	PlainText     string    `json:"plain_text"`
	Author        string    `json:"author,omitempty"`
	PublishedDate *time.Time `json:"published_date,omitempty"`
	ScrapedAt     time.Time `json:"scraped_at"`
	WordCount     int       `json:"word_count"`
	ContentHash   string    `json:"content_hash"`
}

// CrawlStats represents statistics about the crawling process
type CrawlStats struct {
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	Duration         string    `json:"duration"`
	TotalURLsVisited int       `json:"total_urls_visited"`
	ArticlesFound    int       `json:"articles_found"`
	ErrorsCount      int       `json:"errors_count"`
	SkippedCount     int       `json:"skipped_count"`
}