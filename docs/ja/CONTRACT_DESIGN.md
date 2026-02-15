# Contract 設計思想

## Contract の本質的な役割

Tsubo における Contract は、従来のAPI定義を超えた、**3つの役割を持つ Single Source of Truth** です：

### 1. 人間向け：サービス間の合意仕様書
- チーム間のコミュニケーションツール
- システム全体の理解を助ける設計ドキュメント
- 変更管理の基準

### 2. AI向け：プロンプトコンテキストとしての指示書
- **「このサービスが何をすべきか」の明確な指示**
- 実装時にAIに注入されるコンテキスト
- ハルシネーション防止のためのガードレール

### 3. テスト向け：バリデーションの基準
- Contract Testing の自動生成
- 実装の正しさの検証
- リグレッション防止

## 既存フォーマットの限界と Tsubo の拡張

### OpenAPI / Protobuf の強み
- ✅ 型定義が明確
- ✅ エンドポイント/メソッド定義
- ✅ 広く採用されている
- ✅ ツールエコシステムが豊富

### OpenAPI / Protobuf の限界（AIコンテキストとして）
- ❌ **ビジネス上の目的**が不明確
- ❌ **振る舞いの意図**が欠けている
- ❌ **エッジケースでの期待動作**が記述できない
- ❌ **「なぜそうすべきか」の文脈**が不足

### AIがハルシネーションを起こす理由
型やエンドポイントの定義だけでは不十分。AIは以下が不明確なときに誤った実装を生成する：
- このフィールドが `null` になるのはどんな状況か？
- エラー時の振る舞いはどうあるべきか？
- このステータスコードを返すビジネス上の理由は？
- 同時実行時の整合性はどう保証するか？

## Tsubo Contract Format の設計方針

### 原則1: 既存標準との互換性
- OpenAPI や Protobuf を**完全に置き換えるのではなく、拡張する**
- 既存ツールとの連携を可能にする
- 段階的な移行を可能にする

### 原則2: セマンティック情報をファーストクラスに
- ビジネスコンテキスト、意図、制約を明示的に記述
- AIが理解しやすい構造化された形式
- 人間が読んで理解しやすい

### 原則3: 階層的な詳細度
- 必須：基本的なAPI定義（OpenAPI互換）
- 推奨：セマンティック情報（ビジネスコンテキスト）
- オプション：詳細な振る舞い定義、例、制約

## Tsubo Contract Format の構造

### 基本構造

```yaml
# user-service.tsubo.yaml
version: "1.0"
service:
  name: user-service
  description: ユーザー管理サービス

  # ビジネスコンテキスト（AI向け）
  context:
    purpose: |
      このサービスは、アプリケーションのユーザーライフサイクル全体を管理します。
      ユーザーの作成、認証、プロフィール更新、削除を担当します。

    domain: authentication-and-authorization

    responsibilities:
      - ユーザーの CRUD 操作
      - メールアドレスの一意性保証
      - パスワードのハッシュ化
      - ユーザー削除時の関連データのクリーンアップ

    constraints:
      - メールアドレスは必ず小文字に正規化する
      - 削除されたユーザーのメールアドレスは再利用不可
      - パスワードは bcrypt でハッシュ化（cost=12）

# OpenAPI 3.x ベースのエンドポイント定義
api:
  version: "1.0.0"
  base_path: /api/v1

  endpoints:
    - id: create_user
      method: POST
      path: /users

      # 従来の定義
      request:
        content_type: application/json
        schema:
          type: object
          required: [email, password, name]
          properties:
            email: {type: string, format: email}
            password: {type: string, minLength: 8}
            name: {type: string}

      response:
        200:
          schema: {$ref: "#/types/User"}
        400:
          schema: {$ref: "#/types/Error"}
        409:
          schema: {$ref: "#/types/Error"}

      # Tsubo 拡張: セマンティック情報
      semantics:
        intent: |
          新しいユーザーをシステムに登録します。
          メールアドレスの一意性をチェックし、パスワードを安全にハッシュ化します。

        behavior:
          success: |
            - メールアドレスを小文字に正規化
            - パスワードを bcrypt でハッシュ化（cost=12）
            - データベースに保存
            - 作成されたユーザーオブジェクトを返す

          edge_cases:
            - case: メールアドレスが既に存在する
              response: 409 Conflict
              body: {error: "Email already exists"}
              reason: ユーザーは一意でなければならない

            - case: パスワードが弱い（8文字未満）
              response: 400 Bad Request
              body: {error: "Password too weak"}
              reason: セキュリティ要件

            - case: メールアドレスの形式が不正
              response: 400 Bad Request
              body: {error: "Invalid email format"}

        examples:
          - name: 正常なユーザー作成
            request:
              email: "user@example.com"
              password: "SecurePass123!"
              name: "John Doe"
            response:
              status: 200
              body:
                id: "usr_123456"
                email: "user@example.com"
                name: "John Doe"
                created_at: "2026-02-15T10:00:00Z"

          - name: 重複メールアドレス
            request:
              email: "existing@example.com"
              password: "SecurePass123!"
              name: "Jane Doe"
            response:
              status: 409
              body:
                error: "Email already exists"

# 型定義
types:
  User:
    description: システム内のユーザーを表現します
    properties:
      id:
        type: string
        format: uuid
        description: ユーザーの一意識別子
        immutable: true
      email:
        type: string
        format: email
        description: ユーザーのメールアドレス（小文字正規化済み）
        unique: true
      name:
        type: string
        description: ユーザーの表示名
      created_at:
        type: string
        format: date-time
        description: ユーザー作成日時
        immutable: true

# 依存関係
dependencies:
  services:
    - name: auth-service
      reason: ユーザー作成時に初期認証トークンを発行
      endpoints: ["/tokens"]

    - name: email-service
      reason: ウェルカムメール送信
      endpoints: ["/send"]
      optional: true  # メール送信失敗してもユーザー作成は成功

  databases:
    - name: user-db
      type: postgresql
      tables: [users, user_profiles]

# テスト定義（Contract Testing）
tests:
  contract:
    - name: ユーザー作成の契約テスト
      given: 有効なユーザーデータ
      when: POST /users
      then:
        status: 200
        body_schema: {$ref: "#/types/User"}
        invariants:
          - email は小文字である
          - id は UUID 形式である
          - created_at は現在時刻に近い

    - name: 重複メールアドレスの契約テスト
      given: 既に存在するメールアドレス
      when: POST /users
      then:
        status: 409
        body.error: "Email already exists"

# パフォーマンス要件
performance:
  latency:
    p50: 100ms
    p95: 300ms
    p99: 500ms

  throughput:
    target: 1000 req/sec

  concurrency:
    handling: |
      同一メールアドレスでの同時作成リクエストは、
      データベースの UNIQUE 制約により1つのみ成功する。
      先に完了したトランザクションが成功し、
      他は 409 Conflict を返す。
```

