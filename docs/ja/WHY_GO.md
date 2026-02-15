# なぜ Go言語なのか

## TL;DR

**Tsubo は Go言語をターゲット言語として選択しました。**
理由はシンプルです：**Go の「誰が書いても同じコードになる」という特性が、AIによるハルシネーションを劇的に低減するから。**

## Go言語の設計哲学とAI駆動開発の相性

### 1. 単一の正解パターン（One Obvious Way）

**Go の設計哲学:**
> "There should be one-- and preferably only one --obvious way to do it."

**AI駆動開発にとっての意味:**
- AIが「どう書くべきか」で迷わない
- 選択肢が少ない = ハルシネーションが少ない
- 生成されるコードの一貫性が高い

**例: エラーハンドリング**

Go では、エラーハンドリングのパターンはほぼ1つ：
```go
result, err := doSomething()
if err != nil {
    return nil, fmt.Errorf("failed to do something: %w", err)
}
```

他の言語（例: TypeScript）では複数の方法がある：
```typescript
// Option 1: try-catch
try {
    const result = await doSomething();
} catch (error) {
    // ...
}

// Option 2: Promise .catch()
doSomething().catch(error => {
    // ...
});

// Option 3: async/await with try-catch
// Option 4: Result型パターン
// etc.
```

AIは選択肢が多いと迷い、一貫性のないコードを生成しやすい。

### 2. シンプルな言語仕様

**Go の言語仕様:**
- キーワードは **25個のみ**（C++: 95個、Rust: 50個以上）
- 例外処理なし（エラー値を返す）
- ジェネリクスは最近追加されたが、シンプル
- クラスなし（構造体とインターフェース）
- 継承なし（コンポジション）

**AI駆動開発にとっての意味:**
- AIが理解すべき概念が少ない
- 複雑な言語機能による混乱が少ない
- プロンプトコンテキストがシンプルになる

### 3. 標準フォーマット（gofmt）

**Go の特徴:**
- `gofmt` が標準で提供される
- インデント、改行、スペースなどが自動的に統一される
- **コードスタイルに関する議論がゼロ**

**AI駆動開発にとっての意味:**
- AIが生成したコードも自動的に統一される
- フォーマットに関するハルシネーションがない
- レビューがコードの内容に集中できる

### 4. 明示的なエラーハンドリング

**Go の設計:**
```go
func createUser(email string) (*User, error) {
    if !isValidEmail(email) {
        return nil, errors.New("invalid email")
    }

    user, err := db.Insert(email)
    if err != nil {
        return nil, fmt.Errorf("failed to insert: %w", err)
    }

    return user, nil
}
```

エラーは**値として**扱われ、明示的にチェックされる。

**AI駆動開発にとっての意味:**
- エラーケースが見落とされにくい
- AIが「エラーを返すべき場所」を明確に理解できる
- 例外による暗黙的な制御フローがない

### 5. 標準ライブラリの充実

**Go の標準ライブラリ:**
- `net/http`: HTTPサーバー/クライアント
- `encoding/json`: JSON エンコード/デコード
- `database/sql`: データベース接続
- `testing`: テストフレームワーク
- `context`: コンテキスト管理

**AI駆動開発にとっての意味:**
- 外部依存が少ない = AIが混乱しにくい
- 標準的なパターンが確立されている
- ドキュメントが豊富

## 他の言語との比較

### Rust の課題

**良い点:**
- ✅ 型安全性が非常に高い
- ✅ パフォーマンスが優れている
- ✅ メモリ安全性

**AI駆動開発の観点での課題:**
- ❌ ライフタイム、所有権、借用などの複雑な概念
- ❌ 同じことを実現する方法が複数ある（`String` vs `&str`, `Vec` vs `&[T]`, etc.）
- ❌ AIがハルシネーションを起こしやすい
- ❌ コンパイルエラーが多く、修正が難しい

**例: 文字列の扱い**
```rust
// AIは以下のどれを使うべきか迷う
fn process1(s: String) { }        // 所有権を奪う
fn process2(s: &str) { }          // 借用（文字列スライス）
fn process3(s: &String) { }       // 借用（String への参照）
```

