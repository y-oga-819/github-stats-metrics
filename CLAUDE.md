# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Conversation Guidelines
- 常に日本語で会話する

## 作業フロー

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

##### 基本的なコミット戦略
- **コミット粒度**: 論理的な作業単位ごとに実行
- **コミットメッセージ**: Conventional Commits形式
  - `feat:` - 新機能
  - `fix:` - バグ修正
  - `docs:` - ドキュメント
  - `refactor:` - リファクタリング
  - `test:` - テスト追加・修正

```bash
git add [対象ファイル]
git commit -m "type: 簡潔な説明

詳細説明（必要に応じて）

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

##### TDD（テスト駆動開発）に基づく詳細コミット戦略

**t_wadaのTDD手法を適用した Red-Green-Refactor サイクル**:

**1. テストリスト作成フェーズ**:
```bash
git add test-list.md
git commit -m "docs: テストリスト作成

実装予定の機能のテストシナリオを列挙

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

**2. Red（失敗するテスト）フェーズ**:
```bash
git add [test-file]
git commit -m "test: [機能名] - 失敗するテストを追加

RED: [具体的なテストケース]を実装
期待する動作: [期待値]
現在の状態: テスト失敗

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

**3. Green（テストを通すための最小実装）フェーズ**:
```bash
git add [implementation-file]
git commit -m "feat: [機能名] - テストを通すための最小実装

GREEN: [テストケース]を成功させる仮実装
TODO: リファクタリングが必要

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

**4. Refactor（リファクタリング）フェーズ**:
```bash
git add [refactored-files]
git commit -m "refactor: [機能名] - [具体的な改善内容]

REFACTOR: [改善の詳細]
動作に変更なし、すべてのテスト継続通過

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

**5. 気づいた改善点の記録**:
```bash
git add test-list.md
git commit -m "docs: テストリスト更新

実装中に気づいた追加テストケース:
- [新しいテストケース1]  
- [新しいテストケース2]

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

**TDDコミットメッセージのコンベンション**:
```
type: [機能名] - [TDDフェーズ] [簡潔な説明]

[TDDフェーズ]: RED/GREEN/REFACTOR
[詳細説明]
[テストの状態や次のステップ]

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

**TDD実践時の原則**:
- 1つの機能に対して通常3-6回の細かいコミット
- 各フェーズでの確実な動作確認
- 「動作するきれいなコード」を目標とした段階的な改善
- Red-Green-Refactorサイクルの可視化によるプロセス追跡

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

## Development Commands

### Backend (Go)
- **Run backend locally**: `cd backend/app && go run cmd/main.go`
- **Build backend**: `cd backend/app && go build cmd/main.go`
- **Install dependencies**: `cd backend/app && go mod tidy`
- **Backend runs on**: http://localhost:8080

### Frontend (React + TypeScript)
- **Install dependencies**: `cd frontend && yarn install`
- **Run development server**: `cd frontend && yarn dev`
- **Build for production**: `cd frontend && yarn build`
- **Lint code**: `cd frontend && yarn lint`
- **Frontend runs on**: http://localhost:3000

### Docker Development
- **Start full stack**: `docker-compose up`
- **Rebuild containers**: `docker-compose up --build`
- Backend container: `dev-backend` (port 8080)
- Frontend container: `dev-frontend` (port 3000)

## Architecture Overview

### Backend Architecture (Clean Architecture)
```
app/
├── cmd/main.go              # Application entry point
├── server/webserver.go      # HTTP server setup with routing
├── application/             # Use cases and business logic
│   ├── pull_request/        # PR-related use cases
│   └── todo/               # Todo-related use cases
├── domain/                  # Core business entities
│   ├── developer/          # Developer domain objects
│   ├── pull_request/       # PR domain objects and requests
│   └── todo/               # Todo domain objects
├── infrastructure/         # External integrations
│   └── github_api/         # GitHub GraphQL API client
└── presentation/           # HTTP response formatting
    ├── pull_request/       # PR response presenters
    └── todo/               # Todo response presenters
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

## Environment Configuration

### Required Environment Variables (.env)
```
GITHUB_TOKEN=<your_github_token>
GITHUB_GRAPHQL_SEARCH_QUERY_TARGET_REPOSITORIES=owner/repo1,owner/repo2
```

### Vite Configuration
- API proxy configured for `/api` routes to backend
- Supports Docker environment with `API_URL` environment variable

## Key Technical Details

### GitHub Integration
- Uses GitHub GraphQL API v4 with `githubv4` Go library
- Searches for merged PRs within date ranges and specific repositories
- Filters by author (developer) and excludes epic branches
- Handles pagination for large result sets

### Metrics Calculated
- **Review Time**: Time from PR creation to first review
- **Approval Time**: Time from first review to final approval  
- **Merge Time**: Time from approval to merge
- **PR Count**: Number of PRs per sprint
- **Dev/Day/Developer**: PRs per developer per day (assuming 5-day sprints)

### Current Issues
- Chart data has duplicate entries (current branch: `frontend/fix/sync-metrics-and-sprint`)
- Data is converted to Maps to eliminate duplicates
- Hardcoded sprint data instead of dynamic API

### Development Notes
- Backend uses clean architecture with domain-driven design
- Frontend uses feature-based organization with React Query for API calls
- Chart.js and ApexCharts for data visualization
- TailwindCSS for styling