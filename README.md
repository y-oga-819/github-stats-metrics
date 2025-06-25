# GitHub Stats Metrics

> GitHub開発メトリクスの可視化・分析ツール

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Docker](https://img.shields.io/badge/Docker-supported-2496ED.svg?logo=docker)
![React](https://img.shields.io/badge/React-18-61DAFB.svg?logo=react)
![TypeScript](https://img.shields.io/badge/TypeScript-latest-3178C6.svg?logo=typescript)
![Go](https://img.shields.io/badge/Go-1.21-00ADD8.svg?logo=go)
![Node.js](https://img.shields.io/badge/Node.js-16+-339933.svg?logo=node.js)

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
- Chart.js (データ可視化)
- React Router (ナビゲーション)
- Tailwind CSS (スタイリング)

**バックエンド**
- Gorilla Mux (HTTPルーティング)
- GitHub GraphQL API v4

## 🚀 クイックスタート

### 開発環境

```bash
# リポジトリのクローン
git clone https://github.com/y-oga-819/github-stats-metrics.git
cd github-stats-metrics

# 開発環境で起動
make dev
# または
docker-compose up
```

### 本番環境

```bash
# 本番環境で起動
make prod
# または
docker-compose -f docker-compose.prod.yml up
```

### アクセス

- フロントエンド: http://localhost:3000
- バックエンド: http://localhost:8080
- ヘルスチェック: http://localhost:8080/health

### その他のコマンド

```bash
make help          # 利用可能なコマンド一覧
make logs          # ログ表示
make health        # コンテナ状態確認
make test          # テスト実行
make clean         # コンテナ・イメージ削除
```

### 監視環境

```bash
# 監視スタック起動（Prometheus + Grafana + Loki）
make monitoring

# 監視サービスURL表示
make monitoring-urls
```

#### 監視サービス

- **Prometheus**: http://localhost:9090 (メトリクス収集)
- **Grafana**: http://localhost:3001 (ダッシュボード - admin/admin123)
- **Loki**: http://localhost:3100 (ログ集約)
- **アプリメトリクス**: http://localhost:8080/metrics
- **ヘルスチェック**: http://localhost:8080/health

詳細なセットアップ手順については、[セットアップガイド](./docs/SETUP.md)を参照してください。


## 📄 ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルを参照

---

**最終更新**: 2024-06-22