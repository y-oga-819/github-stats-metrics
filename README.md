# GitHub Stats Metrics

> GitHub開発メトリクスの可視化・分析ツール

![License](https://img.shields.io/badge/license-MIT-blue.svg)
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

## 🚀 セットアップ

詳細なセットアップ手順については、[セットアップガイド](./docs/SETUP.md)を参照してください。


## 📄 ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルを参照

---

**最終更新**: 2024-06-22