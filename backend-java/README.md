# backend-java — Spring Boot 版 FreStyle バックエンド（移行作業中）

Go(`backend/`) で動いている FreStyle バックエンドを **Java / Spring Boot へ段階移行**するための作業ディレクトリ。
Go 版は当面**残す**（フォールバック / 比較用）。本ディレクトリが本番相当になるまでは ECS は Go 版のまま。

## なぜ Java へ移すのか

- チーム開発を見据え、アノテーションで意図が明示される Spring Boot の方がレビューしやすい
- 層・責務の境界がフレームワークの仕組み（DI / `@Service` / `@RestController` 等）で明示され、
  新しく加わるメンバーが把握しやすい

## 技術スタック

- Java 21 / Spring Boot 4.0.6 / Gradle 9.5.1（Wrapper 同梱、システム Gradle 不要）
- Spring MVC（`spring-boot-starter-webmvc`）/ Spring Data JPA / Bean Validation / Actuator / Lombok
- DB: ローカル/テストは H2（インメモリ）、本番は Supabase(PostgreSQL) を環境変数で注入

## クリーンアーキテクチャ対応（Go → Java）

Go 版の層構造をそのまま写経している。

| 層 | Go(`backend/`) | Java(`backend-java/`) |
|---|---|---|
| handler | `internal/handler` (Gin) | `handler/` (`@RestController`) |
| usecase | `internal/usecase` | `usecase/` (`@Service`) |
| repository(port) | `internal/usecase/repository` | `repository/` (`JpaRepository` interface) |
| 実装(adapter) | `internal/adapter/persistence` | Spring Data JPA が自動生成 |
| domain | `internal/domain` | `domain/` (`@Entity`) |

依存方向は handler → usecase → repository → domain（Go 版と同一）。

## 現在の実装範囲（最小骨格）

| エンドポイント | 説明 | 状態 |
|---|---|---|
| `GET /api/v2/health` | 死活確認（`{"status":"UP"}`、ALB/Blue-Green 用） | ✅ Go 版互換 |
| `GET /api/v2/notes` | ノート一覧 | ✅（認証は未移植、暫定 userId=1） |
| `POST /api/v2/notes` | ノート作成 | ✅（同上） |

JSON フィールド名は Go 版（`isPublic`/`isPinned` 等）と一致させている。

### 未移植（今後 PR で段階移植）

認証(Cognito JWT) / AI チャット(Bedrock SSE) / コース・教材 / 演習(サンドボックス実行) / 通知 / レポート(SQS) +
S3 / DynamoDB / SES 連携。これらを 1 機能 = 1 PR で移していく。

## メモリ実測（ECS コスト検証）

ユーザーの最大懸念は「現行 ECS スペック（0.25 vCPU / 0.5 GB）で Spring Boot が動くか＝コスト増を避けられるか」。
最小骨格をローカル(macOS arm64 / Temurin 21 / H2)で実測した結果:

| 条件 | 起動 | RSS | ヒープ上限 | 実使用ヒープ |
|---|---|---|---|---|
| 無制約(`MaxRAMPercentage=75`) | 4 秒 | 361MB（負荷後ピーク 374MB） | — | — |
| **0.5GB 制約(`MaxRAM=512m`)** | **3 秒** | **270MB** | 360MB(自動) | 55MB |

- JVM は container-aware にヒープ上限を自動計算（512MB → ヒープ 360MB cap）。実使用ヒープは 55MB と余裕
- **最小骨格なら 0.5GB で動く**（RSS 270MB、~240MB の余裕）
- ⚠️ ただしフル機能（Cognito/Bedrock/S3/DynamoDB/SQS/SES + 全 Controller）を載せると Metaspace/ロードクラス増で
  RSS は **400〜500MB** に上がる見込み。移植が進んだ段階で再測定し、必要なら 0.5 vCPU / 1 GB へ
- CPU: Fargate 0.25 vCPU では起動が 20〜40 秒になり得るため、ヘルスチェック猶予期間の調整が要る

## ローカルでの動かし方

```bash
cd backend-java

# ビルド + テスト（Java 21 が必要。toolchain で自動解決されるが JAVA_HOME=21 推奨）
export JAVA_HOME=$(/usr/libexec/java_home -v 21)
./gradlew build

# 起動（H2 インメモリ）
./gradlew bootRun
# または
java -XX:MaxRAMPercentage=70 -jar build/libs/backend-0.0.1-SNAPSHOT.jar

# 動作確認
curl localhost:8080/api/v2/health
curl -X POST localhost:8080/api/v2/notes -H 'Content-Type: application/json' \
  -d '{"title":"はじめてのノート","content":"hello"}'
curl localhost:8080/api/v2/notes
```

## Docker

`Dockerfile` は 2 stage（Temurin 21 JDK でビルド → JRE で実行）。コンテナ memory limit からヒープを自動計算する
`-XX:MaxRAMPercentage=70` で起動する。Go 版同様 nonroot(UID=65532) / port 8080。

```bash
docker build -t frestyle-backend-java .
docker run -p 8080:8080 frestyle-backend-java
```

本番デプロイ（buildspec / taskdef / ECS 切替）は移植が一通り済んでから別 PR で整備する。
