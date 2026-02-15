# Tsubo（坪）

> AI駆動開発のためのマイクロサービスフレームワーク

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-proof%20of%20concept-green.svg)]()

## 概要

**Tsubo（壺）** は、AI（LLM）による並列実装を加速させ、ハルシネーションを削減するために設計された、マイクロサービス開発フレームワークです。

### 壺のメタファー

```
   ┌─────────────────────────────────────┐
   │  壺（Tsubo）= アプリケーション全体   │  ← 人間が決める
   │                                     │
   │  ┌──────────┐  ┌──────────┐        │
   │  │  TODO    │  │   User   │  ...   │  ← 固体オブジェクト
   │  │ Contract │  │ Contract │        │     (ドメイン)
   │  │  ┌────┐  │  │  ┌────┐  │        │
   │  │  │実装│  │  │  │実装│  │        │  ← AIが決める
   │  │  └────┘  │  │  └────┘  │        │
   │  └──────────┘  └──────────┘        │
   │       ↓              ↓              │
   │  todo-service   user-service       │  ← マイクロサービス
   └─────────────────────────────────────┘
```

**人は壺の形（アプリケーション）と中に入れる固体オブジェクト（ドメイン）を決める。AIは各オブジェクトの内部構造を作る。**

- **壺**: アプリケーション全体（容器）
- **固体オブジェクト（ドメイン）**: 具体的なビジネス概念（触れるもの）
- **マイクロサービス**: 各固体オブジェクトの実装
- **実装の詳細**: オブジェクトの内部構造（AIが決める）

### なぜ Tsubo なのか？

現代のAI駆動開発では、以下の課題があります：
- 大規模なコードベースでは、AIがコンテキストを失いやすい
- モノリシックな実装では、並列開発が困難
- 明確な境界がないと、AIが整合性のないコードを生成する

Tsubo は、これらの課題を「**壺＝コンテキストの境界**」という発想で解決します。

**Tsubo の哲学:**
- 人間は「**何をすべきか**」（Contract定義、ドメインの境界）に集中
- AIは「**どう実装するか**」（実装の詳細）に集中
- **1つの壺（アプリケーション）に複数の固体オブジェクト（ドメイン/マイクロサービス）を入れる**
- 各固体オブジェクトは独立し、疎結合を実現

## 核心的なアイデア

```
小さなサービス → AIが理解しやすい → ハルシネーション削減
     ↓
明確な契約 → 並列実装可能 → 開発速度向上
     ↓
自動検証 → 品質保証 → 信頼性の高いコード
```

## 主要機能

### 🎯 サービス定義の標準化
宣言的なYAMLフォーマットで、マイクロサービスの仕様を定義。AIが理解しやすい形式。

### 🔄 並列実装オーケストレーション
複数のAIエージェントが、依存関係を考慮しながら並列にサービスを実装。

### ✅ 自動検証・テスト
契約テスト、型チェック、統合テストを自動的に実行し、品質を保証。

### 🚀 高速な開発サイクル
従来の3-5倍の速度でマイクロサービスを実装。

## クイックスタート

### インストール

```bash
# リポジトリをクローン
git clone https://github.com/staka121/tsubo.git
cd tsubo

# Tsubo CLI をビルド
go build -o tsubo ./cmd/tsubo

# またはインストール
go install ./cmd/tsubo
```

### AI駆動で新しいサービスを実装（完全自動化）

```bash
# 1. 新しいサービステンプレートを作成
tsubo new user-service

# 2. .tsubo.yaml ファイルを作成・編集してサービスを追加
# （例: poc/contracts/tsubo-todo-app.tsubo.yaml を参照）

# 3. プロンプト生成のみ（手動実行用）
tsubo build ./poc/contracts/tsubo-todo-app.tsubo.yaml

# または

# 4. AI駆動で自動実装（Claude API使用）
export ANTHROPIC_API_KEY=your-api-key
tsubo build ./poc/contracts/tsubo-todo-app.tsubo.yaml --ai

# 並行数を制限する場合
tsubo build ./poc/contracts/tsubo-todo-app.tsubo.yaml --ai --concurrency 4

# 5. 実装完了後、サービスを起動
tsubo run -d

# 6. テスト実行
tsubo verify
```

### PoC の実行（Tsubo TODO アプリケーション）

```bash
# リポジトリをクローン
git clone https://github.com/staka121/tsubo.git
cd tsubo

# Tsubo CLI をビルド
go build -o tsubo ./cmd/tsubo

# 実装プランを確認
tsubo build ./poc/contracts/tsubo-todo-app.tsubo.yaml

# 実装済みサービスを起動
tsubo run -d

# 統合テスト
tsubo verify
```

**含まれるドメイン（固体オブジェクト）:**
- User ドメイン（user-service: port 8080）
- TODO ドメイン（todo-service: port 8081）

