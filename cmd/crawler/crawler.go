package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/yourname/collycrawler/internal/collector"
	"github.com/yourname/collycrawler/internal/models"
	"github.com/yourname/collycrawler/internal/scraper"
	"github.com/yourname/collycrawler/internal/storage"
)

// CrawlerApp はクローラーアプリケーションのメイン構造体です
type CrawlerApp struct {
	config    *models.Config
	collector *collector.Collector
	scraper   *scraper.Scraper
	storage   storage.Storage
	stats     *CrawlStats
}

// CrawlStats はクローリングの統計情報を保持します
type CrawlStats struct {
	StartTime       time.Time
	EndTime         time.Time
	ProcessedURLs   int
	SavedArticles   int
	SkippedArticles int
	ErrorCount      int
	DryRun          bool
}

// NewCrawlerApp は新しいクローラーアプリケーションを作成します
func NewCrawlerApp(config *models.Config, dryRun bool) (*CrawlerApp, error) {
	// ストレージ初期化
	store, err := storage.NewStorage(config)
	if err != nil {
		return nil, fmt.Errorf("ストレージ初期化エラー: %w", err)
	}

	// スクレイパー初期化
	scraperInstance := scraper.NewScraper(config)

	// コレクター初期化
	c, err := collector.NewCollector(config)
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("コレクター初期化エラー: %w", err)
	}

	app := &CrawlerApp{
		config:    config,
		collector: c,
		scraper:   scraperInstance,
		storage:   store,
		stats: &CrawlStats{
			StartTime: time.Now(),
			DryRun:    dryRun,
		},
	}

	// ハンドラー設定
	app.setupHandlers()

	return app, nil
}

// setupHandlers はコレクターのハンドラーを設定します
func (app *CrawlerApp) setupHandlers() {
	// 記事コンテンツハンドラー
	app.collector.SetupArticleHandler(func(e *colly.HTMLElement) {
		app.handleArticle(e)
	})

	// リンクハンドラー
	app.collector.SetupLinkHandler(func(e *colly.HTMLElement) {
		app.handleLinks(e)
	})

	// エラーハンドラー
	app.collector.OnError(func(r *colly.Response, err error) {
		app.stats.ErrorCount++
		log.Printf("❌ エラー [%s]: %v", r.Request.URL.String(), err)
	})

	// リクエストハンドラー（進捗表示用）
	app.collector.OnRequest(func(r *colly.Request) {
		if app.stats.ProcessedURLs%50 == 0 && app.stats.ProcessedURLs > 0 {
			fmt.Printf("🔄 処理中: %d URL訪問済み\n", app.stats.ProcessedURLs)
		}
	})
}

// handleArticle は記事の処理を行います
func (app *CrawlerApp) handleArticle(e *colly.HTMLElement) {
	app.stats.ProcessedURLs++

	// 記事を抽出
	article := app.scraper.ExtractArticle(e)
	if article == nil {
		app.stats.SkippedArticles++
		return
	}

	// 重複チェック
	exists, err := app.storage.Exists(article.ContentHash)
	if err != nil {
		log.Printf("❌ 重複チェックエラー: %v", err)
		app.stats.ErrorCount++
		return
	}

	if exists {
		log.Printf("⏭️  重複記事をスキップ: %s", article.Title)
		app.stats.SkippedArticles++
		return
	}

	// ドライランモードでない場合のみ保存
	if !app.stats.DryRun {
		if err := app.storage.Save(article); err != nil {
			log.Printf("❌ 記事保存エラー: %v", err)
			app.stats.ErrorCount++
			return
		}
	} else {
		fmt.Printf("🔍 [DRY-RUN] 記事検出: %s (文字数: %d)\n", article.Title, article.WordCount)
	}

	app.stats.SavedArticles++

	// 進捗表示
	if app.stats.SavedArticles%5 == 0 {
		fmt.Printf("📝 進捗: %d記事処理済み\n", app.stats.SavedArticles)
	}
}

// handleLinks はリンクの処理を行います
func (app *CrawlerApp) handleLinks(e *colly.HTMLElement) {
	links := app.scraper.ExtractLinks(e)
	
	for _, link := range links {
		if app.collector.IsAllowedURL(link) {
			// 訪問済みURLのチェックは Colly が自動で行う
			e.Request.Visit(link)
		}
	}
}

// Run はクローリングを実行します
func (app *CrawlerApp) Run() error {
	fmt.Printf("\n🕷️  クローリング開始\n")
	fmt.Printf("🎯 対象: %s\n", app.config.Target.BaseURL)
	fmt.Printf("🔗 開始URL: %d件\n", len(app.config.Target.StartURLs))
	
	if app.stats.DryRun {
		fmt.Printf("🔍 ドライランモード\n")
	}

	// クローリング実行
	err := app.collector.Start()
	
	app.stats.EndTime = time.Now()
	
	return err
}

// GetStats は統計情報を返します
func (app *CrawlerApp) GetStats() *CrawlStats {
	return app.stats
}

// Close はリソースを解放します
func (app *CrawlerApp) Close() error {
	if app.storage != nil {
		return app.storage.Close()
	}
	return nil
}

// PrintStats は統計情報を表示します
func (app *CrawlerApp) PrintStats() {
	duration := app.stats.EndTime.Sub(app.stats.StartTime)
	
	fmt.Printf("\n📊 クローリング統計:\n")
	fmt.Printf("   実行時間: %v\n", duration)
	fmt.Printf("   処理URL数: %d\n", app.stats.ProcessedURLs)
	fmt.Printf("   保存記事数: %d\n", app.stats.SavedArticles)
	fmt.Printf("   スキップ記事数: %d\n", app.stats.SkippedArticles)
	fmt.Printf("   エラー数: %d\n", app.stats.ErrorCount)
	
	if app.stats.SavedArticles > 0 {
		avgTime := duration / time.Duration(app.stats.SavedArticles)
		fmt.Printf("   平均処理時間: %v/記事\n", avgTime)
	}

	// ストレージ統計
	if stats, err := app.storage.GetStats(); err == nil {
		fmt.Printf("\n💾 ストレージ統計:\n")
		fmt.Printf("   総記事数: %d\n", stats.TotalArticles)
		fmt.Printf("   ファイルサイズ: %d バイト\n", stats.TotalSizeBytes)
		fmt.Printf("   出力ファイル: %s\n", stats.OutputFile)
	}
}