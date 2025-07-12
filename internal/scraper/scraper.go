package scraper

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/yourname/collycrawler/internal/models"
)

// Scraper handles the extraction of article content from HTML pages
type Scraper struct {
	config      *models.Config
	articles    []*models.Article
	visitedURLs map[string]bool
	urlFilter   *URLFilter
}

// NewScraper creates a new scraper instance
func NewScraper(config *models.Config) *Scraper {
	return &Scraper{
		config:      config,
		articles:    make([]*models.Article, 0),
		visitedURLs: make(map[string]bool),
		urlFilter:   NewURLFilter(),
	}
}

// ExtractArticle extracts article content from an HTML element
func (s *Scraper) ExtractArticle(e *colly.HTMLElement) *models.Article {
	// Check if we've already processed this URL
	urlStr := e.Request.URL.String()
	if s.visitedURLs[urlStr] {
		log.Printf("Skipping already visited URL: %s", urlStr)
		return nil
	}
	s.visitedURLs[urlStr] = true

	// 個別記事ページかどうかをチェック
	if !s.urlFilter.ShouldExtractContent(urlStr) {
		log.Printf("Skipping non-article page: %s (type: %s)", urlStr, s.urlFilter.GetURLType(urlStr))
		return nil
	}

	// Extract title
	title := s.extractTitle(e)
	if title == "" {
		log.Printf("No title found for %s, skipping", urlStr)
		return nil
	}

	// Extract content
	content := s.extractContent(e)
	if content == "" {
		log.Printf("No content found for %s, skipping", urlStr)
		return nil
	}

	// Extract metadata
	author := s.extractAuthor(e)
	publishedDate := s.extractPublishedDate(e)

	// Convert HTML to plain text
	plainText := s.htmlToPlainText(content)

	// Calculate word count
	wordCount := s.calculateWordCount(plainText)

	// Generate content hash for deduplication
	contentHash := s.generateContentHash(title + plainText)

	article := &models.Article{
		URL:           urlStr,
		Title:         strings.TrimSpace(title),
		Content:       content,
		PlainText:     plainText,
		Author:        author,
		PublishedDate: publishedDate,
		ScrapedAt:     time.Now(),
		WordCount:     wordCount,
		ContentHash:   contentHash,
	}

	s.articles = append(s.articles, article)
	log.Printf("Extracted article: %s (words: %d)", title, wordCount)

	return article
}

// extractTitle extracts the article title using configured selectors
func (s *Scraper) extractTitle(e *colly.HTMLElement) string {
	log.Printf("🔍 タイトル抽出開始: %s", e.Request.URL.String())
	
	selectors := strings.Split(s.config.Selectors.Article.Title, ",")
	
	for _, selector := range selectors {
		selector = strings.TrimSpace(selector)
		title := e.ChildText(selector)
		log.Printf("  セレクター [%s]: '%s'", selector, title)
		if title != "" {
			log.Printf("🎯 タイトル発見 [%s]: %s", selector, title)
			return s.cleanText(title)
		}
	}
	
	// Fallback to page title (titleタグから記事タイトル部分を抽出)
	pageTitle := s.cleanText(e.ChildText("title"))
	log.Printf("  ページタイトル: '%s'", pageTitle)
	if pageTitle != "" {
		// "タイトル | サイト名" の形式から記事タイトルを抽出
		if parts := strings.Split(pageTitle, "|"); len(parts) > 0 {
			articleTitle := strings.TrimSpace(parts[0])
			if articleTitle != "" {
				log.Printf("🎯 タイトル発見 [title fallback]: %s", articleTitle)
				return articleTitle
			}
		}
		log.Printf("🎯 タイトル発見 [title full]: %s", pageTitle)
		return pageTitle
	}
	
	log.Printf("❌ タイトルが見つかりません: %s", e.Request.URL.String())
	return ""
}

// extractContent extracts the article content using configured selectors
func (s *Scraper) extractContent(e *colly.HTMLElement) string {
	selectors := strings.Split(s.config.Selectors.Article.Content, ",")
	
	for _, selector := range selectors {
		selector = strings.TrimSpace(selector)
		
		// Get the HTML content
		var content string
		e.ForEach(selector, func(i int, el *colly.HTMLElement) {
			if content == "" { // Take the first match
				html, err := el.DOM.Html()
				if err == nil && html != "" {
					content = html
				}
			}
		})
		
		if content != "" {
			return s.cleanHTML(content)
		}
	}
	
	return ""
}

// extractAuthor extracts the article author using configured selectors
func (s *Scraper) extractAuthor(e *colly.HTMLElement) string {
	if s.config.Selectors.Article.Author == "" {
		return ""
	}
	
	selectors := strings.Split(s.config.Selectors.Article.Author, ",")
	
	for _, selector := range selectors {
		selector = strings.TrimSpace(selector)
		author := e.ChildText(selector)
		if author != "" {
			return s.cleanText(author)
		}
	}
	
	return ""
}

