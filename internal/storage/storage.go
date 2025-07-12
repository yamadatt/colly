package storage

import (
	"github.com/yourname/collycrawler/internal/models"
)

// Storage インターフェースは、記事データの保存方法を定義します
type Storage interface {
	// Save は単一の記事を保存します
	Save(article *models.Article) error
	
	// SaveBatch は複数の記事をバッチで保存します
	SaveBatch(articles []*models.Article) error
	
	// Load は保存された記事を読み込みます
	Load() ([]*models.Article, error)
	
	// Exists は指定されたURLまたはハッシュの記事が既に存在するかチェックします
	Exists(contentHash string) (bool, error)
	
	// GetStats は保存統計を取得します
	GetStats() (*StorageStats, error)
	
	// Close はストレージ接続を閉じます
	Close() error
}

// StorageStats はストレージの統計情報を表します
type StorageStats struct {
	TotalArticles    int    `json:"total_articles"`
	TotalSizeBytes   int64  `json:"total_size_bytes"`
	LastSavedAt      string `json:"last_saved_at"`
	StorageFormat    string `json:"storage_format"`
	OutputFile       string `json:"output_file"`
}

// StorageConfig はストレージの設定を表します
type StorageConfig struct {
	OutputFile      string
	BackupEnabled   bool
	BackupDirectory string
	MaxBackupFiles  int
	Format          string
}