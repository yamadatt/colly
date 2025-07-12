package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yourname/collycrawler/internal/models"
)

// JSONLStorage はJSONL形式でのストレージ実装です
type JSONLStorage struct {
	config       *StorageConfig
	outputFile   string
	existingHashes map[string]bool
}

// NewJSONLStorage は新しいJSONLストレージインスタンスを作成します
func NewJSONLStorage(config *models.Config) (*JSONLStorage, error) {
	storageConfig := &StorageConfig{
		OutputFile:      config.Storage.OutputFile,
		BackupEnabled:   config.Storage.BackupEnabled,
		BackupDirectory: config.Storage.BackupDirectory,
		MaxBackupFiles:  config.Storage.MaxBackupFiles,
		Format:          config.Storage.OutputFormat,
	}

	// 出力ディレクトリを作成
	outputDir := filepath.Dir(storageConfig.OutputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("出力ディレクトリの作成に失敗しました: %w", err)
	}

	// バックアップディレクトリを作成（有効な場合）
	if storageConfig.BackupEnabled {
		if err := os.MkdirAll(storageConfig.BackupDirectory, 0755); err != nil {
			return nil, fmt.Errorf("バックアップディレクトリの作成に失敗しました: %w", err)
		}
	}

	storage := &JSONLStorage{
		config:         storageConfig,
		outputFile:     storageConfig.OutputFile,
		existingHashes: make(map[string]bool),
	}

	// 既存のハッシュを読み込み
	if err := storage.loadExistingHashes(); err != nil {
		log.Printf("既存ハッシュの読み込み中に警告: %v", err)
	}

	return storage, nil
}

// Save は単一の記事をJSONL形式で保存します
func (j *JSONLStorage) Save(article *models.Article) error {
	// 重複チェック
	if j.existingHashes[article.ContentHash] {
		log.Printf("重複記事をスキップ: %s (ハッシュ: %s)", article.Title, article.ContentHash)
		return nil
	}

	// バックアップ作成（有効な場合）
	if j.config.BackupEnabled {
		if err := j.createBackup(); err != nil {
			log.Printf("バックアップ作成中に警告: %v", err)
		}
	}

	// ファイルを追記モードで開く
	file, err := os.OpenFile(j.outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("出力ファイルのオープンに失敗: %w", err)
	}
	defer file.Close()

	// JSONエンコード
	jsonData, err := json.Marshal(article)
	if err != nil {
		return fmt.Errorf("記事のJSONエンコードに失敗: %w", err)
	}

	// JSONL形式で書き込み（各行に1つのJSONオブジェクト）
	if _, err := file.Write(jsonData); err != nil {
		return fmt.Errorf("ファイルへの書き込みに失敗: %w", err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		return fmt.Errorf("改行の書き込みに失敗: %w", err)
	}

	// ハッシュを記録
	j.existingHashes[article.ContentHash] = true

	log.Printf("記事を保存しました: %s", article.Title)
	return nil
}

// SaveBatch は複数の記事をバッチで保存します
func (j *JSONLStorage) SaveBatch(articles []*models.Article) error {
	if len(articles) == 0 {
		return nil
	}

	// バックアップ作成（有効な場合）
	if j.config.BackupEnabled {
		if err := j.createBackup(); err != nil {
			log.Printf("バックアップ作成中に警告: %v", err)
		}
	}

	// ファイルを追記モードで開く
	file, err := os.OpenFile(j.outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("出力ファイルのオープンに失敗: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	savedCount := 0
	skippedCount := 0

	for _, article := range articles {
		// 重複チェック
		if j.existingHashes[article.ContentHash] {
			skippedCount++
			continue
		}

		// JSONエンコード
		jsonData, err := json.Marshal(article)
		if err != nil {
			log.Printf("記事のJSONエンコードに失敗: %s - %v", article.Title, err)
			continue
		}

		// JSONL形式で書き込み
		if _, err := writer.Write(jsonData); err != nil {
			log.Printf("記事の書き込みに失敗: %s - %v", article.Title, err)
			continue
		}
		if _, err := writer.WriteString("\n"); err != nil {
			log.Printf("改行の書き込みに失敗: %s - %v", article.Title, err)
			continue
		}

		// ハッシュを記録
		j.existingHashes[article.ContentHash] = true
		savedCount++
	}

	log.Printf("バッチ保存完了: %d件保存、%d件スキップ", savedCount, skippedCount)
	return nil
}

// Load は保存された記事を読み込みます
func (j *JSONLStorage) Load() ([]*models.Article, error) {
	file, err := os.Open(j.outputFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.Article{}, nil // ファイルが存在しない場合は空のスライスを返す
		}
		return nil, fmt.Errorf("ファイルのオープンに失敗: %w", err)
	}
	defer file.Close()

	var articles []*models.Article
	scanner := bufio.NewScanner(file)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // 空行をスキップ
		}

		var article models.Article
		if err := json.Unmarshal([]byte(line), &article); err != nil {
			log.Printf("行 %d のJSONパースに失敗: %v", lineNumber, err)
			continue
		}

		articles = append(articles, &article)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ファイル読み込み中にエラー: %w", err)
	}

	log.Printf("%d件の記事を読み込みました", len(articles))
	return articles, nil
}