// extractPublishedDate extracts the published date using configured selectors
func (s *Scraper) extractPublishedDate(e *colly.HTMLElement) *time.Time {
	if s.config.Selectors.Article.PublishedDate == "" {
		return nil
	}
	
	selectors := strings.Split(s.config.Selectors.Article.PublishedDate, ",")
	
	for _, selector := range selectors {
		selector = strings.TrimSpace(selector)
		
		// Try to get datetime attribute first
		var dateStr string
		e.ForEach(selector, func(i int, el *colly.HTMLElement) {
			if dateStr == "" {
				// Check for datetime attribute
				if datetime := el.Attr("datetime"); datetime != "" {
					dateStr = datetime
				} else {
					// Fallback to text content
					dateStr = el.Text
				}
			}
		})
		
		if dateStr != "" {
			if parsedDate := s.parseDate(dateStr); parsedDate != nil {
				return parsedDate
			}
		}
	}
	
	return nil
}

// ExtractLinks extracts internal links for further crawling
func (s *Scraper) ExtractLinks(e *colly.HTMLElement) []string {
	var links []string
	linkMap := make(map[string]bool) // For deduplication
	
	// デバッグ用ログ
	log.Printf("🔍 リンク抽出開始: %s", e.Request.URL.String())
	
	// すべてのリンクを抽出（デバッグ用）
	e.ForEach("a[href]", func(i int, el *colly.HTMLElement) {
		href := el.Attr("href")
		if href == "" {
			return
		}
		
		// Resolve relative URLs
		absoluteURL := s.resolveURL(e.Request.URL, href)
		if absoluteURL == "" {
			return
		}
		
		// 記事URLパターンをチェック
		if strings.Contains(absoluteURL, "/posts/") && 
		   strings.HasSuffix(absoluteURL, "/") && 
		   !strings.HasSuffix(absoluteURL, "/posts/") &&
		   !strings.Contains(absoluteURL, "/page/") {
			
			if !linkMap[absoluteURL] {
				log.Printf("  ✅ 記事リンク発見: %s", absoluteURL)
				links = append(links, absoluteURL)
				linkMap[absoluteURL] = true
			}
		}
		
		// ページネーションリンクもチェック
		if strings.Contains(absoluteURL, "/page/") || 
		   strings.Contains(absoluteURL, "/posts/page/") {
			if !linkMap[absoluteURL] {
				log.Printf("  📄 ページネーションリンク発見: %s", absoluteURL)
				links = append(links, absoluteURL)
				linkMap[absoluteURL] = true
			}
		}
	})
	
	log.Printf("🔗 抽出されたリンク数: %d", len(links))
	return links
}

// GetArticles returns all extracted articles
func (s *Scraper) GetArticles() []*models.Article {
	return s.articles
}

// GetArticleCount returns the number of extracted articles
func (s *Scraper) GetArticleCount() int {
	return len(s.articles)
}

// Helper methods

// cleanText removes extra whitespace and normalizes text
func (s *Scraper) cleanText(text string) string {
	// Remove extra whitespace
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// cleanHTML removes unwanted HTML elements and attributes
func (s *Scraper) cleanHTML(html string) string {
	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html
	}
	
	// Remove script and style tags
	doc.Find("script, style, nav, header, footer, aside").Remove()
	
	// Remove comments
	doc.Find("*").Each(func(i int, sel *goquery.Selection) {
		sel.Contents().FilterFunction(func(i int, sel *goquery.Selection) bool {
			return sel.Get(0).Type == 8 // Comment node
		}).Remove()
	})
	
	// Get cleaned HTML
	cleanedHTML, err := doc.Html()
	if err != nil {
		return html
	}
	
	return cleanedHTML
}

// htmlToPlainText converts HTML content to plain text
func (s *Scraper) htmlToPlainText(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html
	}
	
	// Extract text content
	text := doc.Text()
	return s.cleanText(text)
}

// calculateWordCount counts words in text
func (s *Scraper) calculateWordCount(text string) int {
	words := strings.Fields(text)
	return len(words)
}

// generateContentHash creates a hash of the content for deduplication
func (s *Scraper) generateContentHash(content string) string {
	hash := md5.Sum([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// parseDate attempts to parse various date formats
func (s *Scraper) parseDate(dateStr string) *time.Time {
	// Common date formats to try
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"January 2, 2006",
		"Jan 2, 2006",
		"2006/01/02",
	}
	
	dateStr = strings.TrimSpace(dateStr)
	
	for _, format := range formats {
		if parsedTime, err := time.Parse(format, dateStr); err == nil {
			return &parsedTime
		}
	}
	
	log.Printf("Could not parse date: %s", dateStr)
	return nil
}

// resolveURL resolves relative URLs to absolute URLs
func (s *Scraper) resolveURL(base *url.URL, href string) string {
	parsedURL, err := url.Parse(href)
	if err != nil {
		return ""
	}
	
	resolvedURL := base.ResolveReference(parsedURL)
	return resolvedURL.String()
}

// isValidInternalLink checks if a URL is a valid internal link
func (s *Scraper) isValidInternalLink(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	
	// Check if it's in allowed domains
	for _, domain := range s.config.Target.AllowedDomains {
		if parsedURL.Host == domain {
			// Check against exclude patterns
			for _, pattern := range s.config.Target.ExcludePatterns {
				if s.matchesPattern(urlStr, pattern) {
					return false
				}
			}
			return true
		}
	}
	
	return false
}

// matchesPattern checks if a URL matches a glob pattern
func (s *Scraper) matchesPattern(url, pattern string) bool {
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