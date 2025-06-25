# Phase 1 実装TODOリスト

**目標**: PR分析エンジンの拡張と基本メトリクス算出システムの実装

**期間**: 2-3スプリント

---

## 🏗️ 1. PR分析エンジンの拡張

### 1.1 ドメインモデルの拡張

- [ ] **PRメトリクス構造体の実装**
  - [ ] `PRMetrics` ドメインオブジェクトの作成
    - サイズメトリクス（行数、ファイル数、複雑度）
    - 時間メトリクス（サイクルタイム、各段階の時間）
    - 品質メトリクス（レビューコメント数、修正回数）
  - [ ] ファイル: `backend/app/domain/pull_request/pr_metrics.go`

- [ ] **PRComplexity 分析オブジェクトの実装**
  - [ ] 複雑度算出ロジック（変更ファイル種別による重み付け）
  - [ ] PRサイズ判定（Small/Medium/Large/XLarge）
  - [ ] ファイル: `backend/app/domain/pull_request/pr_complexity.go`

### 1.2 GitHub API データ取得の拡張

- [ ] **PR詳細データの取得機能強化**
  - [ ] レビューコメント数の取得
  - [ ] ファイル変更詳細の取得（追加/削除行数をファイル毎に）
  - [ ] レビュー履歴の詳細取得（レビュアー、時間、状態）
  - [ ] ファイル: `backend/app/infrastructure/github_api/github_api.go`

- [ ] **GraphQLクエリの拡張**
  - [ ] レビューノード情報の追加
  - [ ] ファイル変更詳細の追加
  - [ ] レビューコメント情報の追加
  - [ ] ファイル: `backend/app/infrastructure/github_api/graphql_queries.go` (新規)

### 1.3 データ変換・分析ロジック

- [ ] **PR分析サービスの実装**
  - [ ] PRメトリクス算出ロジック
  - [ ] 複雑度スコア算出
  - [ ] レビュー効率分析
  - [ ] ファイル: `backend/app/domain/pull_request/pr_analysis_service.go`

- [ ] **コンバーター機能の拡張**
  - [ ] GitHub APIレスポンスからPRMetricsへの変換
  - [ ] レビューデータの正規化
  - [ ] ファイル: `backend/app/infrastructure/github_api/metrics_converter.go` (新規)

---

## 📊 2. 基本メトリクス算出システム

### 2.1 時間計測機能

- [ ] **サイクルタイム計算の実装**
  - [ ] PR作成〜初回レビューまでの時間
  - [ ] 初回レビュー〜承認までの時間
  - [ ] 承認〜マージまでの時間
  - [ ] 全体サイクルタイム
  - [ ] ファイル: `backend/app/domain/pull_request/cycle_time_calculator.go`

- [ ] **レビュー時間分析の実装**
  - [ ] レビュー待ち時間の算出
  - [ ] レビュー実施時間の推定
  - [ ] レビューラウンド数の計算
  - [ ] ファイル: `backend/app/domain/pull_request/review_time_analyzer.go`

### 2.2 集計・統計機能

- [ ] **メトリクス集計サービスの実装**
  - [ ] 期間別の統計算出（日/週/スプリント/月別）
  - [ ] 開発者別の統計算出
  - [ ] リポジトリ別の統計算出
  - [ ] ファイル: `backend/app/application/analytics/metrics_aggregator.go` (新規ディレクトリ)

- [ ] **統計計算ユーティリティ**
  - [ ] 平均値、中央値、標準偏差の算出
  - [ ] パーセンタイル計算（50%, 75%, 90%, 95%）
  - [ ] トレンド分析（改善/悪化傾向）
  - [ ] ファイル: `backend/app/shared/utils/statistics.go` (新規)

---

## 🗄️ 3. データストレージ設計・実装

### 3.1 データモデルの設計

- [ ] **PRメトリクスストレージモデルの設計**
  - [ ] 時系列データストレージの検討（JSON/SQLite/InfluxDB）
  - [ ] データスキーマの定義
  - [ ] インデックス設計
  - [ ] ファイル: `backend/app/domain/analytics/storage_models.go` (新規)

- [ ] **集計データモデルの設計**
  - [ ] 日次/週次/月次集計テーブル設計
  - [ ] 開発者別統計テーブル設計
  - [ ] パフォーマンス最適化考慮
  - [ ] ファイル: `backend/app/domain/analytics/aggregation_models.go` (新規)

### 3.2 リポジトリ層の実装

