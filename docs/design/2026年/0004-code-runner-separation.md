---
status: approved
area: infra
date: 2026-06-10
issue: norman6464/FreStyle#NN
---

> **決定（2026-06-10）**:
> - デプロイ構成は **A. 同一タスク内サイドカー（runner=`essential:false`、コスト増なし）** を採用。Phase 1（本体コード分離）→ Phase 2（infra サイドカー化 + コールドスタート実測、ユーザー GO で実反映）の順で進める。
> - **Java を撤去**（OOP は PHP で扱う）。JDK が最大の肥大化要因のため、サイドカー分離に先立って Java を実行系・UI・テストから削除し、現行の単一イメージを即縮小する（先行 PR）。対応言語は **php / go / bash**。
> - **エディタ入場時 warmup**: 実行時にランタイムを立ち上げるのではなく、学習者がコードエディタ（演習詳細）ページに入った時点で実行環境が warm であることを担保する。サイドカー runner は常駐なので原理的に warm。加えてエディタ入場で warmup ping を送り、Go のコンパイルキャッシュを温めて初回 Run も即時化し、UI に「実行環境 準備完了」を出す。

# 0004 コード実行サンドボックスを backend から分離してイメージを軽量化

> 関連: `backend/Dockerfile` / `backend/internal/usecase/execute_code_usecase.go` / `frestyle-infrastructure` runtime/ecs.yml

## 1. 基本情報

| 項目 | 内容 |
|---|---|
| タイトル | コード実行サンドボックス(code-runner)の分離とAPIイメージ軽量化 |
| 担当者 | @norman6464 |
| ステータス | Reviewing（実装着手はユーザー承認後） |
| 関連 | コールドスタート遅延の改善 |

## 2. 背景と目的

**背景**: 本番は夜間 teardown で ECS タスクを 0 に落としている（$0 化）。そのため**朝/初回アクセスがコールドスタート**になり、ユーザーから「最初だけ遅い（その後は速い）」と指摘があった。原因は **backend イメージが 474 MB（圧縮後）と大きく、タスク起動時の image pull に時間がかかる**こと。

肥大化の主因は、**API 本体と「学習者コードを実行する polyglot サンドボックス」が 1 つのイメージに同居**していること。`backend/Dockerfile` の runtime stage は `golang:1.26-bookworm`（Go ツールチェイン同梱）に加えて `php-cli` を apt-install しており、これらサンドボックス用ランタイムだけで数百 MB を占める。**当初は `default-jdk-headless`（JDK）も同梱していたが、JDK が最大の肥大化要因のため Java を撤去**した（先行 PR、OOP は PHP で扱う）。残る実行系は php / go / bash。

一方、**コールドスタートのクリティカルパス（ログイン / health / courses / notes）は go・php を一切必要としない**。これらは学習者が演習で「実行」を押したときにだけ要る。つまり、初回アクセスを速くするためにサンドボックス用ランタイムをクリティカルパスから外せる。

**目的**: API 本体イメージを distroless ベースの Go 単体（~60 MB 目標）まで絞り、コールドスタート時の pull を高速化する。重い polyglot ランタイムは別コンポーネント（code-runner）に分離し、コード実行が必要なときだけ使う。

## 3. スコープ (Goals / Non-Goals)

**Goals**
- backend API イメージを ~60 MB（distroless static Go）に縮小し、コールドスタートの image pull 時間を短縮する
- コード実行（php / go / bash）の os/exec ロジックを別バイナリ `cmd/coderunner` に切り出し、HTTP 経由で呼ぶ
- クリーンアーキテクチャを維持（handler→usecase→**CodeRunner port**→infra クライアント）
- サンドボックスが backend の機密 env（DATABASE_URL / Cognito secret / ECS Task Role 一時クレデンシャル）と**同一プロセスに同居しない**セキュリティ強化を副次的に得る

**Non-Goals**
- サンドボックスの言語追加・実行制限の仕様変更（現行の timeout / disable_functions / sandboxEnv はそのまま移設）
- 夜間 teardown の仕組み変更（runner も同じ stop/start サイクルに乗せる）
- 採点ロジック（ExecuteCodeOutput の契約）の変更

## 4. アーキテクチャ・設計詳細

### 概要設計（分離後）

