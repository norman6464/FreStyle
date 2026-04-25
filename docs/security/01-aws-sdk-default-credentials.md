# AWS SDK の標準クレデンシャルチェーン採用

## 背景

旧実装では Spring Boot 起動時に `application.properties` の `aws.access-key` / `aws.secret-key` を読み、各 AWS Client（S3 / SQS / DynamoDB / Bedrock / Cognito）の builder に `StaticCredentialsProvider.create(AwsBasicCredentials.create(...))` で明示注入していた。

これは以下の問題があった:

1. **長寿命のアクセスキー** が ECS タスクの環境変数経由で注入されており、漏洩リスクが大きい
2. **ローテーションが手動** — キー変更時に Task Definition / Secret / 環境変数 を全て更新する必要がある
3. **ECS Task Role への移行が阻害される** — IAM Role を使うときも明示クレデンシャルが優先されるため、Task Role に切り替えてもアクセスキーが残っている限り Role は使われない

## 変更内容

8 ファイル全てで `credentialsProvider` の明示指定を削除し、AWS SDK builder に委譲する形に変更:

| ファイル | 変更前 | 変更後 |
|---|---|---|
| `config/S3Config.java` | `.credentialsProvider(StaticCredentialsProvider.create(...))` | builder に渡さない |
| `config/SqsConfig.java` | 同上 | 同上 |
| `service/AiChatMessageDynamoService.java` | 同上 | 同上 |
| `service/AiChatService.java` | 同上 | 同上 |
| `service/BedrockService.java` | 同上 | 同上 |
| `service/ChatMessageDynamoService.java` | 同上 | 同上 |
| `service/CognitoAuthService.java` | コンストラクタで access/secret 受け取り | region のみ受け取る |
| `service/NoteService.java` | 同 PostConstruct パターン | 同上 |

## なぜ「明示しない」が正解か

AWS SDK v2 の builder は `credentialsProvider` を呼ばないと **`DefaultCredentialsProvider`** を自動的に使用する。これは以下を順に試すクレデンシャルチェーン:

1. **環境変数** `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` / `AWS_SESSION_TOKEN`
2. Java システムプロパティ `aws.accessKeyId` / `aws.secretAccessKey`
3. **Web Identity Token File**（GitHub Actions OIDC など）
4. AWS SSO（プロファイル設定）
5. `~/.aws/credentials` のプロファイル
6. `~/.aws/config` のプロファイル
7. **ECS コンテナクレデンシャル**（Task Role）
8. **EC2 Instance Profile**

これにより:
- **ローカル開発**: 環境変数 or `~/.aws/credentials` で動く
- **ECS 本番**: Task Role が自動採用される（環境変数を消すだけで切り替わる）
- **GitHub Actions**: OIDC を使えば `Web Identity Token File` 経由で短期トークンが取れる

つまり **同一コードでローカル / 本番 / CI のいずれでも動く** 上、本番では IAM Role により credentials が短命化される。

## アプリ側の動作確認

```bash
# ローカル: 環境変数から取得（後方互換）
export AWS_ACCESS_KEY_ID=AKIA...
export AWS_SECRET_ACCESS_KEY=...
export AWS_REGION=ap-northeast-1
./gradlew bootRun

# ECS 本番: Task Role が自動的に効く（次の IaC PR でセットアップ）
```

## `application.properties` の変更

```diff
- aws.access-key=${AWS_ACCESS_KEY}
- aws.secret-key=${AWS_SECRET_KEY}
  aws.region=${AWS_REGION}
```

`AWS_ACCESS_KEY` / `AWS_SECRET_KEY` 環境変数は ECS Task Definition から削除する（次のインフラ PR で対応）。AWS SDK は標準環境変数 `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` を直接読むので Spring 経由の中継は不要。

## テスト

- 新たな failing は発生していない（既存の 3 件 = `FreStyleApplicationTests.contextLoads`, `AiChatWebSocketController.rephraseMessage` × 2 は本変更前から失敗中）
- `CognitoAuthServiceTest` のコンストラクタ呼び出しを 3 引数 → 1 引数に更新

## 関連リソース

- 次に予定している IaC 変更: `frestyle-infrastructure` 側で
  - ECS Task Role を新設し、Bedrock / DynamoDB / S3 / SQS / Secrets Manager の最小権限を付与
  - Task Definition の `Environment` から `AWS_ACCESS_KEY` / `AWS_SECRET_KEY` を削除
  - DB パスワードや Cognito Client Secret は `Secrets:` フィールドで Secrets Manager から直接 inject

## 参考

- [AWS SDK for Java v2 - Default Credentials Provider Chain](https://docs.aws.amazon.com/sdk-for-java/latest/developer-guide/credentials-chain.html)
- [Amazon ECS Task IAM Roles](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html)
