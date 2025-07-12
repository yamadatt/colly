package scraper

import (
	"regexp"
)

// URLFilter は個別記事ページかどうかを判定するフィルター
type URLFilter struct {
	articlePatterns []string
	excludePatterns []string
}

// NewURLFilter は新しいURLフィルターを作成
func NewURLFilter() *URLFilter {
	return &URLFilter{
		articlePatterns: []string{
			// 個別記事ページのパターン（より柔軟に）
			`/posts/[^/]+/$`,                    // /posts/article-name/
			`/posts/\d+/[^/]+/$`,               // /posts/2023/article-name/
			`/posts/[^/]+/[^/]+/$`,             // /posts/category/article-name/
			`/posts/[^/]+-[^/]+/$`,             // /posts/article-name-with-dashes/
			`/posts/[^/]+_[^/]+/$`,             // /posts/article_name_with_underscores/
		},
		excludePatterns: []string{
			// 除外するページのパターン
			`/posts/$`,                 // 記事一覧ページ
			`/posts/index\.xml$`,       // RSS フィード
			`/tags/`,                   // タグページ
			`/categories/`,             // カテゴリページ
			`\.(jpg|jpeg|png|gif|pdf|css|js)$`, // 静的ファイル
			// ページネーションは除外しない（記事リンクを含むため）
		},
	}
}

// IsArticlePage は個別記事ページかどうかを判定
func (uf *URLFilter) IsArticlePage(url string) bool {
	// 除外パターンをチェック
	for _, pattern := range uf.excludePatterns {
		if matched, _ := regexp.MatchString(pattern, url); matched {
			return false
		}
	}
	
	// 記事パターンをチェック
	for _, pattern := range uf.articlePatterns {
		if matched, _ := regexp.MatchString(pattern, url); matched {
			return true
		}
	}
	
	return false
}

// IsListPage は記事一覧ページかどうかを判定
func (uf *URLFilter) IsListPage(url string) bool {
	listPatterns := []string{
		`/posts/$`,
		`/posts/page/\d+/$`,
		`/$`,  // トップページ
	}
	
	for _, pattern := range listPatterns {
		if matched, _ := regexp.MatchString(pattern, url); matched {
			return true
		}
	}
	
	return false
}

// ShouldExtractContent はコンテンツを抽出すべきかを判定
func (uf *URLFilter) ShouldExtractContent(url string) bool {
	// 個別記事ページのみコンテンツを抽出
	return uf.IsArticlePage(url)
}

// ShouldFollowLinks はリンクを辿るべきかを判定
func (uf *URLFilter) ShouldFollowLinks(url string) bool {
	// 記事一覧ページと個別記事ページの両方でリンクを辿る
	return uf.IsListPage(url) || uf.IsArticlePage(url)
}

// GetURLType はURLの種類を返す
func (uf *URLFilter) GetURLType(url string) string {
	if uf.IsArticlePage(url) {
		return "article"
	} else if uf.IsListPage(url) {
		return "list"
	} else {
		return "other"
	}
}