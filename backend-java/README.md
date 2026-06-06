# backend-java — FreStyle バックエンド (Spring Boot)

FreStyle のバックエンドを Java / Spring Boot で実装するプロジェクト。
チーム開発を見据え、アノテーションで層・責務が明示され、新しく加わるメンバーが
レビュー・把握しやすい構成を狙う。

> 本ディレクトリが本番相当になるまでは段階的に機能を追加していく。1 機能 = 1 PR を基本とする。

## 技術スタック

- Java 21 / Spring Boot 4.0.6 / Gradle 9.5.1（Wrapper 同梱、システム Gradle 不要）
- Spring MVC（`spring-boot-starter-webmvc`）/ Spring Data JPA / Bean Validation / Actuator / Lombok
- DB マイグレーション: Flyway（スキーマの正は `db/migration` の SQL）
- DB: ローカル/テストは H2（インメモリ）、本番は Supabase(PostgreSQL) を環境変数で注入

## コーディング規約（クリーンコード）

『プリンシプル オブ プログラミング』『クリーンコード』を意識する。特に **空行で「処理の意味の
まとまり」を区切る**。読み手が一目で段落を掴めるようにする。

| 場面 | 空行 |
|---|---|
| メソッド間 | ✅ 必須（1 行） |
| 処理グループの切れ目（入力検証 → 本処理 → 後処理 など） | ✅ 推奨 |
| `return` の直前（手前に処理が続いた後） | ✅ 推奨 |
| 変数宣言 1〜2 個の直後 | ❌ 不要 |
| `{` の直後（メソッド冒頭） | ❌ 不要 |

その他: 1 メソッド 1 責務 / 早期 return でネストを浅く / 名前で意図を語り「何を」コメントしない
（「なぜ」を残す）。

## DB マイグレーション（Flyway）

スキーマは Flyway が所有する。Hibernate の自動 DDL は使わない（`spring.jpa.hibernate.ddl-auto=none`）。

- マイグレーション SQL は `src/main/resources/db/migration/V{n}__{説明}.sql` に積む（版番号は連番）
- 起動時に Flyway が未適用分を順に流し、`flyway_schema_history` で適用済みを管理する
- SQL は H2（テスト/ローカル）と PostgreSQL（本番）の双方で通る書き方にする
- 既存 GORM 製スキーマが入った本番 DB に後から載せる場合に備え `baseline-on-migrate=true` を設定済
- Spring Boot 4.0 は autoconfig がモジュール分割されたため、Flyway 連携には `spring-boot-flyway` モジュールが必要

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

## 認証（Cognito JWT）

Cognito の **access_token(JWT)を HttpOnly Cookie で受け取り、Cognito の JWKS で署名検証**する
（Spring Security OAuth2 Resource Server）。検証済み JWT の `sub` をキーに users 行を引き、
業務処理では内部の数値 `id` を使う 2 層 ID 構成。

- `GET /api/v2/health` は認証なしで通す。それ以外は JWT 必須（無ければ 401）
- 標準の `Authorization: Bearer` ではなく、`access_token` Cookie からトークンを取り出す
  （`CognitoCookieBearerTokenResolver`）。XSS 耐性のため HttpOnly Cookie 方式
- `cognito:groups` に admin が含まれると `super_admin` に自動昇格（`AuthService`）

### ログインフロー

- `POST /api/v2/auth/login`: 認可コードを Cognito token endpoint で token に交換し、
  `access_token` / `refresh_token` を HttpOnly Cookie で発行（`CognitoTokenClient` / `AuthCookies`）
- **招待ゲート**（`UserProvisioningService`）: 新規ユーザーは「pending な招待」か「Cognito admin
  グループ」のいずれかが必要。無ければ 403 で拒否。招待の role / company を反映し accepted にマーク
- `POST /api/v2/auth/logout`: Cookie を破棄
- `POST /api/v2/auth/refresh`: `refresh_token` Cookie で access_token を再発行
- ⚠️ 未対応(フォローアップ): Cookie 認証向け CSRF 対策（Go の `CsrfMiddleware` 相当）

## 現在の実装範囲

| エンドポイント | 説明 | 認証 |
|---|---|---|
| `GET /api/v2/health` | 死活確認（`{"status":"UP"}`、ロードバランサ用） | 不要 |
| `GET /api/v2/auth/me` | 現在ユーザー（+ groups / isAdmin / onboarded） | 必須 |
| `POST /api/v2/auth/login` | 認可コード→token 交換 + Cookie 発行（招待ゲート） | 不要 |
| `POST /api/v2/auth/logout` | Cookie 破棄 | 不要 |
| `POST /api/v2/auth/refresh` | access_token 再発行 | 不要 |
| `GET /api/v2/notes` | ノート一覧（認証ユーザーのもの） | 必須 |
| `POST /api/v2/notes` | ノート作成（`title` 必須） | 必須 |
| `GET /api/v2/courses` | コース一覧（company/role でフィルタ） | 必須 |
| `GET /api/v2/courses/{id}` | コース詳細 | 必須 |
| `GET /api/v2/courses/{id}/materials` | コース内教材一覧 | 必須 |
| `GET /api/v2/teaching-materials/{id}` | 教材詳細 | 必須 |
| `POST /api/v2/company-applications` | 企業の利用申請（公開フォーム） | 不要 |
| `GET /api/v2/admin/company-applications` | 申請一覧（新しい順） | super_admin |
| `PATCH /api/v2/admin/company-applications/{id}/status` | 申請 status 更新（approved/rejected/pending） | super_admin |

コース/教材の閲覧権: trainee は自社の published のみ、company_admin / super_admin は draft 含む。
super_admin は会社を跨いで閲覧可。アクセス制御は `CourseService` で actor の company/role を見て判定。

### 今後追加する機能（1 機能 = 1 PR）

企業申請受理時の super_admin 通知 / コース/教材の CRUD(company_admin) / AI チャット(Bedrock SSE) / 演習(サンドボックス実行) /
通知 / レポート(SQS) + S3 / DynamoDB / SES 連携。CSRF 対策。

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
