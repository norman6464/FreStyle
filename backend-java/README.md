# backend-java — FreStyle バックエンド (Spring Boot)

FreStyle のバックエンドを Java / Spring Boot で実装するプロジェクト。
チーム開発を見据え、アノテーションで層・責務が明示され、新しく加わるメンバーが
レビュー・把握しやすい構成を狙う。

> 本ディレクトリが本番相当になるまでは段階的に機能を追加していく。1 機能 = 1 PR を基本とする。

## 技術スタック

- Java 21 / Spring Boot 4.0.6 / Gradle 9.5.1（Wrapper 同梱、システム Gradle 不要）
- Spring MVC（`spring-boot-starter-webmvc`）/ Spring Data JPA / Bean Validation / Actuator / Lombok
- DB: ローカル/テストは H2（インメモリ）、本番は Supabase(PostgreSQL) を環境変数で注入

## パッケージ構成

Spring Boot の標準的なレイヤ構成。

```
com.normanblog.frestyle
├── FrestyleBackendApplication   起動クラス
├── controller   HTTP エンドポイント (@RestController)
├── service      ビジネスロジック (@Service)
├── repository   永続化 (Spring Data JPA, JpaRepository)
├── entity       JPA エンティティ (@Entity)
└── dto          リクエスト/レスポンスの転送オブジェクト (record)
```

- `@Entity` は永続化専用とし、API には公開しない。レスポンスは `dto` の record（例: `NoteResponse`）に詰め替える
- リクエストボディは `dto` の record に Bean Validation（`@NotBlank` 等）を付け、`controller` で `@Valid` 検証する

## 現在の実装範囲

| エンドポイント | 説明 | 状態 |
|---|---|---|
| `GET /api/v2/health` | 死活確認（`{"status":"UP"}`、ロードバランサ用） | ✅ |
| `GET /api/v2/notes` | ノート一覧 | ✅（認証は未実装、暫定 userId=1） |
| `POST /api/v2/notes` | ノート作成（`title` 必須） | ✅（同上） |

### 今後追加する機能（1 機能 = 1 PR）

認証(Cognito JWT) / AI チャット(Bedrock SSE) / コース・教材 / 演習(サンドボックス実行) /
通知 / レポート(SQS) + S3 / DynamoDB / SES 連携。

## メモリ実測（ECS スペック検証）

現行の ECS スペック（0.25 vCPU / 0.5 GB）で動くかを確認するため、最小構成を実測した
（macOS arm64 / Temurin 21）。

| 条件 | 起動 | RSS |
|---|---|---|
| ローカル無制約 | 4 秒 | 361MB（負荷後ピーク 374MB） |
| ローカル 0.5GB 制約(`MaxRAM=512m`) | 3 秒 | 270MB（ヒープ cap 360MB / 実使用 55MB） |
| Docker `--memory 512m` | 4 秒 | 274.9MiB / 512MiB（53.7%） |

- JVM は container-aware にヒープ上限を自動計算（512MB → ヒープ 360MB cap、実使用 55MB と余裕）
- **最小構成なら 0.25 vCPU / 0.5 GB で動作**
- ⚠️ フル機能（認証 / AI チャット / S3 / DynamoDB / SQS / SES + 全 Controller）を載せると
  ロードクラス増で RSS は 400〜500MB に上がる見込み。実装が進んだ段階で再測定し、必要なら
  0.5 vCPU / 1 GB へ
- CPU: Fargate 0.25 vCPU では起動が 20〜40 秒になり得るため、ヘルスチェック猶予期間の調整が要る

## ローカルでの動かし方

```bash
cd backend-java

export JAVA_HOME=$(/usr/libexec/java_home -v 21)
./gradlew build              # ビルド + テスト

./gradlew bootRun            # H2 インメモリで起動
# または
java -XX:MaxRAMPercentage=70 -jar build/libs/backend-0.0.1-SNAPSHOT.jar

curl localhost:8080/api/v2/health
curl -X POST localhost:8080/api/v2/notes -H 'Content-Type: application/json' \
  -d '{"title":"はじめてのノート","content":"hello"}'
curl localhost:8080/api/v2/notes
```

## Docker

`Dockerfile` は 2 stage（Temurin 21 JDK でビルド → JRE で実行）。コンテナの memory limit から
ヒープを自動計算する `-XX:MaxRAMPercentage=70` で起動し、nonroot(UID=65532) / port 8080 で動く。

```bash
docker build -t frestyle-backend-java .
docker run -p 8080:8080 frestyle-backend-java
```

本番デプロイ（buildspec / taskdef / ECS 切替）は実装が一通り済んでから別 PR で整備する。
