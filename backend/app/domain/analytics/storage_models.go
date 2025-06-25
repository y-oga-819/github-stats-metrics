package analytics

import (
	"time"

	prDomain "github-stats-metrics/domain/pull_request"
)

// PRMetricsStorage はPRメトリクスの永続化モデル
type PRMetricsStorage struct {
	// 基本識別情報
	ID        string    `json:"id" db:"id"`                 // ユニークID
	PRID      string    `json:"prId" db:"pr_id"`            // Pull Request ID
	PRNumber  int       `json:"prNumber" db:"pr_number"`    // Pull Request番号
	Title     string    `json:"title" db:"title"`           // PRタイトル
	Author    string    `json:"author" db:"author"`         // 作成者
	Repository string   `json:"repository" db:"repository"` // リポジトリ名
	
	// 日時情報
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`     // PR作成日時
	MergedAt    *time.Time `json:"mergedAt" db:"merged_at"`       // マージ日時
	CollectedAt time.Time  `json:"collectedAt" db:"collected_at"` // データ収集日時
	
	// サイズメトリクス（JSON格納）
	SizeMetricsJSON string `json:"sizeMetricsJson" db:"size_metrics_json"`
	
	// 時間メトリクス（個別カラム + JSON）
	TotalCycleTimeSeconds    *int64  `json:"totalCycleTimeSeconds" db:"total_cycle_time_seconds"`
	TimeToFirstReviewSeconds *int64  `json:"timeToFirstReviewSeconds" db:"time_to_first_review_seconds"`
	TimeToApprovalSeconds    *int64  `json:"timeToApprovalSeconds" db:"time_to_approval_seconds"`
	TimeToMergeSeconds       *int64  `json:"timeToMergeSeconds" db:"time_to_merge_seconds"`
	TimeMetricsJSON          string  `json:"timeMetricsJson" db:"time_metrics_json"`
	
	// 品質メトリクス（個別カラム + JSON）
	ReviewCommentCount   int     `json:"reviewCommentCount" db:"review_comment_count"`
	ReviewRoundCount     int     `json:"reviewRoundCount" db:"review_round_count"`
	ReviewerCount        int     `json:"reviewerCount" db:"reviewer_count"`
	FirstReviewPassRate  float64 `json:"firstReviewPassRate" db:"first_review_pass_rate"`
	QualityMetricsJSON   string  `json:"qualityMetricsJson" db:"quality_metrics_json"`
	
	// 複雑度情報
	ComplexityScore float64                    `json:"complexityScore" db:"complexity_score"`
	SizeCategory    prDomain.PRSizeCategory    `json:"sizeCategory" db:"size_category"`
	
	// インデックス用フィールド
	YearMonth  string `json:"yearMonth" db:"year_month"`   // YYYY-MM形式
	WeekOfYear string `json:"weekOfYear" db:"week_of_year"` // YYYY-WW形式
	DayOfYear  string `json:"dayOfYear" db:"day_of_year"`   // YYYY-DDD形式
}

// PRMetricsStorageSchema はデータベーススキーマ定義
type PRMetricsStorageSchema struct {
	TableName string
	Indexes   []IndexDefinition
}

// IndexDefinition はインデックス定義
type IndexDefinition struct {
	Name    string
	Columns []string
	Unique  bool
	Type    IndexType
}

// IndexType はインデックスタイプ
type IndexType string

const (
	IndexTypeBTree IndexType = "btree"
	IndexTypeHash  IndexType = "hash"
	IndexTypeGIN   IndexType = "gin"
	IndexTypeGIST  IndexType = "gist"
)

// GetPRMetricsSchema はPRメトリクスのスキーマ定義を返す
func GetPRMetricsSchema() PRMetricsStorageSchema {
	return PRMetricsStorageSchema{
		TableName: "pr_metrics",
		Indexes: []IndexDefinition{
			// 主キー
			{
				Name:    "pk_pr_metrics",
				Columns: []string{"id"},
				Unique:  true,
				Type:    IndexTypeBTree,
			},
			// PR識別用
			{
				Name:    "idx_pr_metrics_pr_id",
				Columns: []string{"pr_id"},
				Unique:  true,
				Type:    IndexTypeBTree,
			},
			// 日時検索用
			{
				Name:    "idx_pr_metrics_created_at",
				Columns: []string{"created_at"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			{
				Name:    "idx_pr_metrics_merged_at",
				Columns: []string{"merged_at"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 集計検索用
			{
				Name:    "idx_pr_metrics_author",
				Columns: []string{"author"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			{
				Name:    "idx_pr_metrics_repository",
				Columns: []string{"repository"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 期間検索用（複合インデックス）
			{
				Name:    "idx_pr_metrics_author_period",
				Columns: []string{"author", "year_month"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			{
				Name:    "idx_pr_metrics_repo_period",
				Columns: []string{"repository", "year_month"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// メトリクス検索用
			{
				Name:    "idx_pr_metrics_complexity",
				Columns: []string{"complexity_score"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			{
				Name:    "idx_pr_metrics_size_category",
				Columns: []string{"size_category"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 日時範囲検索用
			{
				Name:    "idx_pr_metrics_date_range",
				Columns: []string{"created_at", "merged_at"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
		},
	}
}

// FileChangeStorage はファイル変更情報の永続化モデル
type FileChangeStorage struct {
	ID           string  `json:"id" db:"id"`                     // ユニークID
	PRMetricsID  string  `json:"prMetricsId" db:"pr_metrics_id"` // 親PRメトリクスID
	FileName     string  `json:"fileName" db:"file_name"`        // ファイル名
	FileType     string  `json:"fileType" db:"file_type"`        // ファイルタイプ
	LinesAdded   int     `json:"linesAdded" db:"lines_added"`    // 追加行数
	LinesDeleted int     `json:"linesDeleted" db:"lines_deleted"` // 削除行数
	IsNewFile    bool    `json:"isNewFile" db:"is_new_file"`     // 新規ファイルか
	IsDeleted    bool    `json:"isDeleted" db:"is_deleted"`      // 削除ファイルか
	IsRenamed    bool    `json:"isRenamed" db:"is_renamed"`      // リネームファイルか
	CollectedAt  time.Time `json:"collectedAt" db:"collected_at"` // データ収集日時
}

// GetFileChangeSchema はファイル変更のスキーマ定義を返す
func GetFileChangeSchema() PRMetricsStorageSchema {
	return PRMetricsStorageSchema{
		TableName: "file_changes",
		Indexes: []IndexDefinition{
			// 主キー
			{
				Name:    "pk_file_changes",
				Columns: []string{"id"},
				Unique:  true,
				Type:    IndexTypeBTree,
			},
			// 外部キー
			{
				Name:    "idx_file_changes_pr_metrics_id",
				Columns: []string{"pr_metrics_id"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// ファイル検索用
			{
				Name:    "idx_file_changes_file_type",
				Columns: []string{"file_type"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			{
				Name:    "idx_file_changes_file_name",
				Columns: []string{"file_name"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
		},
	}
}

// ReviewEventStorage はレビューイベント情報の永続化モデル
type ReviewEventStorage struct {
	ID          string                      `json:"id" db:"id"`                     // ユニークID
	PRMetricsID string                      `json:"prMetricsId" db:"pr_metrics_id"` // 親PRメトリクスID
	EventType   prDomain.ReviewEventType    `json:"eventType" db:"event_type"`      // イベントタイプ
	CreatedAt   time.Time                   `json:"createdAt" db:"created_at"`      // イベント発生日時
	Actor       string                      `json:"actor" db:"actor"`               // 実行者
	Reviewer    *string                     `json:"reviewer" db:"reviewer"`         // レビュアー（該当する場合）
	CollectedAt time.Time                   `json:"collectedAt" db:"collected_at"`  // データ収集日時
}

// GetReviewEventSchema はレビューイベントのスキーマ定義を返す
func GetReviewEventSchema() PRMetricsStorageSchema {
	return PRMetricsStorageSchema{
		TableName: "review_events",
		Indexes: []IndexDefinition{
			// 主キー
			{
				Name:    "pk_review_events",
				Columns: []string{"id"},
				Unique:  true,
				Type:    IndexTypeBTree,
			},
			// 外部キー
			{
				Name:    "idx_review_events_pr_metrics_id",
				Columns: []string{"pr_metrics_id"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// イベント検索用
			{
				Name:    "idx_review_events_type_created",
				Columns: []string{"event_type", "created_at"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// レビュアー検索用
			{
				Name:    "idx_review_events_reviewer",
				Columns: []string{"reviewer"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
		},
	}
}

// DataRetentionPolicy はデータ保持ポリシー
type DataRetentionPolicy struct {
	// 詳細データの保持期間
	DetailedDataRetentionDays int `json:"detailedDataRetentionDays"` // デフォルト: 90日
	
	// 集計データの保持期間
	DailyAggregationRetentionDays   int `json:"dailyAggregationRetentionDays"`   // デフォルト: 365日
	WeeklyAggregationRetentionDays  int `json:"weeklyAggregationRetentionDays"`  // デフォルト: 730日（2年）
	MonthlyAggregationRetentionDays int `json:"monthlyAggregationRetentionDays"` // デフォルト: 1095日（3年）
	
	// アーカイブ設定
	EnableArchiving    bool   `json:"enableArchiving"`    // アーカイブ機能有効化
	ArchiveStoragePath string `json:"archiveStoragePath"` // アーカイブ保存先
	
	// 圧縮設定
	EnableCompression     bool `json:"enableCompression"`     // 圧縮機能有効化
	CompressionThresholdDays int `json:"compressionThresholdDays"` // 圧縮開始日数（デフォルト: 30日）
}

// GetDefaultRetentionPolicy はデフォルトの保持ポリシーを返す
func GetDefaultRetentionPolicy() DataRetentionPolicy {
	return DataRetentionPolicy{
		DetailedDataRetentionDays:       90,
		DailyAggregationRetentionDays:   365,
		WeeklyAggregationRetentionDays:  730,
		MonthlyAggregationRetentionDays: 1095,
		EnableArchiving:                 false,
		ArchiveStoragePath:              "/data/archive",
		EnableCompression:               true,
		CompressionThresholdDays:        30,
	}
}

// PartitionStrategy はパーティション戦略
type PartitionStrategy struct {
	// パーティション方式
	PartitionType PartitionType `json:"partitionType"` // 時間、ハッシュなど
	
	// 時間パーティション設定
	TimePartitionInterval string `json:"timePartitionInterval"` // "monthly", "weekly", "daily"
	TimePartitionColumn   string `json:"timePartitionColumn"`   // パーティション対象カラム
	
	// ハッシュパーティション設定
	HashPartitionColumn string `json:"hashPartitionColumn"` // ハッシュ対象カラム
	HashPartitionCount  int    `json:"hashPartitionCount"`  // パーティション数
	
	// パーティション管理設定
	AutoCreatePartitions bool `json:"autoCreatePartitions"` // 自動パーティション作成
	PartitionMaintenanceEnabled bool `json:"partitionMaintenanceEnabled"` // パーティション保守有効化
}

// PartitionType はパーティションタイプ
type PartitionType string

const (
	PartitionTypeTime PartitionType = "time"
	PartitionTypeHash PartitionType = "hash"
	PartitionTypeRange PartitionType = "range"
	PartitionTypeList PartitionType = "list"
)

// GetRecommendedPartitionStrategy は推奨パーティション戦略を返す
func GetRecommendedPartitionStrategy() PartitionStrategy {
	return PartitionStrategy{
		PartitionType:                "time",
		TimePartitionInterval:        "monthly",
		TimePartitionColumn:          "created_at",
		HashPartitionColumn:          "",
		HashPartitionCount:           0,
		AutoCreatePartitions:         true,
		PartitionMaintenanceEnabled:  true,
	}
}

// StorageConfiguration はストレージ設定
type StorageConfiguration struct {
	// データベース設定
	DatabaseType     string `json:"databaseType"`     // "postgresql", "mysql", "sqlite"
	ConnectionString string `json:"connectionString"` // 接続文字列
	MaxConnections   int    `json:"maxConnections"`   // 最大接続数
	
	// パフォーマンス設定
	BatchInsertSize    int  `json:"batchInsertSize"`    // バッチ挿入サイズ
	EnableQueryCache   bool `json:"enableQueryCache"`   // クエリキャッシュ有効化
	QueryTimeoutSeconds int  `json:"queryTimeoutSeconds"` // クエリタイムアウト
	
	// バックアップ設定
	EnableBackup       bool   `json:"enableBackup"`       // バックアップ有効化
	BackupInterval     string `json:"backupInterval"`     // バックアップ間隔
	BackupStoragePath  string `json:"backupStoragePath"`  // バックアップ保存先
	BackupRetentionDays int   `json:"backupRetentionDays"` // バックアップ保持日数
	
	// メンテナンス設定
	AutoVacuumEnabled bool `json:"autoVacuumEnabled"` // 自動VACUUM有効化
	AutoAnalyzeEnabled bool `json:"autoAnalyzeEnabled"` // 自動ANALYZE有効化
}

// GetDefaultStorageConfiguration はデフォルトのストレージ設定を返す
func GetDefaultStorageConfiguration() StorageConfiguration {
	return StorageConfiguration{
		DatabaseType:        "postgresql",
		ConnectionString:    "postgres://user:password@localhost/github_metrics?sslmode=disable",
		MaxConnections:      25,
		BatchInsertSize:     100,
		EnableQueryCache:    true,
		QueryTimeoutSeconds: 30,
		EnableBackup:        true,
		BackupInterval:      "daily",
		BackupStoragePath:   "/data/backups",
		BackupRetentionDays: 30,
		AutoVacuumEnabled:   true,
		AutoAnalyzeEnabled:  true,
	}
}

// AggregatedMetricsStorage は集計済みメトリクスの永続化モデル
type AggregatedMetricsStorage struct {
	// 基本識別情報
	ID               string    `json:"id" db:"id"`                             // ユニークID
	AggregationLevel string    `json:"aggregationLevel" db:"aggregation_level"` // "team", "developer", "repository"
	AggregationPeriod string   `json:"aggregationPeriod" db:"aggregation_period"` // "daily", "weekly", "monthly"
	
	// 集計対象の識別
	TargetID   string    `json:"targetId" db:"target_id"`     // 対象ID（開発者名、リポジトリ名など）
	TargetName string    `json:"targetName" db:"target_name"` // 対象名（表示用）
	
	// 期間情報
	PeriodStart time.Time `json:"periodStart" db:"period_start"` // 集計期間開始
	PeriodEnd   time.Time `json:"periodEnd" db:"period_end"`     // 集計期間終了
	
	// 基本メトリクス
	TotalPRs            int     `json:"totalPRs" db:"total_prs"`                       // 総PR数
	MergedPRs           int     `json:"mergedPRs" db:"merged_prs"`                     // マージされたPR数
	ClosedPRs           int     `json:"closedPRs" db:"closed_prs"`                     // クローズされたPR数
	
	// サイクルタイム統計（秒単位）
	AvgCycleTimeSeconds    *int64  `json:"avgCycleTimeSeconds" db:"avg_cycle_time_seconds"`
	MedianCycleTimeSeconds *int64  `json:"medianCycleTimeSeconds" db:"median_cycle_time_seconds"`
	P95CycleTimeSeconds    *int64  `json:"p95CycleTimeSeconds" db:"p95_cycle_time_seconds"`
	
	// レビュー時間統計（秒単位）
	AvgReviewTimeSeconds    *int64  `json:"avgReviewTimeSeconds" db:"avg_review_time_seconds"`
	MedianReviewTimeSeconds *int64  `json:"medianReviewTimeSeconds" db:"median_review_time_seconds"`
	P95ReviewTimeSeconds    *int64  `json:"p95ReviewTimeSeconds" db:"p95_review_time_seconds"`
	
	// 承認時間統計（秒単位）
	AvgApprovalTimeSeconds    *int64  `json:"avgApprovalTimeSeconds" db:"avg_approval_time_seconds"`
	MedianApprovalTimeSeconds *int64  `json:"medianApprovalTimeSeconds" db:"median_approval_time_seconds"`
	P95ApprovalTimeSeconds    *int64  `json:"p95ApprovalTimeSeconds" db:"p95_approval_time_seconds"`
	
	// サイズ統計
	AvgLinesChanged     float64 `json:"avgLinesChanged" db:"avg_lines_changed"`
	MedianLinesChanged  float64 `json:"medianLinesChanged" db:"median_lines_changed"`
	AvgFilesChanged     float64 `json:"avgFilesChanged" db:"avg_files_changed"`
	MedianFilesChanged  float64 `json:"medianFilesChanged" db:"median_files_changed"`
	
	// 品質統計
	AvgReviewComments   float64 `json:"avgReviewComments" db:"avg_review_comments"`
	AvgReviewRounds     float64 `json:"avgReviewRounds" db:"avg_review_rounds"`
	FirstPassRate       float64 `json:"firstPassRate" db:"first_pass_rate"`
	
	// 複雑度統計
	AvgComplexityScore    float64 `json:"avgComplexityScore" db:"avg_complexity_score"`
	MedianComplexityScore float64 `json:"medianComplexityScore" db:"median_complexity_score"`
	
	// 生産性指標
	PRsPerDay           float64 `json:"prsPerDay" db:"prs_per_day"`
	LinesPerDay         float64 `json:"linesPerDay" db:"lines_per_day"`
	Throughput          float64 `json:"throughput" db:"throughput"`
	
	// トレンド情報
	CycleTimeTrend      string  `json:"cycleTimeTrend" db:"cycle_time_trend"`           // "increasing", "decreasing", "stable"
	ReviewTimeTrend     string  `json:"reviewTimeTrend" db:"review_time_trend"`
	QualityTrend        string  `json:"qualityTrend" db:"quality_trend"`
	
	// メタデータ
	GeneratedAt time.Time `json:"generatedAt" db:"generated_at"` // 集計実行日時
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`     // 最終更新日時
	Version     int       `json:"version" db:"version"`          // バージョン（楽観的ロック用）
	
	// 詳細情報（JSON格納）
	DetailedStatsJSON string `json:"detailedStatsJson" db:"detailed_stats_json"` // 詳細統計情報
	
	// インデックス用フィールド
	YearMonth  string `json:"yearMonth" db:"year_month"`   // YYYY-MM形式
	WeekOfYear string `json:"weekOfYear" db:"week_of_year"` // YYYY-WW形式
	DayOfYear  string `json:"dayOfYear" db:"day_of_year"`   // YYYY-DDD形式
}

