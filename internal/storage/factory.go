package storage

import (
	"fmt"
	"strings"

	"github.com/yourname/collycrawler/internal/models"
)

// NewStorage は設定に基づいて適切なストレージインスタンスを作成します
func NewStorage(config *models.Config) (Storage, error) {
	format := strings.ToLower(config.Storage.OutputFormat)
	
	switch format {
	case "jsonl":
		return NewJSONLStorage(config)
	case "json":
		// 将来的にJSON形式をサポートする場合
		return nil, fmt.Errorf("JSON形式は現在サポートされていません。JSONLを使用してください")
	case "csv":
		// 将来的にCSV形式をサポートする場合
		return nil, fmt.Errorf("CSV形式は現在サポートされていません。JSONLを使用してください")
	default:
		return nil, fmt.Errorf("サポートされていないストレージ形式: %s", format)
	}
}

// ValidateStorageConfig はストレージ設定を検証します
func ValidateStorageConfig(config *models.Config) error {
	if config.Storage.OutputFile == "" {
		return fmt.Errorf("storage.output_file は必須です")
	}

	format := strings.ToLower(config.Storage.OutputFormat)
	supportedFormats := []string{"jsonl"}
	
	isSupported := false
	for _, supported := range supportedFormats {
		if format == supported {
			isSupported = true
			break
		}
	}
	
	if !isSupported {
		return fmt.Errorf("サポートされていないストレージ形式: %s (サポート形式: %v)", format, supportedFormats)
	}

	if config.Storage.BackupEnabled {
		if config.Storage.BackupDirectory == "" {
			return fmt.Errorf("バックアップが有効な場合、storage.backup_directory は必須です")
		}
		if config.Storage.MaxBackupFiles <= 0 {
			return fmt.Errorf("storage.max_backup_files は1以上である必要があります")
		}
	}

	return nil
}