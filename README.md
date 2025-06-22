# GitHub Stats Metrics

> GitHub開発メトリクスの可視化・分析ツール

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## 🎯 概要

GitHub Stats Metricsは、GitHubのPull Requestsデータを分析し、開発チームのメトリクスを可視化するWebアプリケーションです。

## ✨ 主な機能

- **レビューまでの時間**: PR作成からレビューまでの平均時間
- **承認までの時間**: レビューから承認までの平均時間  
- **マージまでの時間**: 承認からマージまでの平均時間
- **PR数**: スプリントごとのPull Request数
- **Dev/Day/Developer**: 開発者1人あたりの日別開発効率

## 🛠️ 技術スタック

**フロントエンド**
- React 18 + TypeScript
- Chart.js (データ可視化)
- React Router (ナビゲーション)
- Tailwind CSS (スタイリング)

**バックエンド**
- Go 1.21
- Gorilla Mux (HTTPルーティング)
- GitHub GraphQL API v4

## 🚀 セットアップ

### 前提条件
- Node.js 16+ 
- Go 1.21+
- GitHub Personal Access Token

### インストール

1. **リポジトリのクローン**
```bash
git clone https://github.com/y-oga-819/github-stats-metrics.git
cd github-stats-metrics
```

2. **バックエンドセットアップ**
```bash
cd backend/app

# 環境変数設定
cp .env.example .env
# .envファイルを編集してGITHUB_TOKENを設定

# 依存関係インストール
go mod download

# 実行
go run cmd/main.go
```

3. **フロントエンドセットアップ**
```bash
cd frontend

# 依存関係インストール
npm install

# 開発サーバー起動
npm start
```

### 環境変数

**.env設定例**
```bash
GITHUB_TOKEN=your_github_personal_access_token
GITHUB_GRAPHQL_SEARCH_QUERY_TARGET_REPOSITORIES=repo1,repo2,repo3
ALLOWED_ORIGINS=http://localhost:3000
```


## 📄 ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルを参照

---

**最終更新**: 2024-06-22