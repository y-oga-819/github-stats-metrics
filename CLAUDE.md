# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Conversation Guidelines
- 常に日本語で会話する

## 📋 クイックリファレンス

### よく使うコマンド
```bash
# 開発環境起動
cd frontend && yarn dev          # Frontend (http://localhost:3000)
cd backend/app && go run cmd/main.go  # Backend (http://localhost:8080)
docker-compose up               # 全体起動

# 依存関係
cd frontend && yarn install     # Frontend
cd backend/app && go mod tidy   # Backend

# 品質チェック
cd frontend && yarn lint        # Frontend Lint
cd backend/app && go fmt ./...  # Backend Format
```

### 新規参加者向けセットアップ
1. 環境変数設定: `.env`ファイルに`GITHUB_TOKEN`を設定
2. 依存関係インストール: Frontend(`yarn install`) + Backend(`go mod tidy`)
3. 開発サーバー起動: `docker-compose up`または個別起動

## 🔄 開発ワークフロー

### 基本的な作業手順
1. **新規ブランチ作成**: 指示を受けたら必ずmainブランチから新規ブランチを作成
2. **細かなコミット**: 作業進行中は適時、細かい粒度でコミットを実行
3. **プルリクエスト作成**: 作業完了後、ghコマンドでPRを作成

### 詳細手順

#### 1. ブランチ作成

**基本パターン（mainから派生）**:
```bash
git checkout main
git pull origin main
git checkout -b [category]/[feature-name]
```

**依存関係がある場合（既存ブランチから派生）**:
```bash
git checkout [base-branch]
git pull origin [base-branch]
git checkout -b [category]/[feature-name]
```

**ブランチ選択の判断基準**:
- **mainから派生**: 独立した機能・修正の場合
- **既存ブランチから派生**: 以下の場合
  - 未マージブランチの機能に依存する作業
  - 同一機能の段階的実装
  - 前の作業の続きや改良
  - 連続する作業の流れがある場合

**依存関係がある場合の対応方針**:
1. 作業を中断せず、既存ブランチから派生して継続
2. PR作成時に依存関係を明記
3. マージ順序の調整はユーザーが判断

**ブランチ命名規則**:
- `feature/[機能名]` - 新機能追加
- `fix/[修正内容]` - バグ修正
- `refactor/[対象]` - リファクタリング
- `docs/[ドキュメント名]` - ドキュメント作成・更新
- `documentation/[分析内容]` - 分析・調査系ドキュメント

#### 2. 作業とコミット

**基本コミット戦略**:
- **粒度**: 論理的な作業単位ごと（TDDの場合は Red-Green-Refactor サイクル）
- **形式**: Conventional Commits（`feat:`, `fix:`, `docs:`, `refactor:`, `test:`）

```bash
git add [対象ファイル]
git commit -m "type: 簡潔な説明

詳細説明（必要に応じて）

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

**TDD適用時の細分化コミット**:
```bash
# RED: 失敗するテスト → GREEN: 最小実装 → REFACTOR: 改善
git commit -m "test: [機能名] - RED: [テストケース]"
git commit -m "feat: [機能名] - GREEN: テスト通過実装"  
git commit -m "refactor: [機能名] - [改善内容]"
```

#### 3. プルリクエスト作成
```bash
git push -u origin [ブランチ名]
gh pr create --title "[タイトル]" --body "[詳細説明]"
```

**PRテンプレート構成**:
- Summary: 変更概要
- 主要な変更点
- 依存関係: 他PRへの依存がある場合は明記
- Test plan: テスト/確認事項
- 🤖 Generated with [Claude Code] 署名

**依存関係があるPRの場合**:
- ベースブランチを明記: "depends on PR #XX"
- マージ順序の指示: "Merge after PR #XX"
- 影響範囲の説明: 依存する機能との関係性

## 🛠️ 開発環境・コマンド

### 開発サーバー起動
```bash
# Frontend (React + TypeScript)
cd frontend && yarn dev              # http://localhost:3000

# Backend (Go)  
cd backend/app && go run cmd/main.go # http://localhost:8080

# Docker（全体）
docker-compose up                    # Backend:8080, Frontend:3000
```

### 依存関係管理
```bash
# Frontend
cd frontend && yarn install

# Backend  
cd backend/app && go mod tidy
```

### ビルド・テスト
```bash
# Frontend
cd frontend && yarn build           # 本番ビルド
cd frontend && yarn test            # テスト実行
cd frontend && yarn lint            # Lint実行