// GetAggregatedMetricsSchema は集計メトリクスのスキーマ定義を返す
func GetAggregatedMetricsSchema() PRMetricsStorageSchema {
	return PRMetricsStorageSchema{
		TableName: "aggregated_metrics",
		Indexes: []IndexDefinition{
			// 主キー
			{
				Name:    "pk_aggregated_metrics",
				Columns: []string{"id"},
				Unique:  true,
				Type:    IndexTypeBTree,
			},
			// 複合ユニークキー（重複集計防止）
			{
				Name:    "uk_aggregated_metrics_unique",
				Columns: []string{"aggregation_level", "aggregation_period", "target_id", "period_start", "period_end"},
				Unique:  true,
				Type:    IndexTypeBTree,
			},
			// 集計レベル検索用
			{
				Name:    "idx_aggregated_metrics_level",
				Columns: []string{"aggregation_level"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 期間検索用
			{
				Name:    "idx_aggregated_metrics_period",
				Columns: []string{"aggregation_period"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 対象検索用
			{
				Name:    "idx_aggregated_metrics_target",
				Columns: []string{"target_id"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 日時範囲検索用
			{
				Name:    "idx_aggregated_metrics_time_range",
				Columns: []string{"period_start", "period_end"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 集計レベル＋対象の複合検索用
			{
				Name:    "idx_aggregated_metrics_level_target",
				Columns: []string{"aggregation_level", "target_id"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 期間＋対象の複合検索用
			{
				Name:    "idx_aggregated_metrics_period_target",
				Columns: []string{"aggregation_period", "target_id"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 最新データ検索用
			{
				Name:    "idx_aggregated_metrics_generated_at",
				Columns: []string{"generated_at"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 月次集計検索用
			{
				Name:    "idx_aggregated_metrics_year_month",
				Columns: []string{"year_month"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 週次集計検索用
			{
				Name:    "idx_aggregated_metrics_week_of_year",
				Columns: []string{"week_of_year"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
		},
	}
}

// TrendDataStorage はトレンドデータの永続化モデル
type TrendDataStorage struct {
	// 基本識別情報
	ID               string    `json:"id" db:"id"`                             // ユニークID
	MetricType       string    `json:"metricType" db:"metric_type"`            // "cycle_time", "review_time", "quality"
	AggregationLevel string    `json:"aggregationLevel" db:"aggregation_level"` // "team", "developer", "repository"
	TargetID         string    `json:"targetId" db:"target_id"`                // 対象ID
	
	// 期間情報
	PeriodStart time.Time `json:"periodStart" db:"period_start"` // 分析期間開始
	PeriodEnd   time.Time `json:"periodEnd" db:"period_end"`     // 分析期間終了
	
	// トレンド分析結果
	Slope            float64 `json:"slope" db:"slope"`                       // 傾き
	Intercept        float64 `json:"intercept" db:"intercept"`               // 切片
	CorrelationCoeff float64 `json:"correlationCoeff" db:"correlation_coeff"` // 相関係数
	Trend            string  `json:"trend" db:"trend"`                       // "increasing", "decreasing", "stable"
	Confidence       float64 `json:"confidence" db:"confidence"`             // 信頼度
	
	// 統計情報
	DataPoints      int     `json:"dataPoints" db:"data_points"`          // データポイント数
	StartValue      float64 `json:"startValue" db:"start_value"`          // 開始値
	EndValue        float64 `json:"endValue" db:"end_value"`              // 終了値
	ChangePercent   float64 `json:"changePercent" db:"change_percent"`    // 変化率（%）
	
	// メタデータ
	GeneratedAt time.Time `json:"generatedAt" db:"generated_at"` // 分析実行日時
	
	// 詳細データ（JSON格納）
	TimeSeriesData string `json:"timeSeriesData" db:"time_series_data"` // 時系列データ
}

// GetTrendDataSchema はトレンドデータのスキーマ定義を返す
func GetTrendDataSchema() PRMetricsStorageSchema {
	return PRMetricsStorageSchema{
		TableName: "trend_data",
		Indexes: []IndexDefinition{
			// 主キー
			{
				Name:    "pk_trend_data",
				Columns: []string{"id"},
				Unique:  true,
				Type:    IndexTypeBTree,
			},
			// 複合ユニークキー（重複分析防止）
			{
				Name:    "uk_trend_data_unique",
				Columns: []string{"metric_type", "aggregation_level", "target_id", "period_start", "period_end"},
				Unique:  true,
				Type:    IndexTypeBTree,
			},
			// メトリクスタイプ検索用
			{
				Name:    "idx_trend_data_metric_type",
				Columns: []string{"metric_type"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 集計レベル検索用
			{
				Name:    "idx_trend_data_level",
				Columns: []string{"aggregation_level"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 対象検索用
			{
				Name:    "idx_trend_data_target",
				Columns: []string{"target_id"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 期間検索用
			{
				Name:    "idx_trend_data_period",
				Columns: []string{"period_start", "period_end"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// トレンド検索用
			{
				Name:    "idx_trend_data_trend",
				Columns: []string{"trend"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
		},
	}
}

// BottleneckDataStorage はボトルネックデータの永続化モデル
type BottleneckDataStorage struct {
	// 基本識別情報
	ID         string    `json:"id" db:"id"`                 // ユニークID
	Type       string    `json:"type" db:"type"`             // "long_cycle_time", "multiple_review_rounds", "large_pr"
	PRID       string    `json:"prId" db:"pr_id"`            // 関連PR ID
	Severity   string    `json:"severity" db:"severity"`     // "high", "medium", "low"
	
	// ボトルネック詳細
	Description string  `json:"description" db:"description"` // 説明
	Value       float64 `json:"value" db:"value"`             // 値（時間、回数など）
	Threshold   float64 `json:"threshold" db:"threshold"`     // 閾値
	
	// 分析情報
	DetectedAt  time.Time `json:"detectedAt" db:"detected_at"`   // 検出日時
	ResolvedAt  *time.Time `json:"resolvedAt" db:"resolved_at"`  // 解決日時
	Status      string    `json:"status" db:"status"`           // "active", "resolved", "ignored"
	
	// 関連情報
	Author     string `json:"author" db:"author"`         // PR作成者
	Repository string `json:"repository" db:"repository"` // リポジトリ
	
	// メタデータ
	CreatedAt time.Time `json:"createdAt" db:"created_at"` // 作成日時
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"` // 更新日時
}

// GetBottleneckDataSchema はボトルネックデータのスキーマ定義を返す
func GetBottleneckDataSchema() PRMetricsStorageSchema {
	return PRMetricsStorageSchema{
		TableName: "bottleneck_data",
		Indexes: []IndexDefinition{
			// 主キー
			{
				Name:    "pk_bottleneck_data",
				Columns: []string{"id"},
				Unique:  true,
				Type:    IndexTypeBTree,
			},
			// PR ID検索用
			{
				Name:    "idx_bottleneck_data_pr_id",
				Columns: []string{"pr_id"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// タイプ検索用
			{
				Name:    "idx_bottleneck_data_type",
				Columns: []string{"type"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 重要度検索用
			{
				Name:    "idx_bottleneck_data_severity",
				Columns: []string{"severity"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// ステータス検索用
			{
				Name:    "idx_bottleneck_data_status",
				Columns: []string{"status"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 作成者検索用
			{
				Name:    "idx_bottleneck_data_author",
				Columns: []string{"author"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// リポジトリ検索用
			{
				Name:    "idx_bottleneck_data_repository",
				Columns: []string{"repository"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// 検出日時検索用
			{
				Name:    "idx_bottleneck_data_detected_at",
				Columns: []string{"detected_at"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
			// アクティブボトルネック検索用
			{
				Name:    "idx_bottleneck_data_active",
				Columns: []string{"status", "detected_at"},
				Unique:  false,
				Type:    IndexTypeBTree,
			},
		},
	}
}