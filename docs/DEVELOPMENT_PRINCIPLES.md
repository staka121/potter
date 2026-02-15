# Tsubo 開発原則

## 🎯 核心原則

### 1. 仮想環境必須（Docker First）

**すべての実装は仮想環境（Docker）で実行する。**

#### 理由
- ✅ **ローカル環境への影響ゼロ**: 依存関係、ポート、プロセスがホストに影響しない
- ✅ **再現性の保証**: どの環境でも同じ動作
- ✅ **依存関係の隔離**: ライブラリやツールのバージョンを固定
- ✅ **クリーンアップが容易**: `docker-compose down` で完全に削除

#### 実装ルール
```bash
# すべてのサービスは Docker Compose で起動
docker-compose up -d

# 開発・テスト・実行はすべて Docker 内で完結
docker-compose exec service-name <command>

# 終了時は完全にクリーンアップ
docker-compose down
```

### 2. 質疑のタイミング

**実装前（Contract段階）のみ質疑を行い、実装中は自律的に進める。**

#### 質疑が許される場面

**実装開始前:**
- ✅ **Contract定義の曖昧性を排除する質問**
  - 例: "このフィールドが `null` になるのはどんな場合ですか？"
  - 例: "同時実行時の振る舞いは？"
  - 例: "エラー時のロールバックは必要ですか？"

- ✅ **セキュリティ的欠陥の指摘**
  - 例: "このエンドポイントは認証が必要では？"
  - 例: "パスワードのハッシュ化は？"
  - 例: "SQLインジェクションのリスクがあります"

- ✅ **ローカル環境への影響確認**
  - 例: "ポート8080を使用しますが、問題ありませんか？"
  - 例: "新しいDockerイメージをpullしますが、よろしいですか？"

**実装中:**
- ❌ **実装の詳細に関する質問は行わない**
  - 例: "この処理はどのパターンで実装しますか？" → AI が自律的に決定
  - 例: "エラーハンドリングはどうしますか？" → Contract に従って実装
  - 例: "ファイル構成はどうしますか？" → ベストプラクティスに従う

### 3. Contract is Everything（契約がすべて）

**Contract は実装の唯一の真実（Single Source of Truth）である。**

#### Contract に含まれるべき情報

```yaml
service:
  context:
    purpose: |
      このサービスの目的とビジネス上の意図
    responsibilities:
      - 具体的な責務1
      - 具体的な責務2
    constraints:
      - 制約1
      - 制約2

api:
  endpoints:
    - semantics:
        intent: この操作の意図
        behavior:
          success: 正常時の振る舞い
          edge_cases:
            - case: エッジケースの説明
              response: 期待されるレスポンス
              reason: なぜそうすべきか
```

#### Contract が曖昧な場合の対処

**実装前に質問する:**
- "このフィールドの `null` は何を意味しますか？"
- "同時実行時の振る舞いは？"
- "このステータスコードを返す条件は？"

**実装中は推測しない:**
- Contract に書かれていないことは実装しない
- 過度な一般化や抽象化を避ける
- 必要最小限の実装に留める

## 🏗️ 開発フロー

### Phase 1: Contract 定義（人間の仕事）

```
1. ビジネス要件を整理
2. Contract を YAML で定義
   - API スキーマ
   - セマンティック情報
   - エッジケースの期待動作
3. レビュー & 曖昧性の排除
4. Contract を確定
```

### Phase 2: 実装（AI の仕事）

```
1. Contract を読み込む
2. Docker 環境をセットアップ
3. 実装を自律的に進める
   - エンドポイントの実装
   - エラーハンドリング
   - バリデーション
   - テスト
4. Contract との適合性を検証
5. 完了報告
```

### Phase 3: 検証（自動）

```
1. Contract から自動生成されたテストを実行
2. すべてのエンドポイントとエッジケースを確認
3. 結果をレポート
```

## 🐳 Docker ベストプラクティス

### 1. マルチステージビルド

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o service .

# Runtime stage
FROM alpine:latest
COPY --from=builder /build/service .
CMD ["./service"]
```

### 2. docker-compose.yml の構造

```yaml
services:
  service-name:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: tsubo-service-name
    ports:
      - "8080:8080"
    environment:
      - ENV_VAR=value
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3
```

### 3. ポート管理

```
- 各サービスは固有のホストポートを使用
- コンテナ内は標準ポート（例: 8080）を使用
- ポート競合を避けるため、8080-8099 の範囲を使用
```

### 4. ボリューム管理

```yaml
volumes:
  # 永続化が必要なデータのみボリューム化
  - ./data:/app/data

# 開発中のコード変更を反映する場合
  - .:/app
```

## 📝 ファイル構成

### 各サービスの標準構成

```
service-name/
├── Dockerfile           # Docker イメージ定義
├── docker-compose.yml   # Docker Compose 設定
├── .dockerignore        # Docker ビルド除外設定
├── main.go              # エントリーポイント
├── handler.go           # HTTPハンドラー
├── model.go             # データモデル
├── storage.go           # データストレージ
├── go.mod               # 依存関係
├── test.sh              # Contract テストスクリプト
└── README.md            # ドキュメント
```

## ✅ チェックリスト

### Contract 定義時
- [ ] すべてのエンドポイントに `semantics.intent` が記述されている
- [ ] エッジケースと期待される振る舞いが明記されている
- [ ] 型定義が完全である（`null` 許容も含む）
- [ ] セキュリティ要件が明確である
- [ ] パフォーマンス要件が定義されている

### 実装時
- [ ] Dockerfile を作成した
- [ ] docker-compose.yml を作成した
- [ ] .dockerignore を作成した
- [ ] 環境変数で設定を外部化した
- [ ] ヘルスチェックを実装した
- [ ] Contract テストスクリプトを作成した
- [ ] すべてのエッジケースをテストした

### レビュー時
- [ ] Docker コンテナ内で動作確認した
- [ ] Contract との適合性を確認した
- [ ] すべてのテストが pass した
- [ ] ローカル環境への影響がないことを確認した
- [ ] README が更新されている

## 🚫 アンチパターン

### やってはいけないこと

❌ **ローカルに直接依存関係をインストール**
```bash
# NG
go install github.com/some/tool

# OK
docker-compose run service go install github.com/some/tool
```

❌ **Contract にない機能を実装**
```go
// NG: Contract にない「優先度」フィールドを追加
type Todo struct {
    Priority int // Contract にない！
}

// OK: Contract 通りに実装
type Todo struct {
    // Contract で定義されたフィールドのみ
}
```

❌ **実装中に仕様を質問**
```
NG: "このエラーメッセージはどうしますか？"
    → Contract に定義されているはず

OK: "Contract のこの部分が曖昧です"
    → 実装前に質問する
```

❌ **過度な一般化・抽象化**
```go
// NG: 使われないかもしれない汎用的な仕組み
type GenericRepository[T any] interface {
    // 複雑な汎用インターフェース
}

// OK: 必要最小限の実装
type TodoStorage interface {
    Create(todo *Todo) error
    Get(id string) (*Todo, error)
    // Contract で必要な操作のみ
}
```

## 📊 成功の指標

### 質疑の削減
- 実装前の質疑: 3-5個（適切）
- 実装中の質疑: 0個（目標）

### 環境の隔離
- ローカル環境への変更: 0件
- Docker コンテナのみで完結: 100%

### Contract との適合性
- Contract テストの pass 率: 100%
- エッジケースのカバー率: 100%

---

**これらの原則に従うことで、Tsubo の哲学を体現した開発が実現できます。**

> "人間は Contract を決め、AI は実装を決める。環境は Docker が隔離する。"
