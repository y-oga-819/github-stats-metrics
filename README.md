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

詳細なセットアップ手順については、[セットアップガイド](./docs/SETUP.md)を参照してください。


## 📄 ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルを参照

---

**最終更新**: 2024-06-22