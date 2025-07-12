package scraper

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ContentProcessor handles advanced content processing and cleaning
type ContentProcessor struct {
	// Configuration for content processing
	removeElements []string
	preserveElements []string
}

// NewContentProcessor creates a new content processor
func NewContentProcessor() *ContentProcessor {
	return &ContentProcessor{
		removeElements: []string{
			"script", "style", "nav", "header", "footer", "aside",
			".advertisement", ".ads", ".social-share", ".comments",
			".sidebar", ".menu", ".navigation",
		},
		preserveElements: []string{
			"p", "h1", "h2", "h3", "h4", "h5", "h6",
			"ul", "ol", "li", "blockquote", "pre", "code",
			"strong", "em", "b", "i", "a",
		},
	}
}

// ProcessContent cleans and processes HTML content
func (cp *ContentProcessor) ProcessContent(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html
	}

	// Remove unwanted elements
	for _, selector := range cp.removeElements {
		doc.Find(selector).Remove()
	}

	// Clean attributes (keep only essential ones)
	doc.Find("*").Each(func(i int, sel *goquery.Selection) {
		// Keep only href for links and src for images
		if sel.Is("a") {
			href, exists := sel.Attr("href")
			sel.RemoveAttr("*")
			if exists {
				sel.SetAttr("href", href)
			}
		} else if sel.Is("img") {
			src, exists := sel.Attr("src")
			alt, altExists := sel.Attr("alt")
			sel.RemoveAttr("*")
			if exists {
				sel.SetAttr("src", src)
			}
			if altExists {
				sel.SetAttr("alt", alt)
			}
		} else {
			// Remove all attributes for other elements
			sel.RemoveAttr("*")
		}
	})

	// Get processed HTML
	processedHTML, err := doc.Html()
	if err != nil {
		return html
	}

	return processedHTML
}

// ExtractMainContent attempts to identify and extract the main content area
func (cp *ContentProcessor) ExtractMainContent(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html
	}

	// Try common main content selectors
	mainSelectors := []string{
		"main",
		"article",
		".main-content",
		".post-content",
		".article-content",
		".content",
		"#content",
		".entry-content",
		".post-body",
	}

	for _, selector := range mainSelectors {
		if content := doc.Find(selector).First(); content.Length() > 0 {
			if contentHTML, err := content.Html(); err == nil && contentHTML != "" {
				return contentHTML
			}
		}
	}

	// Fallback: try to find the largest text block
	return cp.findLargestTextBlock(doc)
}

// findLargestTextBlock finds the element with the most text content
func (cp *ContentProcessor) findLargestTextBlock(doc *goquery.Document) string {
	var largestElement *goquery.Selection
	maxTextLength := 0

	// Check common content containers
	contentSelectors := []string{"div", "section", "article"}

	for _, selector := range contentSelectors {
		doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
			text := strings.TrimSpace(sel.Text())
			if len(text) > maxTextLength {
				maxTextLength = len(text)
				largestElement = sel
			}
		})
	}

	if largestElement != nil {
		if html, err := largestElement.Html(); err == nil {
			return html
		}
	}

	// Ultimate fallback: return body content
	if bodyHTML, err := doc.Find("body").Html(); err == nil {
		return bodyHTML
	}

	return ""
}

// NormalizeWhitespace normalizes whitespace in text
func (cp *ContentProcessor) NormalizeWhitespace(text string) string {
	// Replace multiple whitespace characters with single space
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	
	// Remove leading/trailing whitespace
	text = strings.TrimSpace(text)
	
	return text
}

// RemoveEmptyParagraphs removes empty paragraph tags
func (cp *ContentProcessor) RemoveEmptyParagraphs(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html
	}

	// Remove empty paragraphs
	doc.Find("p").Each(func(i int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())
		if text == "" {
			sel.Remove()
		}
	})

	if processedHTML, err := doc.Html(); err == nil {
		return processedHTML
	}

	return html
}

// ExtractImages extracts image information from content
func (cp *ContentProcessor) ExtractImages(html string) []ImageInfo {
	var images []ImageInfo
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return images
	}

	doc.Find("img").Each(func(i int, sel *goquery.Selection) {
		src, _ := sel.Attr("src")
		alt, _ := sel.Attr("alt")
		title, _ := sel.Attr("title")

		if src != "" {
			images = append(images, ImageInfo{
				Src:   src,
				Alt:   alt,
				Title: title,
			})
		}
	})

	return images
}

// ImageInfo represents information about an image
type ImageInfo struct {
	Src   string `json:"src"`
	Alt   string `json:"alt"`
	Title string `json:"title"`
}

// ExtractLinks extracts link information from content
func (cp *ContentProcessor) ExtractLinks(html string) []LinkInfo {
	var links []LinkInfo
	
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return links
	}

	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		text := strings.TrimSpace(sel.Text())
		title, _ := sel.Attr("title")

		if href != "" {
			links = append(links, LinkInfo{
				Href:  href,
				Text:  text,
				Title: title,
			})
		}
	})

	return links
}

// LinkInfo represents information about a link
type LinkInfo struct {
	Href  string `json:"href"`
	Text  string `json:"text"`
	Title string `json:"title"`
}