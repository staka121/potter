# Potter と Kubernetes

> **Potter のマルチ環境哲学：同じ定義で、異なるターゲット**

Potter は、単一の Tsubo 定義からローカル開発と本番デプロイの両方をサポートするよう設計されており、Kubernetes を主要な本番ターゲットとしています。

## 概要

Potter は、ローカル開発のシンプルさを維持しながら、Kubernetes とシームレスに統合します：

- 🏠 **ローカル開発**: Docker Compose + gateway-service
- ☁️ **本番環境**: Kubernetes + Ingress
- 📝 **単一ソース**: 両環境に対応する1つの `.tsubo.yaml` ファイル

## アーキテクチャ比較

### ローカル開発 (Docker Compose)

```
┌─────────────────────────────────────────┐
│  開発者のラップトップ                        │
│                                          │
│  potter run app.tsubo.yaml               │
│         ↓                                │
│  ┌──────────────────┐                    │
│  │ gateway-service  │ :8080              │
│  │   (自動起動)      │                    │
│  └────────┬─────────┘                    │
│           │                              │
│     ┌─────┴─────┐                        │
│     ▼           ▼                        │
│  ┌────────┐ ┌────────┐                   │
│  │  user  │ │  todo  │                   │
│  │ :8084  │ │ :8083  │                   │
│  └────────┘ └────────┘                   │
│                                          │
│  アクセス: http://localhost:8080          │
└─────────────────────────────────────────┘
```

**特徴:**
- ✅ シンプル: 1つのコマンドですべて起動
- ✅ 高速: クラスタのセットアップ不要
- ✅ 統一: 単一のエントリーポイント (localhost:8080)
- ✅ 隔離: Docker コンテナ内で実行

### 本番環境 (Kubernetes)

```
┌─────────────────────────────────────────┐
│  Kubernetes クラスタ                      │
│                                          │
│  potter deploy generate --ingress        │
│         ↓                                │
│  ┌──────────────────┐                    │
│  │     Ingress      │                    │
│  │ (nginx/traefik)  │                    │
│  └────────┬─────────┘                    │
│           │                              │
│     ┌─────┴─────┐                        │
│     ▼           ▼                        │
│  ┌────────┐ ┌────────┐                   │
│  │ user-  │ │ todo-  │                   │
│  │ service│ │ service│                   │
│  │  Pods  │ │  Pods  │                   │
│  └────────┘ └────────┘                   │
│                                          │
│  アクセス: https://api.example.com        │
└─────────────────────────────────────────┘
```

**特徴:**
- ✅ スケーラブル: HPA による自動スケーリング
- ✅ 回復力: セルフヒーリングとローリングアップデート
- ✅ 本番グレード: 実績ある Ingress Controller
- ✅ セキュア: TLS 終端、ネットワークポリシー

## ゲートウェイ比較

### gateway-service (Docker Compose)

**概要:**
- リバースプロキシとして機能するカスタム Go アプリケーション
- Potter によって自動生成・自動起動
- バックエンドサービスへトラフィックをルーティング

**ルーティングロジック:**
```go
/api/v1/users/* → user-service:8084
/api/v1/todos/* → todo-service:8083
/health         → ゲートウェイのヘルスチェック
```

**使用場面:**
- `potter run` によるローカル開発
- シンプルなデプロイメントシナリオ
- Docker Compose 環境

**メリット:**
- 設定不要
- Docker が動けばどこでも動作
- シンプルなデバッグ

**デメリット:**
- メンテナンスが必要なカスタムコード
- スケーラビリティの限界
- TLS/認証の組み込みサポートなし

### Ingress (Kubernetes)

**概要:**
- Kubernetes ネイティブリソース
- 実績ある Ingress Controller (nginx, Traefik など) を使用
- Tsubo 定義から Potter が生成

**ルーティングロジック:**
```yaml
/api/v1/users(/|$)(.*) → user-service:80
/api/v1/todos(/|$)(.*) → todo-service:80
```

**使用場面:**
- Kubernetes デプロイメント
- 本番環境
- 高度な機能が必要な場合 (TLS, 認証, レート制限)

**メリット:**
- 本番グレード
- 豊富なエコシステム (cert-manager, OAuth など)
- ネイティブ K8s 統合
- 大規模で実証済み

**デメリット:**
- Kubernetes クラスタが必要
- 初期セットアップがやや複雑

## マルチ環境ワークフロー

### 1. 一度定義

Tsubo 定義を作成：