# Backend
cd backend/app && go build cmd/main.go  # ビルド
cd backend/app && go test ./...         # テスト実行
cd backend/app && go fmt ./...          # フォーマット
```

## 🏗️ アーキテクチャ概要

### Backend Architecture (Clean Architecture + DDD)
```
app/
├── cmd/main.go              # Application entry point
├── server/                  # HTTP server setup
│   ├── webserver.go         # Router and middleware setup
│   └── cors.go              # CORS configuration
├── application/             # Use cases and business logic
│   ├── analytics/           # Analytics metrics aggregation
│   ├── pull_request/        # PR-related use cases
│   └── todo/               # Todo-related use cases
├── domain/                  # Core business entities
│   ├── analytics/           # Analytics domain models
│   ├── developer/          # Developer domain objects
│   ├── pull_request/       # PR domain objects and requests
│   └── todo/               # Todo domain objects
├── infrastructure/         # External integrations and persistence
│   ├── github_api/         # GitHub GraphQL API client
│   ├── memory/             # In-memory repository implementations
│   └── repository/         # Repository interface implementations
├── presentation/           # HTTP handlers and response formatting
│   ├── analytics/          # Analytics metrics endpoints
│   ├── health/             # Health check endpoints
│   ├── pull_request/       # PR response presenters
│   └── todo/               # Todo response presenters
├── shared/                 # Cross-cutting concerns
│   ├── config/             # Configuration management
│   ├── errors/             # Error handling utilities
│   ├── logger/             # Logging utilities
│   ├── logging/            # Structured logging
│   ├── metrics/            # Metrics collection
│   ├── middleware/         # HTTP middleware
│   ├── monitoring/         # Application monitoring
│   └── utils/              # Common utilities
├── cli/                    # Command-line interface
├── config/                 # Application configuration
└── integration_test/       # Integration test suites
```

### Frontend Architecture (Feature-Based)
```
src/
├── App.tsx                 # Main app component with navigation
├── Router.tsx              # Route definitions
└── features/               # Feature-based organization
    ├── Chart/              # Metrics visualization components
    │   ├── Chart.tsx       # Main chart container with data fetching
    │   ├── MetricsChart.tsx    # PR timing metrics chart
    │   ├── PrCountChart.tsx    # PR count visualization
    │   └── DevDayDeveloperChart.tsx # Developer productivity chart
    ├── pullrequestlist/    # PR list functionality
    │   ├── PullRequestsFetcher.ts  # API client for PR data
    │   └── PullRequest.tsx         # PR display components
    ├── sprint/             # Sprint detail views
    └── sprintlist/         # Sprint list management
```

### API Endpoints
- `GET /api/pull_requests?startdate=YYYY-MM-DD&enddate=YYYY-MM-DD&developers[]=name1&developers[]=name2`
- `GET /api/todos`

### Data Flow
1. **Frontend**: Sprint data (hardcoded) → API requests to backend
2. **Backend**: HTTP request → Use case → GitHub API client → GraphQL query → Response formatting
3. **GitHub API**: GraphQL queries filter by date range, repositories, and developers
4. **Metrics Calculation**: Frontend calculates timing metrics from PR lifecycle data

## ⚙️ 環境設定

### 必須環境変数 (.env)
```env
GITHUB_TOKEN=<your_github_token>
GITHUB_GRAPHQL_SEARCH_QUERY_TARGET_REPOSITORIES=owner/repo1,owner/repo2
```

### Vite設定
- API proxy: `/api` → backend
- Docker対応: `API_URL`環境変数

## 🚨 トラブルシューティング

### よくあるエラーと解決方法

**Docker関連**:
```bash
# コンテナが起動しない
docker-compose down && docker-compose up --build

# ポート競合エラー  
lsof -i :3000  # プロセス確認
kill -9 <PID>  # プロセス終了
```

**GitHub API関連**:
```bash
# API制限エラー
# → GITHUB_TOKENの権限確認
# → リクエスト頻度の調整

# GraphQL構文エラー  
# → クエリの構文チェック
# → GitHub GraphQL Explorer で検証
```

**依存関係エラー**:
```bash
# Frontend
rm -rf node_modules yarn.lock && yarn install

# Backend
go clean -modcache && go mod tidy
```

## 📊 技術詳細

### GitHub連携
- GitHub GraphQL API v4 + `githubv4` Go library
- 日付範囲・リポジトリ・開発者でのPR検索
- Epic branchの除外、ページネーション対応

### 計算メトリクス
- **Review Time**: PR作成 → 初回レビュー
- **Approval Time**: 初回レビュー → 最終承認  
- **Merge Time**: 承認 → マージ
- **PR Count**: スプリント当たりPR数
- **Dev/Day/Developer**: 開発者1日あたりPR数（5日スプリント想定）

### 技術スタック
- **Backend**: Clean Architecture + DDD (Go)
- **Frontend**: Feature-based + React Query (React/TypeScript)
- **可視化**: Chart.js, ApexCharts
- **スタイル**: TailwindCSS