# 0002: バックエンドを Java(Spring Boot) から Go へ差し戻す

## 1. 基本情報

| 項目 | 内容 |
|---|---|
| タイトル | backend-java → backend(Go) 差し戻し + 機能パリティ + ECR/ECS 切替 + Java 資産削除 |
| 担当者 | @norman6464 |
| レビュアー | @norman6464 |
| ステークホルダー | 開発メンバー / 新規参画メンバー |
| ステータス | Phase A 完了（Go 機能パリティ回復済 / 2026-06-09）。次は Phase B 切替（要 GO） |
| 関連 | （Issue 起票予定） |

## 2. 背景と目的 (Background & Objective)

**背景**:
- 本番は **0.25 vCPU / 0.5 GB の極小 Fargate + 夜間 $0 teardown**（scheduled-stop/start で毎日 ALB/ECS を破棄・再作成）。この構成では JVM の起動レイテンシ・常駐メモリがボトルネックになり、**パフォーマンス面で実運用に難**がある。
- Spring Boot → Go 移行（2026-04 完了）後、再び Go → Java(Spring Boot) へ段階移植中（backend-java/, 起点 PR #1802）だったが、上記パフォーマンス問題により **Java をやめて Go に戻す**判断。
- **Go 資産はほぼ残存**（196 .go ファイル）。Java 移植の過程で Go から削除されたのは course/teaching-material(#1831) と learning_report(#1833) の 2 機能のみ。
- **クリーンアーキテクチャ強制は Go 側で生きている**: `backend/cmd/archlint` が依存方向違反（CLAUDE.md §2）を静的検出し exit 1 を返す。CI（`.github/workflows/ci-backend-go.yml`）に組み込み済で現在グリーン。

**目的**:
- Go バックエンドを **Java と機能パリティ**まで戻し、本番を Go に切り替え、最終的に **backend-java/ と ECR の Java イメージを削除**して、コードベースを Go 一本に統一する（新規メンバーが「Java と Go のどちらが正か」で混乱しないようにする）。

## 3. スコープ (Goals / Non-Goals)

**Goals**
- Go を Java と同等のエンドポイント・機能までパリティ回復（§4 のギャップを全て埋める）
- archlint / go test グリーンを維持したまま実装（クリーンアーキテクチャ継続）
- 本番 ECS を Java イメージ → Go イメージへ切替（ダウンタイム最小）
- backend-java/ の削除 + ECR の Java イメージ削除

**Non-Goals**
- 新機能の追加（あくまで Java で動いていた機能の Go 再現に限定）
- インフラ構成（Fargate サイズ / 夜間 teardown / Supabase）の変更
- 教材コンテンツ（frestyle-teaching-materials）の変更

## 4. アーキテクチャ・設計詳細 (Architecture & Design)

### 4.1 機能パリティ ギャップ一覧（Java にあって Go に無い / 食い違う）

| # | 機能 | Java(現状) | Go(現状) | 対応 | 優先 |
|---|---|---|---|---|---|
| G1 | **コード実行: java 言語** | `code/execute` が java/php/go 対応(#1842,#1843) | `php/go/bash` のみ（**java 未対応**） | Go ランナーに java 実行(javac/java, クリーン環境+timeout)を追加 | **最高**（演習 java-1..8 が動く前提） |
| G2 | **認証パスの REST 化** | `/auth/login` `/auth/logout` `/auth/refresh` `/auth/me`(#1812, 単一 Hosted UI) | 旧 `/auth/cognito/callback` `/auth/cognito/logout` `/auth/cognito/refresh-token` `/auth/me` | Go の auth ルートを frontend が呼ぶ REST パスへ整合（Hosted UI 単一導線に合わせる） | **最高**（不一致だとログイン不能） |
| G3 | **courses / teaching-materials** | `/courses` `/courses/{id}` `/courses/{id}/materials` `/teaching-materials/{id}` | **削除済(#1831)** | git 履歴(#1831 の親)から復元し現行 Clean Arch 配置(usecase/repository + adapter/persistence)へ移植 | 高（/courses が 404 になる） |
| G4 | **learning reports** | `/learning-reports` `/learning-reports/generate`(SQS) | **削除済(#1833)** | git 履歴から復元・SQS enqueue を Go の adapter/persistence へ | 高 |
| G5 | **company settings(AI トグル)** | `/company/settings` GET/PUT + companies テーブル(V13) + AiChatAccess gate(#1851) | 無し | Go へ移植: companies(GORM AutoMigrate) + 設定 usecase + ai-chat 利用可否ミドルウェア | 高 |
| G6 | OpenAPI / 採点 | server 側採点一本化(#1845) | `master_exercise_example_repository.go` / `seed_master_exercises.go` 存在（採点ロジックは Go にあり） | 既存 Go 採点が複数テストケース対応か検証（不足あれば補完） | 中 |

> Go が既にパリティ達成済（対応不要）: auth(me/cookie 基盤) / ai-chat(sessions/messages/stream/attachments) / notes(+images) / notifications / profile(+image/onboarding) / exercises(list/detail/submit/submissions) / admin(company-applications/invitations) / public-invitations / company-applications / user-stats / session-note / embed(oembed) / health。

### 4.2 コード実行 java 対応(G1) の設計

Java backend の `ProcessCodeExecutor` と同等のサンドボックス方針を Go の `execute_code_usecase.go` に追加する:

- 一時ディレクトリに `Main.java` を書き出し、`javac Main.java` → `java Main` を **クリーン環境（環境変数除去）+ timeout + 非 root** で実行
- 既存 php の `disable_functions`/`open_basedir`、go の `go run` と同じく、untrusted コードを隔離
- `switch input.Language` に `case "java"` を追加
- Docker イメージ(`backend/Dockerfile`)に JDK を追加（現状 Go ランタイムのみ。`eclipse-temurin` 同様の JDK か `default-jdk` を apt 追加）
- 演習で使う言語の最終形: **php / go / bash / java**（docker は 'qa' モードで非実行、python は現状演習に無いので任意）

### 4.3 ECR / ECS 切替(Phase B)

- 本番 ECS タスク定義の image を **Java イメージ → Go イメージ**へ差し替え
- Go の `backend/Dockerfile`（JDK 追加版）で ECR に build/push（`CD - Backend Deploy to ECS` ワークフロー / infra `make ecr-build-push` + `make deploy-ecs`）
- 環境変数: Go backend が必要とするもの（`DATABASE_URL` / `COGNITO_*` / `BEDROCK_MODEL_ID` / `DYNAMODB_AI_CHAT_TABLE` / `NOTE_IMAGES_BUCKET` / SQS URL 等）を task def に整合
- ヘルスチェック: `GET /api/v2/health`（Go/Gin・既存）
- ロールバック: 切替後に問題が出たら **直前の Java タスク定義 revision に戻す**（task def は revision で保持されるため即時可能）

### 4.4 Java 資産削除(Phase C)

- `backend-java/` ディレクトリ削除
- ECR の **Java イメージ tag を削除**（`aws ecr batch-delete-image` / infra Makefile target 化）。⚠️ 破壊的操作 → infra CLAUDE.md §2.3 によりユーザー明示 GO 必須
- CI: `.github/workflows/ci-backend-java.yml` 等の Java 用ワークフロー削除、cd-backend を Go 用に一本化
- docs: ARCHITECTURE.md / README を Go 一本に整理

## 5. 検討した代替案 (Alternatives & Trade-offs)

| 案 | Pros | Cons | 判断 |
|---|---|---|---|
| **A. Go に戻す（採用）** | 極小 Fargate + 夜間 teardown に最適（即起動・低メモリ）/ Go 資産がほぼ残存 / archlint で CA 強制継続 | 削除済 2 機能の復元 + java 実行/認証パス整合が必要 | ✅ |
| B. Java を性能改善して維持 | 移植やり直し不要 | JVM 起動コストは構成上の制約で根治しづらい / Fargate 増強はコスト増（夜間 $0 方針と衝突） | ✗ |
| C. 両方併存し続ける | 段階移行できる | **新規メンバーが混乱**（どちらが正か不明）/ 二重メンテ | ✗（ユーザー明確に否） |

## 6. 運用・非機能要件 (Operations & Non-Functional)

- **パフォーマンス**: Go 化の主目的。起動時間・常駐メモリの削減（夜間 teardown 後の朝の復帰が速くなる）
- **セキュリティ**: コード実行サンドボックス（java 追加分も含め）クリーン環境 + timeout + 非 root + open_basedir/disable_functions 相当を維持
- **移行・リリース計画（ロールバック含む）**:
  - **Phase A（非破壊・無停止）**: 本番は Java のまま、Go に G1〜G6 を 1 機能 = 1 PR で実装。各 PR で archlint + go test グリーン必須
  - **Phase B（cutover・要 GO）**: ECS image を Go へ切替。問題時は Java task def revision へ即ロールバック
  - **Phase C（削除・要 GO）**: パリティ達成確認後に backend-java/ + ECR Java イメージ削除
- **監視・ログ**: 切替後 `GET /api/v2/health` 200 / CloudWatch ログ / 主要導線（ログイン・コース・演習実行・AI チャット）の本番スモーク

### PR 分割計画 / 進捗

**Phase A（Go の機能パリティ回復）は 2026-06-09 に完了。** 本番は Java 稼働のまま（非破壊）。
データアクセスは「読み取り＝生 SQL 直書き（db.Raw）/ 書き込み＝GORM」のハイブリッドで実装し、
sqlc の足場も導入済（`docs/sqlc-data-access.md`）。整形は gofumpt を CI で強制。

| PR | 内容 | フェーズ | 破壊性 | 状態 |
|---|---|---|---|---|
| 1 | G1: code/execute に java 追加 + Dockerfile に JDK | A | 無 | ✅ #1861 |
| 2 | G2: auth ルートを REST 化（frontend 整合） | A | 無 | ✅ #1875 |
| 3 | G3: courses/teaching-materials 復元 | A | 無 | ✅ #1879 |
| 4 | G4: learning-reports 復元 | A | 無 | ✅ #1880 |
| 5 | G5: company settings(AI トグル) 移植（companies に列を AutoMigrate） | A | 無 | ✅ #1881 |
| 6 | G6: 採点パリティ検証 + docs 更新 | A | 無 | ✅（本 PR） |
| 7 | ECS image を Go へ切替 | B | 本番 | ✅ 2026-06-09 完了（ECR の Go :latest へ force-new-deployment。health 200 / admin・courses endpoint が 404→401 で復活。Java 欠落の招待 404 バグも解消） |
| 8 | backend-java/ 削除 + README/テンプレを Go に + ECR Java イメージ削除 | C | 本番 | 🔄 本 PR（backend-java/ 削除・docs 更新）。ECR の Java イメージ掃除は別途 |

> **Phase B 補足（OIDC）**: CD ワークフロー(cd-backend.yml)の cfn-deploy ジョブが `environment: production` の OIDC subject で `sts:AssumeRoleWithWebIdentity` 不許可となり失敗したため、切替はローカル AWS CLI の `aws ecs update-service --force-new-deployment` で実施した（Go イメージの build/push は CD 側で成功済）。OIDC トラストポリシーの修正（`repo:...:environment:production` を許可）は IAM 変更のため infra 側で別途対応。

**G6 採点パリティの確認結果**: 演習の提出採点（`submit_master_exercise_usecase`）は Go から削除
されておらず、qa モード（提出文字列を normalize 比較）/ execute モード（examples を全件実行し
stdout を normalize 比較・全件 pass かつ exit 0 で正解）/ examples 無し時は exercise 自身の
ExpectedOutput を単一ケース化 / 一覧の published ゲート、を備えており **Java(#1845) と同等**。
補完の必要なし。

**次の一歩（要ユーザー GO）**: Phase B（ECS image を Go へ切替）。問題時は Java の task def
revision へ即ロールバック。切替・本番スモーク後に Phase C（Java 資産削除）。

## 7. テストプラン (Testing)

- **ユニット**: 各機能の usecase は interface モックで単体（testify）。code/execute java はクリーン環境・timeout の検証
- **結合**: handler を `httptest` + `gin.New()` で検証 / repository は sqlite メモリ。archlint・go vet・go test を CI 必須
- **手動・本番検証（Phase B 後）**: ログイン（単一 Hosted UI）/ `/courses` 一覧・詳細 / 演習 java/php/go/bash 実行 + 提出採点 / AI チャット SSE / 学習レポート生成 / AI トグル ON/OFF の trainee 表示。`https://normanblog.com` でスモーク
