## 📱 FreStyleとは

## 🎯 ターゲットユーザー

**新卒ITエンジニアのためのビジネスコミュニケーション練習アプリ**

- **顧客折衝**: 障害報告、要件変更の影響説明、見積もり交渉など、顧客との難しい場面を想定したケーススタディで学べます
- **シニアエンジニアとの対話**: 設計レビューでの意見対立、コードレビューのフィードバック対応、技術負債の改善提案など、経験豊富なエンジニアとのコミュニケーションスキルを磨けます
- **実践的なシナリオ**: AIがロールプレイ相手となり、実務で遭遇する12種類のビジネスシーンを体験できます
- **5軸評価 + IT新卒向けサブ基準**: 論理的構成力（報連相の構造化）、配慮表現（敬語の正確さ）、要約力（技術説明の平易化）、提案力（エスカレーション判断力）、質問・傾聴力（要件確認の網羅性）で評価

技術スキルだけでなく、ビジネスコミュニケーション能力も身につけたい新卒エンジニアを支援します。

---

## 🌐 デプロイURL（料金の関係上サービスを停止する可能性があります）

👉 [https://normanblog.com](https://normanblog.com)

---

## 🎥 Demo（変更前）

👉 [デモ動画](https://myapp-demo-videos.s3.ap-northeast-1.amazonaws.com/Fre-Style-demo.mp4)

---

## 🧰 使用技術

<h3>Frontend</h3>
<a href="https://skillicons.dev">
  <img src="https://skillicons.dev/icons?i=react,ts,tailwind,vite&theme=light" alt="Frontend">
</a>

<h3>Backend</h3>
<a href="https://skillicons.dev">
  <img src="https://skillicons.dev/icons?i=java,spring,gradle&theme=light" alt="Backend">
</a>

<h3>Infrastructure</h3>
<a href="https://skillicons.dev">
  <img src="https://skillicons.dev/icons?i=aws,cloudflare,docker&theme=light" alt="Infrastructure">
</a>

<h3>Database</h3>
<a href="https://skillicons.dev">
  <img src="https://skillicons.dev/icons?i=mysql,dynamodb&theme=light" alt="Database">
</a>

<h3>CI/CD</h3>
<a href="https://skillicons.dev">
  <img src="https://skillicons.dev/icons?i=githubactions&theme=light" alt="CI/CD">
</a>

### AWS サービス詳細

| カテゴリ | サービス |
|---------|----------|
| **Compute** | ECS (Fargate), Lambda |
| **Networking** | API Gateway, Route53, CloudFront, ALB |
| **Database** | RDS (MariaDB), DynamoDB |
| **Storage** | S3 |
| **Auth** | Cognito |

---

## ⚙️ 主な機能
- ユーザー登録・ログイン（JWT 認証 / OIDC）
- ユーザー同士のリアルタイムチャット
- 未読メッセージカウント・既読管理（リアルタイム通知対応）
- AI アシスタントとのチャット（ビジネスコミュニケーション5軸フィードバック対応）
  - **IT新卒エンジニア向けサブ基準**：報連相の構造化・敬語の正確さ・技術説明の平易化・エスカレーション判断力・要件確認の網羅性
- シーン別フィードバックモード（会議/1on1/メール/プレゼン/商談/**コードレビュー**/**障害対応**/**日報・週報** の全8シーン）
- メッセージ言い換え提案（フォーマル版/ソフト版/簡潔版/**質問型**/**提案型** の全5パターン）
- コミュニケーションスコアカード（5軸スコア数値化・棒グラフ可視化・総合スコア表示）
  - **スコアレベルラベル**：総合スコアに応じたレベル表示（優秀/実務/基礎）・色分けプログレスバー・低スコア軸の改善ヒント
  - **スコア履歴ページ**：セッション単位の履歴一覧表示・スコア推移グラフ・前回比変動表示・セッション種別フィルタ（練習/フリー）
- ビジネスシナリオ練習モード（AIロールプレイ・ITエンジニア向け12シナリオ・3カテゴリ対応）
  - **カテゴリ別フィルタリング**：顧客折衝・シニア/上司・チーム内のタブ切り替えで絞り込み
  - **練習終了時スコアリング**：シナリオ固有の評価軸でスコア自動算出
  - **難易度・所要時間表示**：各シナリオに色分け難易度バッジ・説明テキスト・所要時間目安
- シーン別フィードバックモード（全8シーン・カテゴリ分類・新卒向けおすすめ表示・利用例付き）
- メッセージ言い換え提案（全5パターン・利用シーンヒント付き・コピーフィードバック表示）
- ホームページダッシュボード（未読バッジ・最新スコア・新規ユーザー向けおすすめアクション）
- Notion風ミニマルUIデザイン（ニュートラルグレー配色・永続サイドバーナビゲーション・未読バッジ・セカンダリパネル・モバイルオーバーレイドロワー）
- パフォーマンス計測・可視化（Micrometer + CloudWatch メトリクス・AOP処理時間計測）
- Docker Compose によるローカル開発環境（MariaDB 11 + Spring Boot）
- プロフィール編集
- Google ログイン
- GitHub Actions による自動デプロイ（S3同期 + CloudFrontキャッシュ自動無効化）
- **練習結果サマリー**：練習終了後に強み・課題をハイライト表示、改善アドバイス付き
- **弱点ベースのおすすめ練習**：スコア履歴から弱点軸を特定し、おすすめ練習シナリオを提示
- **メッセージ送信状態表示**：送信中インジケーター・入力欄無効化で操作フィードバック向上
- **チャット内メッセージ検索**：キーワード検索・マッチ件数表示・検索クリア機能
- **練習ストリーク・達成バッジ**：連続練習日数カウント・4種類の達成バッジ（初回/3日/7日/10回）
- **グローバルエラーバウンダリー**：予期せぬエラー時のフレンドリーUI・再試行ボタン

### テスト品質
- **フロントエンド252テスト**：Vitest + React Testing Library
  - Repository層テスト（6リポジトリ・29テスト）
  - Hooks層テスト（5フック・26テスト）
  - UIコンポーネントテスト（24コンポーネント・173テスト）
  - ページコンポーネントテスト（7ページ・24テスト）
- **バックエンド**：JUnit 5 + Mockito

---

## 💡 Architecture Highlights（工夫した点）

### ① WebSocket と HTTP API の構成を用途別に完全分離
- **WebSocket**：API Gateway + Lambda + DynamoDB  
- **HTTP（Rest API）**：ECS（Fargate） + Spring Boot  

リアルタイム性と低コストを優先した WebSocket と、安定稼働・複雑処理に適した HTTP API を分離し、性能・コスト・可用性の最適化を実現。

### ② JWT（HttpOnly Cookie）× Spring Security の安全な認証設計
- JWT を HttpOnly Cookie に保存（XSS 対策）
- アクセストークンの有効期間を短くしリフレッシュトークンで再発行をする
- OIDC & JWK を活用した堅牢な認証フロー
- OIDC経由でも当該アプリ経由でも同一ユーザーとして認識

### ③ CloudFront によるグローバル最適化と HTTPS 化
- 高速配信（CDN）
- OIDC と組み合わせてセキュアなフロント構成
- Cognito/OIDCログインを使用しているのでHTTPSの必須になるので採用した

### ④ クリーンアーキテクチャの適用による保守性向上
**Phase 1-3リファクタリング完了** (2026年2月)

バックエンドコードをクリーンアーキテクチャに基づいて全面リファクタリングし、保守性・テスタビリティ・可読性を大幅に向上させました。

#### 実装内容
- **Mapper層**: DTO↔Entity変換を一箇所に集約（3クラス作成）
- **UseCase層**: ビジネスロジックをController層から分離（13クラス作成）
- **依存性逆転の原則**: Controller → UseCase → Repository の明確な依存関係を確立
- **単一責任の原則**: 各クラスの責務を明確化し、1クラス1責務を徹底

#### リファクタリング対象
- **Phase 1**: 練習モード機能（PracticeScenarioService → 3 UseCases）
- **Phase 2**: AI Chat機能（AiChatSessionService/AiChatMessageService → 10 UseCases）
- **Phase 3**: ScoreCard機能（ScoreCardService → 3 UseCases + Mapper）

#### 成果
- コード行数: **+1,849行追加 / -377行削除**
- テスタビリティ向上: モック化が容易な設計に
- 日本語コメント充実: 各クラスの役割・責務を明記

```
【アーキテクチャ構成】
Controller (Presentation層) ← HTTP/WebSocketリクエスト処理
  ↓ 依存
UseCase (Application層) ← ビジネスロジック実行
  ↓ 依存
Service (Domain層) ← ドメインロジック
  ↓ 依存
Repository (Infrastructure層) ← DB操作
```

---

## 🧠 苦労した点・学び
- WebSocket を ECS で保持するか、サーバーレスにするかの検討 → コスト/工数削減/レイテンシから Lambda + APIGW に決定
- Spring Security の JWT / JWK / Cookie 設計
- ALB の TLS Termination と ECS の Backend 構成

---

## ✔ 技術選定理由（HTTP API / ECS Fargate）

1. Docker 化した Spring Boot を安定稼働させるため
   - サーバープロビジョニング不要
   - OS 管理不要

2. ALB と連携した柔軟なルーティング
   - ホストベースルーティングで[BeStyle](https://normanblog.com)にも同じロードバランサーを使用をしコスト削減をした
   - ヘルスチェックをしておりSpring Bootのactuatorでヘルスチェックのエンドポイントにアクセス

3. Blue/Green デプロイ
   - CodeDeploy と連携
   - 新バージョンのヘルスチェック後に切替
   - 即時ロールバック可能

---

## ✔ 技術選定理由（WebSocket / サーバーレス構成）

1. コスト最適化（従量課金）  
   ECS 常時稼働より大幅に低コスト。

2. 低レイテンシ & シンプルな処理  
   Lambda → DynamoDB の最短経路。

3. サーバーレスで構成統一
   - フルマネージド
   - 自動スケーリング
   - 運用負荷最小

---

## 🏗️ AWSアーキテクチャ構成図

### AWS全体構成図（変更前）
![AWSアーキテクチャ構成図](./architecture/aws/image.png)

### ユーザー同士のチャット（変更前）
![ユーザー同士のチャット](./architecture/aws/aws-architecture-chat.png)

### AIとユーザーのチャット（変更前）
![AIとユーザーのチャット](./architecture/aws/aws-architecture-ai-chat.png)

### 変更後のAWSアーキテクチャー図
![AWSアーキテクチャ構成図](./architecture/aws/AWSアーキテクチャー設計修正後.png)

### なぜアーキテクチャーを変えたのか
1. AIへのフィードバックにユーザーがより自分の性格を把握できるように複雑なクエリを実行する必要があったのでDynamoDBではサービス層が複雑になるのでRDSに変更をした
2. Lambda + API Gatewayではトラフィック量が多くなったときに捌きにくいこと
3. 機能の拡張性を踏まえたらECS一本で使用したほうがSQSなどを設定したときに工数を割くことができる

---

## 🚀 今年の目標

### 技術・資格
- AWS SAP、CKA、CKAD
- GO言語でgRPC通信でサービス間接続

### 機能拡張
- 音声チャット
- Polly による AI 音声解答可能
- SageMakerを使用をしチャットの内容をクラスタリングで感情分析をすること
- 未読、既読、プッシュ通知機能の作成でSQS、SNSを使用をする
- チャット以外にもパーソナリティーが見えるシステムを作成する

---

## 🛠 ローカル開発環境セットアップ

### バックエンド（Docker Compose）

```bash
# 1. 環境変数ファイルを作成（.env.example をコピーして値を設定）
cp .env.example .env

# 2. MariaDB + Spring Boot を起動
docker compose up -d --build

# 3. 動作確認
docker compose ps          # コンテナ状態確認
docker compose logs api    # Spring Boot ログ確認
```

MariaDB 11 がコンテナとして起動し、Spring Boot 起動時に `schema.sql` でテーブルが自動作成されます。

### フロントエンド

```bash
# 1. リポジトリをクローンして frontend ディレクトリに移動
cd frontend

# 2. 依存パッケージをインストール
npm install

# 3. 動作確認
npm run dev

# 4. Tailwind CSS が動作しない場合
npm uninstall tailwindcss
npm install -D tailwindcss@バージョン指定
npx tailwindcss init -p
```

---

## 🗄️ 本番環境DBマイグレーション

本番環境（AWS RDS）にデプロイする際、スキーマの更新が必要な場合はマイグレーションSQLを実行してください。

### マイグレーション手順

```bash
# 1. AWS RDSに接続
mysql -h <RDS_ENDPOINT> -u <DB_USER> -p <DB_NAME>

# 2. マイグレーションSQLを実行
source FreStyle/migrations/001_add_practice_mode_support.sql;

# 3. 確認
SHOW COLUMNS FROM ai_chat_sessions;
SELECT COUNT(*) FROM practice_scenarios;
```

### マイグレーション一覧

| ファイル名 | 実行日 | 内容 |
|-----------|--------|------|
| `001_add_practice_mode_support.sql` | 2026-02-12 | 練習モード機能追加（`ai_chat_sessions` に `session_type`, `scenario_id` カラム追加、`practice_scenarios` テーブル作成、初期データ12件投入） |

**注意**: マイグレーションは冪等性があり、複数回実行しても安全です（`IF NOT EXISTS`, `INSERT IGNORE` を使用）。

---

## 📄 ライセンス

MIT License
