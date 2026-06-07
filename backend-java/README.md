# backend-java — FreStyle バックエンド (Spring Boot)

FreStyle のバックエンドを Java / Spring Boot で実装するプロジェクト。
チーム開発を見据え、アノテーションで層・責務が明示され、新しく加わるメンバーが
レビュー・把握しやすい構成を狙う。

> 本ディレクトリが本番相当になるまでは段階的に機能を追加していく。1 機能 = 1 PR を基本とする。

## 採用アーキテクチャ — レイヤードアーキテクチャ（インフラ境界のみ DIP）

本プロジェクトの採用アーキテクチャは **レイヤードアーキテクチャ**。ただし **外部システム連携
（インフラ層）の境界だけ依存性逆転（DIP）** を効かせる。純粋なクリーンアーキテクチャ / ヘキサゴナル
（フレームワーク非依存の domain モジュール・JPA エンティティとドメインエンティティの二重持ち）は **採らない**。

### なぜこの形か（設計判断）

- **目的はチーム開発での分かりやすさ**。`controller / service / repository` は Spring 開発者の共通言語で、
  新しく加わるメンバーが一目で層・責務を把握できる。純粋クリーンアーキのマルチモジュールは初見だと逆に重い
- **ドメインは CRUD + 外部連携が主体**で、複雑な不変条件を持たない。純粋化の儀式（相互マッピング・
  マルチモジュール）は割に合わない（＝オーバーエンジニアリング）
- 一方 **「stub 差し替えでテスト / 資格情報なし起動」の価値は大きい**ため、インフラ境界だけポート
  （interface）で逆転させる。Spring の DI でこの“一番おいしい部分”は低コストで得られる
- 補足: Spring はフレームワークを業務コードに撒く設計のため**純粋クリーンアーキとは構造的に相性が悪い**。
  「レイヤード + インフラ境界 DIP」は Spring の流れに乗りつつ要点を押さえる実用的な最適点

### 層と依存方向

| 層 | パッケージ | 役割 |
|---|---|---|
| プレゼンテーション | `controller` + `dto` | HTTP 受付（inbound）。`@Valid` で入力検証、レスポンスは record DTO |
| アプリケーション | `service` | ビジネスロジック / オーケストレーション。**1 メソッド = 1 ユースケース** |
| 永続化 | `repository` + `entity` | Spring Data JPA。RDB アクセス |
| インフラ | `infra/*` | 外部システム連携（AWS SDK / 外部 HTTP）。**ポート interface + 実装 + stub** |
| 横断 / 設定 | `security` / `config` | 認証ヘルパー、`@Configuration`・各 infra の bean 組み立て |

- 依存方向は **上 → 下**（プレゼン → アプリ → 永続化 / インフラ）
- **唯一の依存性逆転**: `service` は infra の **ポート interface**（例: `ReportEnqueuer` /
  `ProfileImagePresigner` / `AiChatMessageReader`）にのみ依存し、具象実装は `config` が注入する。
  これにより本番実装 / stub を設定有無で差し替えられる

### ユースケースの置き場所

- 原則 **`service` の各 public メソッドが 1 つのユースケース**（DDD の Application Service スタイル）
- 1 操作が複雑化したら（複数ポートを跨ぐ / 分岐が増える / トランザクション境界が重い）、**その操作だけ**
  `XxxUseCase`（`@Component` + `execute()`）に切り出す（“痛くなったら抽出”。例: AI チャット SSE ストリーム）

### やらないこと（アンチパターン）

- ❌ `controller` から `repository` / `infra` を直接呼ぶ（必ず `service` 経由）
- ❌ `service` が infra の **具象クラス** に依存する（必ずポート interface 経由 / 具象は `config` が注入）
- ❌ `entity` を純粋ドメイン化するための二重エンティティ / フレームワーク非依存モジュール化（この規模では過剰）
- ❌ 一貫性を崩してヘキサゴナル / 純粋クリーンアーキへ部分的に振り直す（**一貫性 > アーキの流行**）

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

上記「採用アーキテクチャ（レイヤード + インフラ境界 DIP）」をパッケージに落とした形。

```
com.normanblog.frestyle
├── FrestyleBackendApplication   起動クラス
├── controller   HTTP エンドポイント (@RestController) ＝ inbound adapter
├── service      ビジネスロジック (@Service)。外部連携は infra のポート interface に依存
├── repository   永続化 (Spring Data JPA, JpaRepository)
├── entity       JPA エンティティ (@Entity)
├── dto          リクエスト/レスポンスの転送オブジェクト (record)
├── config       設定 (@ConfigurationProperties / @Configuration / Security / 各 infra の bean 組み立て)
├── security     認証ヘルパー (CurrentUserProvider 等。横断的関心事)
└── infra        外部システムへの outbound adapter（ポート interface + 本番実装 + stub）
    ├── cognito  Cognito token 交換クライアント
    ├── s3       画像 S3 presigner + 添付 downloader
    ├── sqs      学習レポート生成ジョブの enqueuer (現状 no-op stub)
    ├── dynamo   AI チャットメッセージの DynamoDB reader / writer
    ├── bedrock  AI チャットの Bedrock Converse ストリーミング client
    └── exec     コード演習サンドボックス(子プロセス実行)
```