- [ ] **PRメトリクスリポジトリの実装**
  - [ ] PRメトリクスの永続化機能
  - [ ] 期間・条件での検索機能
  - [ ] 一括挿入・更新機能
  - [ ] ファイル: `backend/app/infrastructure/storage/pr_metrics_repository.go` (新規)

- [ ] **集計データリポジトリの実装**
  - [ ] 集計結果の保存・取得
  - [ ] 効率的なクエリ機能
  - [ ] データ保持期間管理
  - [ ] ファイル: `backend/app/infrastructure/storage/aggregation_repository.go` (新規)

---

## 🔧 4. API・インターフェース拡張

### 4.1 新しいAPIエンドポイント

- [ ] **PRメトリクス取得API**
  - [ ] `/api/pull_requests/{id}/metrics` - 個別PRメトリクス
  - [ ] `/api/metrics/cycle_time` - サイクルタイム統計
  - [ ] `/api/metrics/review_time` - レビュー時間統計
  - [ ] ファイル: `backend/app/presentation/analytics/handler.go` (新規)

- [ ] **集計データ取得API**
  - [ ] `/api/analytics/team_metrics` - チーム統計
  - [ ] `/api/analytics/developer_metrics` - 開発者別統計
  - [ ] `/api/analytics/trends` - トレンド分析
  - [ ] ファイル: `backend/app/presentation/analytics/aggregation_handler.go` (新規)

### 4.2 レスポンス形式の定義

- [ ] **APIレスポンス構造の定義**
  - [ ] メトリクスレスポンスDTO
  - [ ] 統計レスポンスDTO
  - [ ] エラーレスポンス統一
  - [ ] ファイル: `backend/app/presentation/analytics/response_models.go` (新規)

---

## 🧪 5. テスト実装

### 5.1 ユニットテスト

- [ ] **ドメインロジックのテスト**
  - [ ] PRメトリクス算出のテスト
  - [ ] サイクルタイム計算のテスト
  - [ ] 複雑度算出のテスト
  - [ ] ファイル: `*_test.go` 各ドメインファイルに対応

- [ ] **リポジトリ層のテスト**
  - [ ] データ永続化のテスト
  - [ ] 検索・取得機能のテスト
  - [ ] エラーハンドリングのテスト

### 5.2 統合テスト

- [ ] **API統合テスト**
  - [ ] 新しいエンドポイントのテスト
  - [ ] データフロー全体のテスト
  - [ ] パフォーマンステスト

---

## ⚙️ 6. 設定・インフラ

### 6.1 設定管理の拡張

- [ ] **分析機能の設定追加**
  - [ ] メトリクス計算の閾値設定
  - [ ] データ保持期間の設定
  - [ ] 集計頻度の設定
  - [ ] ファイル: `backend/app/shared/config/analytics_config.go` (新規)

### 6.2 バッチ処理の実装

- [ ] **定期実行バッチの実装**
  - [ ] 日次メトリクス集計バッチ
  - [ ] 古いデータクリーンアップバッチ
  - [ ] トレンド分析更新バッチ
  - [ ] ファイル: `backend/app/batch/analytics_batch.go` (新規)

---

## 📝 7. ドキュメント

### 7.1 技術ドキュメント

- [ ] **アーキテクチャドキュメント**
  - [ ] データフロー図
  - [ ] メトリクス算出ロジック説明
  - [ ] API仕様書更新

- [ ] **運用ドキュメント**
  - [ ] 設定方法の説明
  - [ ] バッチ処理の運用手順
  - [ ] トラブルシューティング

---

## 🎯 Phase 1 完了の定義

以下がすべて完了した時点でPhase 1完了とする：

1. ✅ **PRメトリクス取得機能** - 基本的なPRサイズ・時間・品質メトリクスの算出
2. ✅ **基本統計API** - サイクルタイム、レビュー時間の統計取得
3. ✅ **データ永続化** - メトリクスデータの保存・取得機能
4. ✅ **テストカバレッジ** - 新機能の80%以上のテストカバレッジ
5. ✅ **ドキュメント整備** - API仕様とアーキテクチャドキュメントの更新

---

## 📅 推奨実装順序

### Week 1-2: ドメインモデル・基盤
1. PRメトリクス構造体の実装
2. PR分析サービスの基本実装
3. サイクルタイム計算機能

### Week 3-4: データ取得・保存
1. GitHub API拡張
2. データストレージ実装
3. メトリクス集計サービス

### Week 5-6: API・統合
1. 新APIエンドポイント実装
2. テスト実装
3. ドキュメント整備

---

*作成日: 2025-06-25*  
*Phase 1 実装ガイド*