```text
[Client] ──HTTP──> [backend API container]  distroless Go ~60MB（クリティカルパス）
                        │ handler → usecase(ExecuteCodeUseCase)
                        │              └ CodeRunner port(interface)
                        │                   └ infra coderunner.HTTPClient
                        ▼ POST http://127.0.0.1:9000/run （内部のみ）
                   [code-runner container]  golang+php+jdk ~474MB
                        └ os/exec で php / go run / bash を実行
                          （AWS 認証情報・DB secret を一切持たない env）
```

- **backend API**: `cmd/server`。distroless（`gcr.io/distroless/static`）に静的 Go バイナリのみ。go/php/bash ランタイムを含まない
- **code-runner**: `cmd/coderunner`。現行 Dockerfile の polyglot ランタイム（golang base + php-cli + GOCACHE warmup）を引き継ぐ。極小 HTTP サーバで `POST /run`（入力 = 現 `ExecuteCodeInput`、出力 = 現 `ExecuteCodeOutput`）、`POST /warmup`（言語指定で事前ウォームアップ）、`GET /healthz` を公開。**機密 env を注入しない**（タスク定義で runner コンテナには DB/Cognito/AWS の env を渡さない）

### 変更点（層ごと）

- **usecase**: `ExecuteCodeUseCase` を「os/exec を直接叩く」から「`CodeRunner` port に委譲する」へ変更。port = `Run(ctx, ExecuteCodeInput) (*ExecuteCodeOutput, error)` の 1 メソッド interface（`-er` 命名）。バリデーション（code サイズ / 言語判定）は usecase 側に残すか runner 側に寄せるかは実装時に決める（テスト容易性で usecase 残しが有力）
- **infra（新規）**: `internal/infra/coderunner/client.go` に HTTP クライアント実装。`CODE_RUNNER_URL`（既定 `http://127.0.0.1:9000`）へ POST。タイムアウトは言語別上限（go=15s 等）＋ネットワーク余裕を見て設定
- **サンドボックス本体（移設）**: 現 `execute_code_usecase.go` の `executePHP/Go/Bash` / `sandboxEnv` / `configureProcessGroup` / `runCommand` を runner パッケージ（`cmd/coderunner` 配下 or `internal/infra/sandbox`）へ移す。**ロジックは変更しない**（移送のみ）
- **handler**: 変更なし（`code_execute_handler.go` は usecase をそのまま呼ぶ）
- **Dockerfile**: backend を distroless 化。runner 用 `Dockerfile.coderunner` を新設（現行 Dockerfile を流用）
- **CI/CD（infra リポ）**: backend と coderunner の 2 イメージを build/push。ECS タスク定義に runner をサイドカーとして追加

### デプロイ構成（重要な決定点）

**推奨: 同一 ECS タスク内のサイドカー（runner は `essential: false`）**

- 1 タスク / 2 コンテナ / 2 イメージ。backend は `127.0.0.1:9000` の runner を呼ぶ
- runner を `essential: false` とし、backend は runner に `dependsOn` しない。**ALB ヘルスチェックは backend コンテナを叩く**ため、軽い backend イメージが先に pull・起動・healthy になればクリティカルパスは速くなる。runner（重いイメージ）はその裏で pull が続き、起動が遅れてもログイン等は影響を受けない
- runner 起動前のコード実行は、infra クライアントが runner `/healthz` 未達を検知して **503「実行環境を準備中です。数秒後に再度お試しください」** を返す（学習者が初回アクセス直後に Run を押す確率は低く許容範囲）
- **追加の常時稼働タスクが不要**＝コスト増なし。夜間 teardown も 1 タスク停止のまま
- メモリ: Java 撤去で JDK ヒープ要件が消えたため `mem512` でも収まる見込み。Go の `go run` コンパイル時メモリのみ考慮し、Phase 2 で実測して必要なら引き上げる（コスト影響を判断）

> ⚠️ **要実証**: Fargate が「軽い backend コンテナを、重い runner イメージの pull 完了を待たずに起動・ALB 登録できるか」は環境依存。Phase 2 の検証で計測し、もし pull がタスク全体を律速するなら下記「別サービス案」へ切り替える。

### データ構造 / API（内部）