- `@Entity` は永続化専用とし、API には公開しない。レスポンスは `dto` の record（例: `NoteResponse`）に詰め替える
- リクエストボディは `dto` の record に Bean Validation（`@NotBlank` 等）を付け、`controller` で `@Valid` 検証する
- **外部システム連携（AWS SDK / 外部 HTTP）は `infra/<システム>` に集約**する。`service` はポート interface
  （例: `ReportEnqueuer` / `ProfileImagePresigner` / `AiChatMessageReader`）にのみ依存し、実装は `infra` 側に置く
  （依存性逆転）。資格情報の無い環境向けに各 infra は stub 実装を持ち、`config` が設定有無で実装を切り替える

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
- ⚠️ 未対応(フォローアップ): Cookie 認証向け CSRF 対策（Go・frontend とも未実装の新規対応。SameSite=None のため
  有効化には frontend の X-XSRF-TOKEN 送信 + 2 段階デプロイが必要）

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
| `POST /api/v2/courses` | コース作成 | 管理者 |
| `PUT /api/v2/courses/{id}` | コース更新 | 管理者 |
| `DELETE /api/v2/courses/{id}` | コース削除（配下教材も cascade） | 管理者 |
| `GET /api/v2/teaching-materials/{id}` | 教材詳細 | 必須 |
| `POST /api/v2/teaching-materials` | 教材作成（`courseId` 必須） | 管理者 |
| `PUT /api/v2/teaching-materials/{id}` | 教材更新 | 管理者 |
| `DELETE /api/v2/teaching-materials/{id}` | 教材削除 | 管理者 |
| `POST /api/v2/company-applications` | 企業の利用申請（公開フォーム / 全 super_admin へ通知） | 不要 |
| `GET /api/v2/admin/company-applications` | 申請一覧（新しい順） | super_admin |
| `PATCH /api/v2/admin/company-applications/{id}/status` | 申請 status 更新（approved/rejected/pending） | super_admin |
| `GET /api/v2/profile/{userId}` | プロフィール取得（`me` か自分の id のみ。他者 403） | 必須 |
| `PUT /api/v2/profile/{userId}`（+ `/update` 別名） | プロフィール更新（displayName / bio / avatarUrl / status） | 必須 |
| `POST /api/v2/profile/me/onboarding/complete` | Welcome 完了（`onboarded_at` を冪等にセット） | 必須 |
| `POST /api/v2/profile/{userId}/image/presigned-url` | プロフィール画像の S3 PUT 署名 URL 発行（`me` か自分の id のみ） | 必須 |
| `GET /api/v2/learning-reports` | 学習レポート一覧（期間降順 / 自分のみ） | 必須 |
| `POST /api/v2/learning-reports/generate` | 月次レポート生成要求（year+month / 202 + pending） | 必須 |
| `GET /api/v2/ai-chat/sessions` | AI チャットセッション一覧（自分のみ・作成日降順） | 必須 |
| `POST /api/v2/ai-chat/sessions` | セッション作成（title 必須 / sessionType 省略時 free） | 必須 |
| `GET /api/v2/ai-chat/sessions/{id}` | セッション取得（所有者のみ / 他人 403） | 必須 |
| `PUT /api/v2/ai-chat/sessions/{id}` | タイトル更新（所有者のみ） | 必須 |
| `DELETE /api/v2/ai-chat/sessions/{id}` | セッション削除（所有者のみ） | 必須 |
| `GET /api/v2/ai-chat/sessions/{id}/messages` | メッセージ履歴（DynamoDB / 所有者のみ・作成順） | 必須 |
| `POST /api/v2/ai-chat/attachments/upload-url` | 添付の S3 PUT 署名 URL（画像のみ / 5MB上限 / 415・413） | 必須 |
| `POST /api/v2/ai-chat/stream` | Bedrock へ送信し token を SSE 配信（session/token/done/error） | 必須 |
| `POST /api/v2/code/execute` | コード実行サンドボックス（java/php/go / stdout・stderr・exitCode） | 必須 |
| `GET /api/v2/notifications` | 通知一覧（作成日降順） | 必須 |
| `GET /api/v2/notifications/unread-count` | 未読通知数（バッジ用） | 必須 |
| `PATCH /api/v2/notifications/{id}/read`（+ `PUT` 別名） | 通知 1 件を既読化（所有者検証込み） | 必須 |
| `PATCH /api/v2/notifications/read-all`（+ `PUT` 別名） | 全未読をまとめて既読化 | 必須 |