```yaml
# app.tsubo.yaml
tsubo:
  name: my-app

objects:
  - name: user-service
    contract: ./user-service.object.yaml
    runtime:
      port: 8080
      health_check: /health
    dependencies: []

  - name: todo-service
    contract: ./todo-service.object.yaml
    runtime:
      port: 8080
      health_check: /health
    dependencies:
      - user-service
```

### 2. ローカルで開発

```bash
# AI が実装を生成
potter build app.tsubo.yaml

# Docker Compose + gateway-service で起動
potter run app.tsubo.yaml

# 統一されたエンドポイントでアクセス
curl http://localhost:8080/api/v1/users
curl http://localhost:8080/api/v1/todos
```

### 3. Kubernetes にデプロイ

```bash
# Ingress 付きの K8s マニフェストを生成
potter deploy generate \
  --namespace production \
  --ingress-host api.example.com \
  --registry docker.io/myorg \
  --tag v1.0.0 \
  app.tsubo.yaml

# クラスタに適用
kubectl apply -f k8s/

# Ingress 経由でアクセス
curl https://api.example.com/api/v1/users
```

## デプロイオプション

### 基本デプロイ

デフォルト設定でマニフェストを生成：

```bash
potter deploy generate app.tsubo.yaml
```

生成されるもの:
- Namespace
- Deployments (各1レプリカ)
- Services (ClusterIP)
- Ingress (nginx, デフォルトホスト)

### 本番デプロイ

本番環境向けの完全な設定：

```bash
potter deploy generate \
  --namespace production \
  --ingress-host api.prod.example.com \
  --ingress-class nginx \
  --registry gcr.io/my-project \
  --tag v1.2.3 \
  --replicas 3 \
  --output k8s-prod \
  app.tsubo.yaml
```

### Ingress なし

Kubernetes で gateway-service を使用（非推奨）：

```bash
potter deploy generate \
  --ingress=false \
  app.tsubo.yaml
```

## 生成されるリソース

Potter は標準的な Kubernetes リソースを生成します：

### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: potter-todo
  labels:
    app.kubernetes.io/managed-by: potter
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: user-service
        image: docker.io/myorg/user-service:v1.0.0
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
        env:
        - name: TODO_SERVICE_URL
          value: "http://todo-service.potter-todo.svc.cluster.local"
```

**主要機能:**
- `health_check` から自動生成されるヘルスプローブ
- サービスの依存関係 → 環境変数
- リソース制限 (requests/limits)
- 標準的な K8s ラベル

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: user-service
```

### Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: "/$2"
    nginx.ingress.kubernetes.io/use-regex: "true"
spec:
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /api/v1/users(/|$)(.*)
        backend:
          service:
            name: user-service
            port:
              number: 80
      - path: /api/v1/todos(/|$)(.*)
        backend:
          service:
            name: todo-service
            port:
              number: 80
```

**主要機能:**
- サービス名からの自動パス推論
- URL リライトを使用した正規表現ベースのルーティング
- 設定可能な Ingress Controller (nginx, Traefik など)
- オプションの TLS 設定

## 設計哲学

### 1. Contract 駆動の K8s マニフェスト

Potter は Contract 定義から K8s リソースを生成します：

```
Contract 定義 (人間/AI が読める)
           ↓
    Tsubo パーサー
           ↓
   K8s ジェネレータ
           ↓
標準的な K8s マニフェスト
```

これにより保証されること:
- ✅ 唯一の情報源
- ✅ 環境間の一貫性
- ✅ バージョン管理された設定
- ✅ AI が理解・変更可能

### 2. ゲートウェイの抽象化

Potter はゲートウェイの概念を抽象化します：

```
Tsubo 定義
      ↓
   [Potter]
      ↓
   ├─→ Docker Compose → gateway-service
   └─→ Kubernetes     → Ingress
```

**メリット:**
- 開発者はインフラではなくサービスで考える
- 同じルーティングロジック、異なる実装
- 各環境に最適なゲートウェイ

### 3. 段階的な拡張

シンプルに始めて、必要に応じてスケール：

```
1日目: potter run (Docker Compose)
         ↓
1週目: さらにサービスを追加
         ↓
1ヶ月目: potter deploy generate (K8s staging)
         ↓