// Exists は指定されたハッシュの記事が既に存在するかチェックします
func (j *JSONLStorage) Exists(contentHash string) (bool, error) {
	return j.existingHashes[contentHash], nil
}

// GetStats はストレージの統計情報を取得します
func (j *JSONLStorage) GetStats() (*StorageStats, error) {
	stats := &StorageStats{
		StorageFormat: "jsonl",
		OutputFile:    j.outputFile,
	}

	// ファイル情報を取得
	if fileInfo, err := os.Stat(j.outputFile); err == nil {
		stats.TotalSizeBytes = fileInfo.Size()
		stats.LastSavedAt = fileInfo.ModTime().Format(time.RFC3339)
	}

	// 記事数をカウント
	articles, err := j.Load()
	if err != nil {
		return stats, err
	}
	stats.TotalArticles = len(articles)

	return stats, nil
}

// Close はストレージ接続を閉じます（JSONLの場合は何もしない）
func (j *JSONLStorage) Close() error {
	log.Printf("JSONLストレージを閉じました。総ハッシュ数: %d", len(j.existingHashes))
	return nil
}

// loadExistingHashes は既存ファイルからハッシュを読み込みます
func (j *JSONLStorage) loadExistingHashes() error {
	articles, err := j.Load()
	if err != nil {
		return err
	}

	for _, article := range articles {
		j.existingHashes[article.ContentHash] = true
	}

	log.Printf("既存ハッシュを読み込みました: %d件", len(j.existingHashes))
	return nil
}

// createBackup は現在のファイルのバックアップを作成します
func (j *JSONLStorage) createBackup() error {
	// 元ファイルが存在しない場合はバックアップ不要
	if _, err := os.Stat(j.outputFile); os.IsNotExist(err) {
		return nil
	}

	// バックアップファイル名を生成
	timestamp := time.Now().Format("20060102_150405")
	backupFileName := fmt.Sprintf("articles_backup_%s.jsonl", timestamp)
	backupPath := filepath.Join(j.config.BackupDirectory, backupFileName)

	// ファイルをコピー
	if err := j.copyFile(j.outputFile, backupPath); err != nil {
		return fmt.Errorf("バックアップファイルの作成に失敗: %w", err)
	}

	// 古いバックアップファイルを削除
	if err := j.cleanupOldBackups(); err != nil {
		log.Printf("古いバックアップの削除中に警告: %v", err)
	}

	log.Printf("バックアップを作成しました: %s", backupPath)
	return nil
}

// copyFile はファイルをコピーします
func (j *JSONLStorage) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// cleanupOldBackups は古いバックアップファイルを削除します
func (j *JSONLStorage) cleanupOldBackups() error {
	files, err := filepath.Glob(filepath.Join(j.config.BackupDirectory, "articles_backup_*.jsonl"))
	if err != nil {
		return err
	}

	if len(files) <= j.config.MaxBackupFiles {
		return nil // 削除不要
	}

	// ファイルを時刻順にソート
	sort.Strings(files)

	// 古いファイルを削除
	filesToDelete := files[:len(files)-j.config.MaxBackupFiles]
	for _, file := range filesToDelete {
		if err := os.Remove(file); err != nil {
			log.Printf("バックアップファイルの削除に失敗: %s - %v", file, err)
		} else {
			log.Printf("古いバックアップを削除: %s", file)
		}
	}

	return nil
}