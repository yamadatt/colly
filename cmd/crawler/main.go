package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/yourname/collycrawler/internal/collector"
	"github.com/yourname/collycrawler/internal/scraper"
	"github.com/yourname/collycrawler/internal/storage"
	"github.com/yourname/collycrawler/pkg/config"
)

// アプリケーション情報
const (
	AppName    = "CollyCrawler"
	AppVersion = "1.0.0"
)

// コマンドライン引数
var (
	configPath = flag.String("config", "configs/config.yaml", "設定ファイルのパス")
	dryRun     = flag.Bool("dry-run", false, "実際の保存を行わずにテスト実行")
	verbose    = flag.Bool("verbose", false, "詳細ログを表示")
	version    = flag.Bool("version", false, "バージョン情報を表示")
	help       = flag.Bool("help", false, "ヘルプを表示")
)

func main() {
	flag.Parse()

	// バージョン情報表示
	if *version {
		fmt.Printf("%s v%s\n", AppName, AppVersion)
		os.Exit(0)
	}

	// ヘルプ表示
	if *help {
		printHelp()
		os.Exit(0)
	}

	// アプリケーション開始
	fmt.Printf("🚀 %s v%s を開始します\n", AppName, AppVersion)
	fmt.Printf("📄 設定ファイル: %s\n", *configPath)

	// 設定読み込み
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ 設定の読み込みに失敗: %v", err)
	}
	fmt.Printf("✅ 設定を読み込みました\n")

	// 詳細ログ設定
	if *verbose {
		cfg.App.LogLevel = "debug"
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// ストレージ設定検証
	if err := storage.ValidateStorageConfig(cfg); err != nil {
		log.Fatalf("❌ ストレージ設定エラー: %v", err)
	}

	// ストレージ初期化
	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("❌ ストレージの初期化に失敗: %v", err)
	}
	defer store.Close()
	fmt.Printf("✅ ストレージを初期化しました (%s)\n", cfg.Storage.OutputFormat)

	// スクレイパー初期化
	scraperInstance := scraper.NewScraper(cfg)
	fmt.Printf("✅ スクレイパーを初期化しました\n")

	// コレクター初期化
	c, err := collector.NewCollector(cfg)
	if err != nil {
		log.Fatalf("❌ コレクターの初期化に失敗: %v", err)
	}
	fmt.Printf("✅ コレクターを初期化しました\n")

	// 統計情報
	startTime := time.Now()
	var processedURLs int
	var savedArticles int
	var skippedArticles int

	// 記事コンテンツハンドラー設定
	c.SetupArticleHandler(func(e *colly.HTMLElement) {
		processedURLs++
		
		// 記事を抽出
		article := scraperInstance.ExtractArticle(e)
		if article == nil {
			skippedArticles++
			return
		}

		// ドライランモードでない場合のみ保存
		if !*dryRun {
			if err := store.Save(article); err != nil {
				log.Printf("❌ 記事保存エラー: %v", err)
				return
			}
		} else {
			fmt.Printf("🔍 [DRY-RUN] 記事を検出: %s\n", article.Title)
		}

		savedArticles++
		
		// 進捗表示
		if savedArticles%10 == 0 {
			fmt.Printf("📊 進捗: %d記事処理済み\n", savedArticles)
		}
	})

	// リンクハンドラー設定
	c.SetupLinkHandler(func(e *colly.HTMLElement) {
		links := scraperInstance.ExtractLinks(e)
		for _, link := range links {
			if c.IsAllowedURL(link) {
				e.Request.Visit(link)
			}
		}
	})

	// シグナルハンドリング（Ctrl+Cでの安全な終了）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Printf("\n⚠️  終了シグナルを受信しました。安全に終了中...\n")
		
		// 統計情報を表示
		printFinalStats(startTime, processedURLs, savedArticles, skippedArticles, store)
		
		// ストレージを閉じる
		store.Close()
		
		os.Exit(0)
	}()

	// クローリング開始
	fmt.Printf("\n🕷️  クローリングを開始します...\n")
	fmt.Printf("🎯 対象サイト: %s\n", cfg.Target.BaseURL)
	fmt.Printf("🔗 開始URL数: %d\n", len(cfg.Target.StartURLs))
	fmt.Printf("⚡ 並行数: %d\n", cfg.Crawler.ParallelJobs)
	fmt.Printf("⏱️  リクエスト間隔: %v\n", cfg.Crawler.RequestDelay)

	if *dryRun {
		fmt.Printf("🔍 ドライランモード: 実際の保存は行いません\n")
	}

	// クローリング実行
	if err := c.Start(); err != nil {
		log.Fatalf("❌ クローリング中にエラー: %v", err)
	}

	// 最終統計情報表示
	printFinalStats(startTime, processedURLs, savedArticles, skippedArticles, store)

	fmt.Printf("\n🎉 クローリングが完了しました！\n")
}

// printHelp はヘルプメッセージを表示します
func printHelp() {
	fmt.Printf("%s v%s - Webクローリング・スクレイピングツール\n\n", AppName, AppVersion)
	fmt.Println("使用方法:")
	fmt.Printf("  %s [オプション]\n\n", os.Args[0])
	fmt.Println("オプション:")
	fmt.Println("  -config string")
	fmt.Println("        設定ファイルのパス (デフォルト: configs/config.yaml)")
	fmt.Println("  -dry-run")
	fmt.Println("        実際の保存を行わずにテスト実行")
	fmt.Println("  -verbose")
	fmt.Println("        詳細ログを表示")
	fmt.Println("  -version")
	fmt.Println("        バージョン情報を表示")
	fmt.Println("  -help")
	fmt.Println("        このヘルプを表示")
	fmt.Println()
	fmt.Println("例:")
	fmt.Printf("  %s                              # デフォルト設定でクローリング実行\n", os.Args[0])
	fmt.Printf("  %s -config custom.yaml          # カスタム設定ファイルを使用\n", os.Args[0])
	fmt.Printf("  %s -dry-run -verbose            # ドライランモードで詳細ログ表示\n", os.Args[0])
	fmt.Println()
	fmt.Println("詳細情報:")
	fmt.Println("  設定ファイルはYAML形式で、クローリング対象やストレージ設定を定義します。")
	fmt.Println("  出力ファイルはJSONL形式で、1行につき1つの記事データが保存されます。")
}

// printFinalStats は最終統計情報を表示します
func printFinalStats(startTime time.Time, processedURLs, savedArticles, skippedArticles int, store storage.Storage) {
	duration := time.Since(startTime)
	
	fmt.Printf("\n📊 実行統計:\n")
	fmt.Printf("   実行時間: %v\n", duration)
	fmt.Printf("   処理URL数: %d\n", processedURLs)
	fmt.Printf("   保存記事数: %d\n", savedArticles)
	fmt.Printf("   スキップ記事数: %d\n", skippedArticles)
	
	if savedArticles > 0 {
		avgTime := duration / time.Duration(savedArticles)
		fmt.Printf("   平均処理時間: %v/記事\n", avgTime)
	}

	// ストレージ統計
	if stats, err := store.GetStats(); err == nil {
		fmt.Printf("\n💾 ストレージ統計:\n")
		fmt.Printf("   総記事数: %d\n", stats.TotalArticles)
		fmt.Printf("   ファイルサイズ: %d バイト\n", stats.TotalSizeBytes)
		fmt.Printf("   出力ファイル: %s\n", stats.OutputFile)
	}
}