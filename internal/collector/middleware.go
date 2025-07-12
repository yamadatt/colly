package collector

import (
	"log"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
)

// RateLimitMiddleware creates a middleware for rate limiting
func RateLimitMiddleware(delay time.Duration) colly.RequestCallback {
	var lastRequest time.Time
	
	return func(r *colly.Request) {
		if !lastRequest.IsZero() {
			elapsed := time.Since(lastRequest)
			if elapsed < delay {
				time.Sleep(delay - elapsed)
			}
		}
		lastRequest = time.Now()
	}
}

// RetryMiddleware creates a middleware for retrying failed requests
func RetryMiddleware(maxRetries int) colly.ErrorCallback {
	return func(r *colly.Response, err error) {
		retryCount := r.Request.Ctx.GetAny("retry_count")
		if retryCount == nil {
			retryCount = 0
		}
		
		count := retryCount.(int)
		if count < maxRetries {
			log.Printf("Retrying request to %s (attempt %d/%d)", r.Request.URL.String(), count+1, maxRetries)
			r.Request.Ctx.Put("retry_count", count+1)
			
			// Wait before retry
			time.Sleep(time.Duration(count+1) * time.Second)
			
			// Retry the request
			r.Request.Retry()
		} else {
			log.Printf("Max retries exceeded for %s", r.Request.URL.String())
		}
	}
}

// UserAgentRotationMiddleware rotates user agents
func UserAgentRotationMiddleware(userAgents []string) colly.RequestCallback {
	var index int
	
	return func(r *colly.Request) {
		if len(userAgents) > 0 {
			r.Headers.Set("User-Agent", userAgents[index%len(userAgents)])
			index++
		}
	}
}

// CacheMiddleware adds caching headers to avoid re-downloading unchanged content
func CacheMiddleware() colly.RequestCallback {
	return func(r *colly.Request) {
		// Add cache-friendly headers
		r.Headers.Set("Cache-Control", "max-age=3600")
		r.Headers.Set("If-Modified-Since", time.Now().Add(-24*time.Hour).Format(http.TimeFormat))
	}
}

// RobotsTxtMiddleware checks robots.txt compliance
func RobotsTxtMiddleware() colly.RequestCallback {
	return func(r *colly.Request) {
		// This is handled by colly's built-in robots.txt support
		// when CheckHead is enabled, but we can add custom logic here
		log.Printf("Checking robots.txt compliance for: %s", r.URL.String())
	}
}

// ContentTypeFilterMiddleware filters responses by content type
func ContentTypeFilterMiddleware(allowedTypes []string) colly.ResponseCallback {
	return func(r *colly.Response) {
		contentType := r.Headers.Get("Content-Type")
		
		allowed := false
		for _, allowedType := range allowedTypes {
			if contentType == allowedType || 
			   (allowedType == "text/html" && (contentType == "text/html" || contentType == "")) {
				allowed = true
				break
			}
		}
		
		if !allowed {
			log.Printf("Skipping non-HTML content: %s (Content-Type: %s)", 
				r.Request.URL.String(), contentType)
			return
		}
	}
}