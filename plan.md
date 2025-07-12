# 実装計画

以下は `aplicarion_desifn.md` の要件をもとに、Go + Colly でクローラを実装するための具体的なステップです。

## 1. 環境準備
1. Go 1.XX をインストール（最新安定版を想定）
2. `colly` をはじめ必要なパッケージを `go mod` で管理できるよう初期化する
3. `zap` などログ用ライブラリも同時に導入

## 2. プロジェクト構成作成
1. 以下のディレクトリを作成
   - `cmd/` - CLI エントリポイント
   - `internal/collector/` - Colly 設定とハンドラー
   - `internal/scraper/` - スクレイピングロジック
   - `internal/storage/` - データ保存処理
   - `internal/models/` - データ構造
   - `pkg/` - 共有ユーティリティ
   - `configs/` - YAML 設定ファイル置き場
   - `data/` - 収集データやログの保存先
   - `test/` - テスト用リソース
2. `go mod init github.com/yourname/collycrawler` を実行

## 3. 設定ファイル（YAML）設計
1. `configs/config.yaml` を作成し、以下の項目を定義
   - アプリ情報（名前、バージョン、ログレベル）
   - 対象サイト（ベースURL、開始URL、許可ドメイン、除外パターン）
   - クローラー設定（並行数、リクエスト間隔、タイムアウトなど）
   - HTML セレクター（記事抽出、リンク抽出）
   - 保存設定（出力ファイルパス、バックアップ設定）
2. Go の構造体にマッピングできるよう `internal/models/config.go` を実装
3. `viper` などのライブラリで YAML を読み込む処理を `pkg/config` に実装

## 4. Collector 実装
1. `internal/collector/collector.go` を作成
2. YAML 設定から Collector を初期化する処理を実装
   - User-Agent
   - Rate Limit (1 req/sec など)
   - `MaxDepth` や `AllowedDomains`
3. 共通のミドルウェア（ログ出力、エラーハンドリング）を組み込む

## 5. スクレイピングロジック
1. `internal/scraper/scraper.go` に記事ページの抽出処理を実装
   - `OnHTML` を用いてタイトル・本文・公開日等を取得
   - 本文 HTML からプレーンテキストへの変換処理を追加
2. ページネーションやリンク巡回のため `OnHTML("a[href]")` で内部リンクを収集
3. 重複 URL を避けるための判定ロジックを追加（ハッシュや既存レコード確認）

## 6. ストレージ層
1. `internal/storage/jsonl.go` を作成し、JSONL 形式で記事を保存する機能を実装
2. 将来的な拡張を考慮し、保存インターフェースを `internal/storage/storage.go` に定義
3. `data/` 配下に出力するように設定

## 7. CLI エントリポイント
1. `cmd/crawler/main.go` を用意
2. 起動時に YAML 設定を読み込み、Collector と Scraper を初期化して実行
3. 実行結果やエラーをログに出力（zap を利用）

## 8. テスト
1. `testdata/` などに HTML スナップショットを用意
2. 各コンポーネント（設定読み込み、スクレイピング処理、保存処理）のユニットテストを `go test` で記述
3. `github.com/stretchr/testify` などテスト用ライブラリを導入

## 9. 運用・改善フェーズ
1. エラーハンドリングを強化しリトライ回数や失敗 URL をログに残す
2. パフォーマンス計測を行い、必要に応じて並行処理数やバッチサイズを調整
3. ドキュメント整備（README 更新、使用方法記載）
4. Cron など定期実行を想定したラッパースクリプトを用意

## 10. 追加検討事項
- SQLite や S3 など他ストレージへの切り替えに備えてインターフェースを拡張
- 取得データのバリデーションやスキーマ管理
- 監視・アラートの仕組み（将来的には）

以上を順に進めることで、丁寧で拡張性の高いクローラを実装できます。
