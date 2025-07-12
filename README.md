# CollyCrawler

特定サイトの記事をクローリング・スクレイピングしてデータ化するGoアプリケーション

## 概要

CollyCrawlerは、[yamada-tech-memo.netlify.app](https://yamada-tech-memo.netlify.app/)から記事を収集し、JSONL形式で保存するWebクローラーです。

## 特徴

- 🕷️ **Collyフレームワーク**: 高性能なGoベースのWebスクレイピング
- 📄 **JSONL出力**: ストリーミング処理に適した形式
- ⚙️ **YAML設定**: 柔軟で読みやすい設定ファイル
- 🤝 **丁寧なクローリング**: サイトに配慮したレート制限とrobotstxt対応
- 🔄 **重複排除**: コンテンツハッシュによる自動重複検出
- 💾 **バックアップ機能**: データ損失を防ぐ自動バックアップ
- 🔍 **ドライランモード**: 実際の保存前のテスト実行

## 技術スタック

- **言語**: Go 1.23+
- **主要ライブラリ**: 
  - [Colly v2](https://go-colly.org/) - Webスクレイピング
  - [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML設定
  - [PuerkitoBio/goquery](https://github.com/PuerkitoBio/goquery) - HTML解析

## インストール

```bash
# リポジトリをクローン
git clone https://github.com/yourname/collycrawler
cd collycrawler

# 依存関係をインストール
go mod tidy

# ビルド
go build -o crawler cmd/crawler/*.go
```

## 使用方法

### 基本的な使用方法

```bash
# デフォルト設定でクローリング実行
./crawler

# ドライランモード（実際の保存は行わない）
./crawler -dry-run

# 詳細ログ表示
./crawler -verbose

# カスタム設定ファイルを使用
./crawler -config custom-config.yaml
```

### コマンドラインオプション

| オプション | 説明 |
|-----------|------|
| `-config` | 設定ファイルのパス (デフォルト: configs/config.yaml) |
| `-dry-run` | 実際の保存を行わずにテスト実行 |
| `-verbose` | 詳細ログを表示 |
| `-version` | バージョン情報を表示 |
| `-help` | ヘルプを表示 |

## 設定ファイル

`configs/config.yaml`で動作をカスタマイズできます：

```yaml
# アプリケーション設定
app:
  name: "collycrawler"
  version: "1.0.0"
  log_level: "info"

# 対象サイト設定
target:
  base_url: "https://yamada-tech-memo.netlify.app"
  start_urls:
    - "https://yamada-tech-memo.netlify.app/posts/"
  allowed_domains:
    - "yamada-tech-memo.netlify.app"

# クローラー設定
crawler:
  parallel_jobs: 2
  request_delay: "1s"
  timeout: "30s"
  max_depth: 10

# ストレージ設定
storage:
  output_format: "jsonl"
  output_file: "data/articles.jsonl"
  backup_enabled: true
```

## 出力形式

記事は以下のJSONL形式で保存されます：

```json
{"url":"https://example.com/article","title":"記事タイトル","content":"<p>記事内容</p>","plain_text":"記事内容","author":"著者名","published_date":"2024-01-15T10:00:00Z","scraped_at":"2024-01-15T11:00:00Z","word_count":150,"content_hash":"abc123"}
```

## プロジェクト構造

```
├── cmd/crawler/          # CLIエントリポイント
├── internal/
│   ├── collector/        # Colly設定とハンドラー
│   ├── scraper/         # スクレイピングロジック
│   ├── storage/         # データ保存処理
│   └── models/          # データ構造
├── pkg/config/          # 設定読み込み
├── configs/             # YAML設定ファイル
├── data/               # 収集データとバックアップ
└── test/               # テスト用リソース
```

## 開発

### テスト実行

```bash
go test ./...
```

### ローカル開発

```bash
# ドライランでテスト
go run cmd/crawler/*.go -dry-run -verbose

# 設定ファイルの検証
go run cmd/crawler/*.go -config configs/config.yaml -dry-run
```

## ライセンス

MIT License

## 注意事項

- robots.txtを尊重し、適切なクローリング間隔を設定してください
- 対象サイトの利用規約を確認してください
- 大量のリクエストを送信する前に、サイト管理者に連絡することを推奨します

## 貢献

プルリクエストやイシューの報告を歓迎します。

## 作者

[あなたの名前](https://github.com/yourname)