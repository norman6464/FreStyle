# FreStyle
ユーザーやAIと手軽にコミュニケーションが取れるチャットアプリです。  
私自身、日常の中で「気軽にコミュニケーションを取れる場」が必要だと感じていました。  
そのため、フォローやフレンド登録をしなくてもユーザー検索だけで会話を始められ、  
さらにAIが会話をサポートしてくれるアプリを開発しました。  

本アプリはAWS上にデプロイしています。
## 🌐 デプロイURL

## 🧰 使用技術
- **Frontend:** React / Tailwind CSS  
- **Backend:** Spring Boot / lambda
- **Infrastructure:** AWS (ECS, RDS, S3, Route53, DynamoDB, lambda, Cognito, API Gateway)  
- **CI/CD:** GitHub Actions 
- **Database:** MariaDB,DynamoDB

## ⚙️ 主な機能
- ユーザー登録・ログイン（JWT認証,OIDC）
- ユーザー同士のチャット、AIとのチャット
- プロフィール編集
- Googleログイン
- GitHub Actionsによる自動デプロイ


## 💡 工夫した点
- WebSocket、Httpの二つでアーキテクチャーを分断したことWebsocket通信ではAPI Gateway,lambdaでHttpではECSを使用をしている
- CloudFormationを用いたインフラのコード管理
- JWT認証をHttpOnly Cookieにし、バックエンド側でフィルターを用いてヘッダーにJWTトークンをセットをしセキュリティー面を考慮した
- OIDCを実現したいため、CloudFrontでHttps化,グローバル化をした


## 🧠 苦労した点・学び
- WebSocketを使ったアプリの実装が初めてだったのでECSで実装するかAPI Gateway + lambdaで実装をするか迷いました。  
  → API Gatewayのカスタムオーソライザーを活用して統一的に解決しました。
- Spring BootのSpring SecurityでJWKを使った認証はAuthorizationヘッダーで行うのでHttpOnly CookieからAuthorizationヘッダーに変換するのが大変だった
- ALBでの設定でTLS/SSLアクセラレータとしてバックエンドのECSはHttp化をするのかを考慮しました


## 🚀 今後の展望

### 技術的な目標
- AWS Solution Architect Professional（SAP）を取得したい  
- CloudFront＋Lambda@Edgeによる認証強化を検討中  

### アプリの機能拡張
- 音声でもコミュニケーションできるようにする  
- AIとの音声応答にPollyを活用する


## フロントエンドセットアップ手順

1. リポジトリをクローンして `frontend` ディレクトリに移動します。

2. `npm install` を実行し、`package.json` に記載されている依存パッケージをインストールして `node_modules` を作成します。

3. `npm run dev` で動作確認を行います。

4. Tailwind CSSの動作も確認してください。  
   ※ 使用しているNode.jsのバージョンによっては、Tailwind CSSが正常に動作しない可能性があります。

5. 動作しない場合は、一度 Tailwind CSSをアンインストールします。  
   ```bash
   npm uninstall tailwindcss
   
6. そのあとにもう一度インストールをします。
   ```bash
   `npm install -D tailwindcss@バージョン指定`

7.  `npx tailwindcss init -p` で初期設定をします。


## AWSアーキテクチャ構成図

以下は本アプリケーションのAWS構成図です。

![AWSアーキテクチャ構成図](./architecture/aws/aws-architecture.png)


## ユーザー同士のチャット
![ユーザー同士のチャット](./architecture/aws/aws-architecture-chat.png)


## AIとユーザーのチャット
![AIとユーザーのチャット](./architecture/aws/aws-architecture-ai-chat.png)

