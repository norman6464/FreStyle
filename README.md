# FreStyle - Chat App with Users & AI

チャットでのやりとりと対面での印象の違いを感じるといわれたことがある人はいると思います。  
その方たちに向けてチャットでのやり取りをAIがレビューをしてその中でギャップを埋めていくことができるのがこのFreStyleになります。

---

## 🌐 デプロイURL（料金の関係上サービスを停止する可能性があります）

👉 [https://normanblog.com](https://normanblog.com)

---

## 🎥 Demo

👉 [デモ動画](https://myapp-demo-videos.s3.ap-northeast-1.amazonaws.com/Fre-Style-demo.mp4)

---

## 🧰 使用技術

### Frontend
- React
- Tailwind CSS

### Backend
- Spring Boot
- AWS Lambda

### Infrastructure
- AWS（ECS / RDS / S3 / Route53 / DynamoDB / Lambda / Cognito / API Gateway）
- CloudFlare

### CI/CD
- GitHub Actions

### Database
- MariaDB（RDS）
- DynamoDB

---

## ⚙️ 主な機能
- ユーザー登録・ログイン（JWT 認証 / OIDC）
- ユーザー同士のリアルタイムチャット
- AI アシスタントとのチャット
- プロフィール編集
- Google ログイン
- GitHub Actions による自動デプロイ

---

## 💡 Architecture Highlights（工夫した点）

### ① WebSocket と HTTP API の構成を用途別に完全分離
- **WebSocket**：API Gateway + Lambda + DynamoDB  
- **HTTP（Rest API）**：ECS（Fargate） + Spring Boot  

リアルタイム性と低コストを優先した WebSocket と、安定稼働・複雑処理に適した HTTP API を分離し、性能・コスト・可用性の最適化を実現。

### ② CloudFormation による IaC（Infrastructure as Code）
- 環境構築を完全コード化
- 再現性・管理性・チーム開発適性を向上

### ③ JWT（HttpOnly Cookie）× Spring Security の安全な認証設計
- JWT を HttpOnly Cookie に保存（XSS 対策）
- アクセストークをReduxにリフレッシュトークンをHttpOnly Cookieに保存
- アクセストークンの有効期間を短くしリフレッシュトークンで再発行をする
- OIDC & JWK を活用した堅牢な認証フロー
- OIDC経由でも当該アプリ経由でも同じアカウントであれば同一ユーザーとして認識

### ④ CloudFront によるグローバル最適化と HTTPS 化
- 高速配信（CDN）
- OIDC と組み合わせてセキュアなフロント構成
- Route53 → CloudFront → S3 Hosting の構成

---

## 🧠 苦労した点・学び
- WebSocket を ECS で保持するか、サーバーレスにするかの検討 → コスト/接続管理/レイテンシから Lambda + APIGW に決定
- Spring Security の JWT / JWK / Cookie 設計
- ALB の TLS Termination と ECS の Backend 構成

---

## ✔ 技術選定理由（HTTP API / ECS Fargate）

1. Docker 化した Spring Boot を安定稼働させるため
   - サーバープロビジョニング不要
   - OS 管理不要
   - 24/7 常時稼働する API に最適

2. ALB と連携した柔軟なルーティング
   - パスベース
   - ヘルスチェック
   - 高可用性のロードバランシング

3. Target Tracking による自動スケーリング
   - CPU 使用率に基づきタスク数を増減
   - 高負荷時に自動スケール
   - 低負荷時にコスト最適化

4. Blue/Green デプロイでゼロダウンタイムを実現
   - CodeDeploy と連携
   - 新バージョンのヘルスチェック後に切替
   - 即時ロールバック可能

5. HTTP API は常時稼働が必要で、Lambda より ECS が適合
   - コールドスタート回避
   - 重量級フレームワーク（Spring）の起動コスト削減
   - 長時間処理が可能

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

## AWSアーキテクチャ構成図


---

以下は本アプリケーションのAWS構成図です。



## AWS全体構成図
![AWSアーキテクチャ構成図](./architecture/aws/image.png)



## ユーザー同士のチャット
![ユーザー同士のチャット](./architecture/aws/aws-architecture-chat.png)





## AIとユーザーのチャット
![AIとユーザーのチャット](./architecture/aws/aws-architecture-ai-chat.png)


---

## 🚀 今後の展望
### 技術資格
- AWS SAP
- 応用情報

### 機能拡張
- 音声チャット
- Polly による AI 音声応答
- CloudFront + Lambda@Edge の認証強化

---

## 🛠 フロントエンドセットアップ手順

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