Goでは：
```go
func process(s string) { }  // これだけ
```

### TypeScript の課題

**良い点:**
- ✅ Web/Node.js エコシステムが豊富
- ✅ 型システムが柔軟

**AI駆動開発の観点での課題:**
- ❌ JavaScriptの柔軟性が仇となる（書き方が多様）
- ❌ 設定が複雑（tsconfig.json, webpack, vite, etc.）
- ❌ `any` 型による型安全性の抜け穴
- ❌ Promise/async/await のエラーハンドリングが複雑

### Python の課題

**良い点:**
- ✅ シンプルで読みやすい
- ✅ AI/ML エコシステムが豊富

**AI駆動開発の観点での課題:**
- ❌ 動的型付け（型ヒントは完全ではない）
- ❌ パフォーマンスが低い（マイクロサービスには不向きな場合も）
- ❌ 実行時エラーが多い

## Go の「シンプルさ」がもたらす効果

### 効果1: ハルシネーションの削減

**測定可能な指標:**
- コンパイルエラーの発生率
- 実行時エラーの発生率
- レビューでの指摘事項の数

**Go のシンプルさによる効果:**
- 選択肢が少ない → AIが正しいパターンを選びやすい
- 明示的なエラーハンドリング → エラーケースの見落としが少ない
- 型システム → 型関連のエラーが少ない

### 効果2: 一貫性のあるコードベース

**人間が書いた Go コード:**
```go
func GetUser(id string) (*User, error) {
    user, err := db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

**AIが生成した Go コード:**
```go
func GetUser(id string) (*User, error) {
    user, err := db.Query("SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

**→ ほぼ同じコードになる！**

### 効果3: レビューの効率化

- コードスタイルの議論が不要（gofmt で統一）
- パターンが一貫しているため、レビューが速い
- AIが生成したコードでも、人間が書いたコードと区別がつかない

## Tsubo におけるGo言語の役割

### 1. Orchestrator（オーケストレーター）
- **言語: Go**
- 理由: 並行処理（goroutine）、シンプルさ

### 2. Validator（検証エンジン）
- **言語: Go**（当初はRustを検討したが、Go に統一）
- 理由: コードベースの統一、十分な型安全性

### 3. 生成されるサービス
- **推奨言語: Go**
- 理由: AIによるハルシネーション最小化
- 将来的には TypeScript, Python もサポート予定

### 4. CLI ツール
- **言語: Go**
- 理由: シングルバイナリで配布可能、クロスプラットフォーム

## Go の制約と対処法

### 制約1: ジェネリクスが弱い（Go 1.18 以降は改善）
**対処法:**
- 必要な箇所でのみジェネリクスを使用
- コード生成で補完

### 制約2: パフォーマンスは Rust に劣る
**対処法:**
- ほとんどのマイクロサービスでは Go で十分
- 本当にパフォーマンスが必要な箇所は個別に最適化

### 制約3: 依存性注入が言語レベルでサポートされていない
**対処法:**
- シンプルな構造体ベースの DI を採用
- 複雑な DI フレームワークは使わない（AIが混乱する）

## 結論

**Tsubo が Go言語を選択する理由:**

1. ✅ **誰が書いても同じコードになる** → AIのハルシネーション最小化
2. ✅ **シンプルな言語仕様** → AIが理解しやすい
3. ✅ **明示的なエラーハンドリング** → エラーケースの見落とし防止
4. ✅ **標準フォーマット（gofmt）** → コードスタイルの統一
5. ✅ **マイクロサービスのエコシステム** → 実用的
6. ✅ **統一されたコードベース** → Orchestrator も Validator も Go

**Tsubo の哲学:**
> "AIによるハルシネーションを最小化するために、最もシンプルで一貫性のある言語を選ぶ。
> それが Go言語である。"

---

## 参考資料

- [Go Proverbs](https://go-proverbs.github.io/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