```text
POST /run     （code-runner、内部ネットワークのみ）
  req:  { "code": string, "language": "php|go|bash", "stdin": string }
  res:  { "stdout": string, "stderr": string, "exitCode": int }
POST /warmup  （code-runner、内部）言語指定で事前ウォームアップ（Go の no-op compile 等）
  req:  { "language": "php|go|bash" }
  res:  { "ready": bool }
GET  /healthz （200 = 実行可能）
```

外部公開 API（`POST /api/v2/code/execute`）の契約・OpenAPI は不変。warmup 用に公開側 `POST /api/v2/code/warmup`（または既存 execute の薄い派生）を追加し、frontend はエディタ入場時に対象言語で 1 回叩く。

### エディタ入場時 warmup フロー

```text
[演習詳細ページ mount]
  └ frontend: POST /api/v2/code/warmup { language }
       └ backend usecase → CodeRunner.Warmup(language)
            └ runner: Go なら trivial main を 1 回 go run してコンパイルキャッシュを温める
                      php/bash は起動が軽量なので no-op（ready=true 即返し）
  → UI に「実行環境 準備完了」を表示。最初の Run はキャッシュ温で即時
```

サイドカー runner は常駐のため「エディタを開いた時点でランタイムは既に起動済み」を満たす。warmup は **Go のコンパイルキャッシュ温め**と **UI の準備完了表示**が主目的（実行時にコンテナ/ランタイムを起動する設計ではない）。

## 5. 検討した代替案

| 案 | Pros | Cons | 判定 |
|---|---|---|---|
| **A. サイドカー（推奨, runner=essential:false）** | 追加タスク無し＝コスト増なし / API イメージ軽量化を即得る / 機密分離 | Fargate の pull 並列性に依存（要実証）/ mem 引き上げの可能性 | ⭕ 採用（実証付き） |
| B. 別 ECS サービス（内部 service discovery） | API タスクは確実に slim イメージのみ pull / runner を独立スケール | 常時 +1 タスク（active 時間で ~$6/月）/ 別 stop/start 管理 / ネットワークホップ | △ A が pull 律速で失敗した場合のフォールバック |
| C. 実行ごとに on-demand Fargate RunTask / Lambda | アイドルコスト 0 / 最強の分離 | 実行ごとにコンテナ/JVM コールドスタート（数秒）/ 実装複雑 | ✕ 体験悪化 |
| D. 現状維持（単一イメージのまま圧縮のみ） | 変更最小 | Go ツールチェインは本質的に大きく、圧縮の余地が乏しい（Java 撤去後も golang base が残る） | ✕ 効果薄 |

## 6. 運用・非機能要件

- **監視・ログ**: runner コンテナの stdout を CloudWatch Logs に分離。`/healthz` を CloudWatch アラーム対象に追加検討
- **セキュリティ**: runner コンテナには DB/Cognito/AWS の env を**注入しない**。`sandboxEnv` による env フィルタは引き続き多層防御として残す。ネットワークは `127.0.0.1` のみ（サイドカー）で外部到達不可
- **パフォーマンス**: 目標はコールドスタート時の API 応答開始までの短縮。Phase 2 で「pull → ALB healthy までの秒数」を slim 化前後で計測する
- **移行・リリース計画**: Phase 1（本体: port 化 + runner バイナリ + 2 Dockerfile、機能は単一タスク内で完結）→ Phase 2（infra: タスク定義サイドカー化 + 計測、ユーザー GO で実反映）。ロールバックは ECS タスク定義 revision を戻すだけ（旧 474MB 単一イメージへ即復帰可）

## 7. テストプラン

- **ユニット**: `ExecuteCodeUseCase` を fake `CodeRunner` でモックし、言語ルーティング / バリデーション / 503 ハンドリングを検証
- **結合**: runner パッケージのサンドボックス実行テスト（現 `execute_code_usecase_test.go` / `execute_code_sandbox_test.go` を runner 側へ移送）。php/go/bash の stdout/exitCode、timeout、secret 非漏洩（sandboxEnv）を担保
- **手動・本番検証**: Phase 2 で「夜間 teardown → 朝のコールドスタート」を実測し、`/api/v2/health` 200 までの時間が短縮されたことを確認。演習画面で各言語の Run が成功することを確認