3ヶ月目: HPA、TLS、監視を備えた本番環境
```

### 4. 設定より規約

Potter は賢いデフォルトを提供します：

- サービス名 → API パスの推論
  - `user-service` → `/api/v1/users`
  - `product-service` → `/api/v1/products`
- ヘルスチェック → liveness/readiness プローブ
- 依存関係 → 環境変数
- 標準的なポートとプロトコル

## ベストプラクティス

### 1. 環境の等価性

開発環境と本番環境を可能な限り似た状態に保つ：

```bash
# 同じ Tsubo 定義を使用
potter run app.tsubo.yaml              # ローカル
potter deploy generate app.tsubo.yaml  # 本番
```

### 2. すべてをバージョン管理

```
my-app/
├── app.tsubo.yaml          # アプリケーション定義
├── user-service.object.yaml # Contract 定義
├── todo-service.object.yaml
└── k8s/                    # 生成物 (gitignore)
    ├── namespace.yaml
    ├── deployment-*.yaml
    ├── service-*.yaml
    └── ingress.yaml
```

`.gitignore` に追加：
```
k8s/
```

CI/CD でマニフェストを再生成：
```bash
potter deploy generate --tag $GIT_SHA app.tsubo.yaml
```

### 3. ブランチではなくフィーチャーフラグを使用

```yaml
# app.tsubo.yaml
objects:
  - name: new-feature-service
    contract: ./new-feature.object.yaml
    # 最初に dev/staging にデプロイ
```

### 4. 段階的なロールアウト

```bash
# ステージ1: Dev 環境
potter deploy generate \
  --namespace dev \
  --ingress-host api.dev.example.com \
  app.tsubo.yaml

# ステージ2: Staging
potter deploy generate \
  --namespace staging \
  --ingress-host api.staging.example.com \
  --replicas 2 \
  app.tsubo.yaml

# ステージ3: 本番
potter deploy generate \
  --namespace production \
  --ingress-host api.example.com \
  --replicas 5 \
  app.tsubo.yaml
```

## 高度なトピック

### TLS 設定

```bash
# TLS プレースホルダー付きで生成
potter deploy generate \
  --ingress-host api.example.com \
  app.tsubo.yaml

# cert-manager アノテーションを手動で追加
# (将来: --ingress-tls フラグ)
```

### カスタム Ingress アノテーション

生成された `ingress.yaml` を編集：

```yaml
metadata:
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
```

### 複数環境

出力ディレクトリを使い分け：

```bash
potter deploy generate --output k8s-dev app.tsubo.yaml
potter deploy generate --output k8s-staging app.tsubo.yaml
potter deploy generate --output k8s-prod app.tsubo.yaml
```

### Helm Chart 生成

(将来の機能拡張)

```bash
potter deploy generate --helm app.tsubo.yaml
```

## トラブルシューティング

### Ingress が動作しない

```bash
# Ingress Controller がインストールされているか確認
kubectl get pods -n ingress-nginx

# Ingress リソースを確認
kubectl describe ingress -n potter-todo

# サービスエンドポイントを確認
kubectl get endpoints -n potter-todo
```

### サービスにアクセスできない

```bash
# Pod が実行中か確認
kubectl get pods -n potter-todo

# Pod のログを確認
kubectl logs -n potter-todo deployment/user-service

# サービスを直接テスト (port-forward)
kubectl port-forward -n potter-todo svc/user-service 8080:80
curl http://localhost:8080/health
```

### gateway-service と Ingress の競合

gateway-service と Ingress の両方が表示される場合：

```bash
# --ingress 付きで再生成（デフォルト）
potter deploy generate --ingress app.tsubo.yaml

# gateway-service は自動的にスキップされます
```

## 今後の機能拡張

- [ ] `potter deploy apply` - 直接 K8s デプロイ
- [ ] Helm chart 生成
- [ ] Horizontal Pod Autoscaler (HPA) 設定
- [ ] Tsubo からの ConfigMap/Secret 管理
- [ ] Service Mesh 統合 (Istio, Linkerd)
- [ ] GitOps 統合 (ArgoCD, Flux)

## まとめ

Potter の Kubernetes 統合は、フレームワークの哲学を体現しています：

> **人間が WHAT（契約、ドメイン）を定義**
> **AI が HOW（サービスロジック）を実装**
> **Potter が WHERE（Docker Compose ⇄ Kubernetes）を橋渡し**

これにより開発者は：
- インフラではなくビジネスロジックに集中できる
- すべての環境で同じ定義を使用できる
- 段階的に Kubernetes を採用できる
- ローカル開発のシンプルさを維持できる
- 自信を持って本番環境にデプロイできる

---

**ステータス:** ✅ 本番対応
**バージョン:** 0.6.0
**最新機能:** Ingress 生成（K8s では gateway-service を置き換え）
