# セットアップガイド

## 前提条件

- Node.js 16+ 
- Go 1.21+
- GitHub Personal Access Token

## インストール

### 1. リポジトリのクローン

```bash
git clone https://github.com/y-oga-819/github-stats-metrics.git
cd github-stats-metrics
```

### 2. バックエンドセットアップ

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

### 3. フロントエンドセットアップ

```bash
cd frontend

# 依存関係インストール
npm install

# 開発サーバー起動
npm start
```

## 環境変数設定

### .env設定例

```bash
GITHUB_TOKEN=your_github_personal_access_token
GITHUB_GRAPHQL_SEARCH_QUERY_TARGET_REPOSITORIES=repo1,repo2,repo3
ALLOWED_ORIGINS=http://localhost:3000
```

### GitHub Personal Access Tokenの取得方法

1. GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. "Generate new token (classic)" をクリック
3. 必要なスコープを選択:
   - `repo` (プライベートリポジトリの場合)
   - `public_repo` (パブリックリポジトリの場合)
4. トークンをコピーして `.env` ファイルに設定

## 起動確認

1. バックエンド: http://localhost:8080
2. フロントエンド: http://localhost:3000

両方のサーバーが正常に起動し、フロントエンドからバックエンドAPIにアクセスできることを確認してください。

## トラブルシューティング

### よくある問題

- **CORS エラー**: `ALLOWED_ORIGINS` の設定を確認
- **GitHub API エラー**: トークンの権限とレート制限を確認
- **ポート競合**: 他のアプリケーションがポート8080/3000を使用していないか確認