## アーキテクチャ

```
┌─────────────────────────────────────────┐
│    Contract Definitions (YAML)          │
│  - tsubo-todo-app.tsubo.yaml            │
│  - user-service.object.yaml             │
│  - todo-service.object.yaml             │
└────────────┬────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────┐
│      tsubo-plan (Go)                    │
│  - Contract 解析                         │
│  - 依存関係分析                          │
│  - Wave 生成（実装順序決定）             │
│  - Implementation Plan 出力             │
└────────────┬────────────────────────────┘
             │
             ▼ JSON Plan
┌──────────┬──────────┬──────────┬────────┐
│ AI Agent │ AI Agent │ AI Agent │  ...   │
│ (Wave 0) │ (Wave 0) │ (Wave 1) │        │
│          │          │          │        │
│ user-    │ other-   │ todo-    │        │
│ service  │ service  │ service  │        │
└────┬─────┴────┬─────┴────┬─────┴────┬───┘
     │          │          │          │
     ▼          ▼          ▼          ▼
┌────────────────────────────────────────┐
│       Generated Services (Go)          │
│  - 100% Contract 準拠                   │
│  - Docker 化済み                        │
│  - テスト付き                            │
└────────────────────────────────────────┘
```

## 技術スタック

- **フレームワーク:** Go 1.22
  - tsubo-plan: Contract解析・プランニングツール
  - 型安全なYAMLパース
  - 依存関係解析
  - シングルバイナリ配布

- **生成サービス:** Go 1.22（推奨）
  - シンプルで一貫性のあるコード
  - ハルシネーション削減
  - 標準ライブラリ中心
  - 将来的に TypeScript, Python もサポート

- **Contract Definition:** YAML
  - `.tsubo.yaml`: 壺（アプリケーション）の定義
  - `.object.yaml`: オブジェクト（サービス）の定義
  - 人間・AI両方が読みやすい

- **デプロイ:** Docker & Docker Compose
  - Docker First 原則
  - 環境の完全分離
  - 再現性の保証

### なぜ Go言語なのか？

**Go の「誰が書いても同じコードになる」という特性が、AIによるハルシネーションを劇的に低減します。**

詳細は [WHY_GO.md](./docs/WHY_GO.md) を参照。

## プロジェクト状況

**現在のステータス: ✅ 完全自動化パイプライン完成**

- [x] **基本思想の整理**
- [x] **サービス定義フォーマットの仕様策定**（Contract Design）
- [x] **開発原則の確立**（Docker First & 質疑のタイミング）
- [x] **ファイルフォーマットの確立**（.tsubo.yaml / .object.yaml）
- [x] **PoC 実装完了**（TODO アプリケーション）
  - [x] 壺（アプリケーション全体）の設計
  - [x] User ドメイン（固体オブジェクト1）
    - [x] Contract 定義
    - [x] **AI による Go 実装**
    - [x] Docker 化
    - [x] テスト（100% Contract 準拠）
  - [x] TODO ドメイン（固体オブジェクト2）
    - [x] Contract 定義
    - [x] **AI による Go 実装**
    - [x] Docker 化
    - [x] User ドメインとの連携
    - [x] テスト（100% Contract 準拠）
  - [x] docker-compose による全体のオーケストレーション
  - [x] 統合テスト（ドメイン間連携の確認）
- [x] **実装ツールチェーン完成**
  - [x] **tsubo-plan** (Go) - Contract解析と実装プラン生成
    - [x] 依存関係の自動解析
    - [x] Wave（実装順序）の自動決定
    - [x] JSON形式の実装プラン出力
  - [x] **tsubo-execute Phase 1** - プロンプト生成
    - [x] 包括的な実装プロンプト生成
    - [x] コンテキストファイル埋め込み
    - [x] Contract 埋め込み
  - [x] **tsubo-execute Phase 2** - Claude API 自動実行 ✨NEW
    - [x] Claude API クライアント実装
    - [x] Wave 単位の並列実行
    - [x] 依存関係に基づく順次実行
    - [x] リアルタイム進捗表示
    - [x] エラーハンドリング
    - [x] 実行サマリー

### 完成した自動化パイプライン

```
Contract 定義 (人間)
   ↓
tsubo-plan (自動解析)
   ↓
tsubo-execute --execute (AI実装) ← 完全自動化！
   ↓
マイクロサービス実装 (100% Contract準拠)
```

### 次のマイルストーン

- [ ] tsubo-plan の機能拡張
  - [x] より複雑な依存関係グラフのサポート（トポロジカルソート実装）
  - [ ] 実装プランの可視化
  - [x] サイクル検出（循環依存の検出）
- [ ] tsubo-execute の強化
  - [ ] リトライロジック
  - [ ] 部分的な再実行
  - [ ] 複数モデルのサポート