通知は WHERE 句で `user_id` を絞り、他人の通知を既読化できないようにする（所有者検証）。
企業申請（`POST /company-applications`）受理時に全 super_admin へ `company_application` 型の通知を
best-effort で作成する（通知失敗は申請の成功を妨げない / ログのみ）。push 連携（SNS 等）は別 PR。

プロフィールは自分のみ操作可（IDOR 対策）。displayName は `users`、bio/avatar/status は `profiles` に保存。
更新は省略フィールドの既存値を保持。画像は `profiles/{userId}/{epochNanos}{ext}` キーへ S3 PUT 署名 URL
（10 分）を発行する（AWS SDK v2）。`frestyle.s3.bucket`（= `NOTE_IMAGES_BUCKET`）未設定時は AWS を
呼ばない stub presigner にフォールバックし、資格情報の無いローカル / テストでも起動・検証できる。

コース/教材の権限: 閲覧は trainee=自社の published のみ / 管理者(company_admin・super_admin)=draft 含む、
super_admin は会社跨ぎ可。編集系は管理者かつ(super_admin または同一会社)。アクセス制御は
`CourseService` / `TeachingMaterialService` に集約。

### 今後追加する機能（1 機能 = 1 PR）

演習の DB 部分(master 演習一覧/詳細・提出履歴) / SES 連携。CSRF 対策。

**コード実行サンドボックス** (`POST /api/v2/code/execute`): ユーザーの java/php/go コードを子プロセスで実行し
stdout/stderr/exitCode を返す。 セキュリティの肝は **子プロセスの環境クリーン化**（backend のシークレット
= DB/AWS/Cognito を一切渡さず PATH/LANG のみ）+ timeout（既定 10s, 超過は destroyForcibly で 124）+
出力上限（64KB）+ JVM ヒープ上限（`-Xmx64m`）+ 専用 temp ディレクトリ + 非 root。 Java は単一ファイル実行
`java Main.java`（runtime image を JRE→**JDK** に変更した理由）。 ネットワーク egress の遮断は infra 側で担保
する前提。 `frestyle.code-exec.enabled=false`（env `CODE_EXEC_ENABLED`）で無効化でき、 無効時は 503。
対応言語は java/php/go（runtime image に JDK + php-cli + golang-go を同梱）。php は disable_functions/open_basedir で堅牢化、go は go run の単一ファイル実行(専用 HOME/GOCACHE)。

AI チャットは **セッション(メタデータ=RDS) + メッセージ(DynamoDB read/write) + 添付(S3) + SSE
ストリーム(Bedrock Converse)** を移植し、機能一式が揃った。セッション / メッセージは取得・更新・削除で
所有者検証を行い、他人のものは 403（Go 版に無かった IDOR 対策を追加）。メッセージは DynamoDB
(`frestyle.dynamo.ai-chat-table` = `DYNAMODB_AI_CHAT_TABLE`、PK=sessionId / SK=messageId) を Query。
添付は画像のみ許可・5MB 上限を `AiChatAttachmentRules` で検証し、未対応 MIME は 415 / サイズ超過は 413。

**SSE ストリーム** (`POST /api/v2/ai-chat/stream`): 事前検証(本文/添付 key prefix/所有者)を同期で行い不正は
400/403、通過後に `SseEmitter` で session/token/done/error を配信する。複数ポート(session/message/S3/
Bedrock)を跨ぐため `SendAiMessageUseCase`(@Component)に切り出し、別スレッドで実行して listener 経由で
SSE へ書く(アプリ層は HTTP/SSE を知らない)。Bedrock は `frestyle.bedrock.model-id`(= `BEDROCK_MODEL_ID`)
未設定時は固定応答を返す stub にフォールバックし、資格情報の無いローカル / テストでも検証できる。

学習レポートは生成要求で pending 行を作りキューに積む（`ReportEnqueuer`）。実 SQS 送信は
コンシューマ契約が未確定のため現状 no-op stub（Go 版と同じ挙動）で、実装は別 PR。

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

`Dockerfile` は 2 stage（Temurin 21 JDK でビルド → **JDK** で実行）。コード演習サンドボックスが
`java`/`php`/`go` を実行するため runtime image は JDK（+ apt で `php-cli` / `golang-go` を同梱）。コンテナの
memory limit からヒープを自動計算する `-XX:MaxRAMPercentage=70` で起動し、nonroot(UID=65532) / port 8080 で動く。

```bash
docker build -t frestyle-backend-java .
docker run -p 8080:8080 frestyle-backend-java
```

本番デプロイ（buildspec / taskdef / ECS 切替）は実装が一通り済んでから別 PR で整備する。
