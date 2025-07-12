package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

func main() {
	c := colly.NewCollector(
		colly.UserAgent("CollyCrawler/1.0 Selector Test"),
	)

	// 個別記事ページのセレクターをテスト
	testURL := "https://yamada-tech-memo.netlify.app/posts/git-clean-untracked-files/"
	
	c.OnHTML("html", func(e *colly.HTMLElement) {
		fmt.Printf("🔍 ページ構造分析: %s\n", e.Request.URL.String())
		
		// タイトルセレクターのテスト
		fmt.Println("\n📝 タイトルセレクターテスト:")
		titleSelectors := []string{
			"h1",
			".post-title", 
			".entry-title",
			"article h1",
			"main h1",
			"title",
		}
		
		for _, selector := range titleSelectors {
			title := strings.TrimSpace(e.ChildText(selector))
			if title != "" {
				fmt.Printf("  ✅ %s: %s\n", selector, title)
			} else {
				fmt.Printf("  ❌ %s: (見つからない)\n", selector)
			}
		}
		
		// コンテンツセレクターのテスト
		fmt.Println("\n📄 コンテンツセレクターテスト:")
		contentSelectors := []string{
			"main",
			"article", 
			".post-content",
			".entry-content",
			".content",
			".post-body",
			".markdown",
			"body",
		}
		
		for _, selector := range contentSelectors {
			var contentLength int
			e.ForEach(selector, func(i int, el *colly.HTMLElement) {
				if i == 0 { // 最初のマッチのみ
					text := strings.TrimSpace(el.Text)
					contentLength = len(text)
				}
			})
			
			if contentLength > 0 {
				fmt.Printf("  ✅ %s: %d文字\n", selector, contentLength)
			} else {
				fmt.Printf("  ❌ %s: (見つからない)\n", selector)
			}
		}
		
		// 日付セレクターのテスト
		fmt.Println("\n📅 日付セレクターテスト:")
		dateSelectors := []string{
			"time[datetime]",
			".post-date",
			".published", 
			".date",
			".post-meta time",
			".meta time",
		}
		
		for _, selector := range dateSelectors {
			var dateText string
			e.ForEach(selector, func(i int, el *colly.HTMLElement) {
				if i == 0 {
					if datetime := el.Attr("datetime"); datetime != "" {
						dateText = datetime
					} else {
						dateText = strings.TrimSpace(el.Text)
					}
				}
			})
			
			if dateText != "" {
				fmt.Printf("  ✅ %s: %s\n", selector, dateText)
			} else {
				fmt.Printf("  ❌ %s: (見つからない)\n", selector)
			}
		}
		
		// ページの基本構造を確認
		fmt.Println("\n🏗️ ページ構造:")
		fmt.Printf("  <main>要素: %d個\n", e.DOM.Find("main").Length())
		fmt.Printf("  <article>要素: %d個\n", e.DOM.Find("article").Length())
		fmt.Printf("  <h1>要素: %d個\n", e.DOM.Find("h1").Length())
		fmt.Printf("  <time>要素: %d個\n", e.DOM.Find("time").Length())
		
		// クラス名を調査
		fmt.Println("\n🎨 主要なクラス名:")
		e.ForEach("[class]", func(i int, el *colly.HTMLElement) {
			if i < 10 { // 最初の10個のみ
				class := el.Attr("class")
				if class != "" {
					fmt.Printf("  %s\n", class)
				}
			}
		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("🌐 訪問中: %s\n", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("❌ エラー: %v\n", err)
	})

	fmt.Println("🔍 個別記事ページのセレクターをテスト中...")
	c.Visit(testURL)
}