## セマンティック情報の構造

### 1. Context（サービス全体のコンテキスト）
```yaml
context:
  purpose: |
    このサービスの存在理由とビジネス上の目的
  domain: ドメイン分類
  responsibilities: [責務のリスト]
  constraints: [制約のリスト]
```

### 2. Semantics（エンドポイントごとのセマンティクス）
```yaml
semantics:
  intent: この操作の意図
  behavior:
    success: 正常時の振る舞い
    edge_cases:
      - case: エッジケースの説明
        response: 期待されるレスポンス
        reason: なぜそうすべきか
  examples: 具体例のリスト
```

### 3. Dependencies（依存関係の明示）
```yaml
dependencies:
  services:
    - name: サービス名
      reason: なぜ依存するか
      endpoints: 使用するエンドポイント
      optional: オプショナルかどうか
```

## AI へのコンテキスト注入

Contract を AI に注入する際の形式：

```markdown
# サービス実装指示

あなたは `user-service` を実装しています。

## サービスの目的
{context.purpose}

## 責務
{context.responsibilities}

## 制約
{context.constraints}

## 実装すべきエンドポイント: POST /users

### 意図
{semantics.intent}

### 正常時の振る舞い
{semantics.behavior.success}

### エッジケース
{semantics.behavior.edge_cases}

### 例
{semantics.examples}

---

上記の契約に従って、実装してください。
型定義、エラーハンドリング、エッジケースの処理を含めてください。
```

## Contract から生成されるもの

### 1. AI プロンプト
- サービス実装時に注入されるコンテキスト

### 2. Contract Tests
- Pact スタイルの契約テスト
- Consumer/Provider 両側のテスト

### 3. OpenAPI スキーマ
- 既存ツールとの互換性のため
- API ドキュメント生成

### 4. Mock Server
- 開発・テスト用のモックサーバー
- 契約に基づいた応答

### 5. クライアントコード
- 型安全なクライアントライブラリ
- 各言語向けの SDK

## 段階的な採用戦略

### Phase 1: 最小限の Contract
```yaml
service:
  name: user-service
api:
  endpoints: [...]  # 基本的な定義のみ
```

### Phase 2: セマンティック情報の追加
```yaml
service:
  name: user-service
  context: {purpose: "..."}
api:
  endpoints:
    - semantics: {intent: "..."}
```

### Phase 3: 完全な Contract
```yaml
# すべての要素を含む
service: ...
api: ...
types: ...
dependencies: ...
tests: ...
performance: ...
```

## まとめ

Tsubo Contract は：
- ✅ OpenAPI/Protobuf の良い部分を継承
- ✅ セマンティック情報を追加してAIのハルシネーションを防ぐ
- ✅ 1つの定義から複数の成果物を生成（DRY原則）
- ✅ 段階的な採用が可能
- ✅ 既存ツールとの互換性を維持