- [ ] 検証エンジンの実装
  - [ ] Contract 準拠チェックの自動化
  - [ ] パフォーマンステスト
  - [ ] セキュリティスキャン
- [ ] 他言語サポート
  - [ ] TypeScript サービス生成
  - [ ] Python サービス生成

## ドキュメント

### 核心思想
- [設計思想（PHILOSOPHY.md）](./docs/PHILOSOPHY.md) - Tsubo の核心的な考え方
- [ドメイン設計（DOMAIN_DESIGN.md）](./docs/DOMAIN_DESIGN.md) - 壺と固体オブジェクトの関係

### 開発ガイド
- [開発原則（DEVELOPMENT_PRINCIPLES.md）](./docs/DEVELOPMENT_PRINCIPLES.md) - Docker First & 質疑のタイミング
- [Contract 設計（CONTRACT_DESIGN.md）](./docs/CONTRACT_DESIGN.md) - Contract フォーマットの詳細
- [ファイルフォーマット（docs/FILE_FORMATS.md）](./docs/FILE_FORMATS.md) - .tsubo.yaml と .object.yaml
- [なぜ Go 言語か（WHY_GO.md）](./docs/WHY_GO.md) - Go 言語選択の理由

### ツール
- [tsubo-plan（cmd/tsubo-plan/README.md）](./cmd/tsubo-plan/README.md) - Contract解析・実装プランニング
- [tsubo-execute（cmd/tsubo-execute/README.md）](./cmd/tsubo-execute/README.md) - プロンプト生成・Claude API自動実行

## コントリビューション

現在は PoC フェーズのため、アイデアやフィードバックを歓迎します。

## ライセンス

MIT License（予定）

## 名前の由来

**壺（Tsubo）** には、深い意味が込められています：

> **壺は、アプリケーション全体を表す容器である。**
>
> **ドメインは、壺の中に入れる固体オブジェクトである。**
>
> 1つの壺の中には、複数の固体オブジェクト（ドメイン）が入る。
> 各固体オブジェクトは独立したマイクロサービスとなる。
>
> 人は**どの固体オブジェクト（ドメイン）を壺に入れるか**を決め、
> 各オブジェクトの**インターフェース（Contract）**を定義する。
> オブジェクトの内部構造がどう作用するかは**AIが決める**。
>
> **1つの壺（アプリケーション）= 複数の固体オブジェクト（ドメイン/マイクロサービス）の集合**

**カプセル化の新しい意味:**
- 伝統的なカプセル化: 内部実装を外部から隠蔽
- Tsubo のカプセル化: **人間から実装の詳細を隠蔽**、AIに任せる
- **固体オブジェクトの独立性**: 各ドメイン（マイクロサービス）は壺の中で独立して存在

壺の中の固体オブジェクト（ドメイン）の集合が、堅牢なアプリケーションを作ります。

## 開発原則

Tsubo は以下の原則に基づいて開発されます：

### 🐳 Docker First
- すべての実装は仮想環境（Docker）で実行
- ローカル環境への影響ゼロ
- 再現性の保証

### 🤐 質疑のタイミング
- **実装前（Contract段階）**: 曖昧性の排除、セキュリティ確認
- **実装中**: 質疑なし、AIが自律的に実装

### 📝 Contract is Everything
- Contract は唯一の真実（Single Source of Truth）
- 人間は「何をすべきか」を定義
- AI は「どう実装するか」を決定

詳細は [DEVELOPMENT_PRINCIPLES.md](./docs/DEVELOPMENT_PRINCIPLES.md) を参照。

---

**Status:** ✅ **Unified CLI Complete**
**Version:** 0.5.0
**Latest Achievement:** 統一 CLI 実装完了（`tsubo` コマンドですべての操作が可能）

**実装済み:**
- ✅ **tsubo CLI** - 統一コマンドラインインターフェース
  - `tsubo new` - サービステンプレート生成
  - `tsubo build` - プラン生成・AI 実装（plan + execute 統合）
  - `tsubo verify` - Contract 検証・テスト実行
  - `tsubo run` - サービス起動
  - 並行数制御（`--concurrency`）
  - トポロジカルソートによる複数 Wave 対応
- ✅ tsubo-plan (Go) - 実装プランニングツール（後方互換）
- ✅ tsubo-execute (Go) - プロンプト生成・Claude API 自動実行（後方互換）
- ✅ 壺（アプリケーション全体）: tsubo-todo-app
- ✅ 2つの固体オブジェクト（AI が並列実装）:
  - user-service (Wave 0) - ユーザー管理
  - todo-service (Wave 1) - TODO管理
- ✅ ドメイン間連携（service-to-service通信）
- ✅ Docker Compose によるオーケストレーション
- ✅ 100% Contract 準拠
- ✅ 統合テスト完備

**`tsubo` コマンド一つで、サービス作成から実装、検証、起動までのすべてが完結します！** 🎉
