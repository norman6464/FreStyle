# 0001: trainee 向け AI エージェント機能の利用可否トグル

> これは Design Doc テンプレートの最初の適用例（試作）です。実装着手前の提案であり、ステータスは Draft。
> 対応する Design Doc Issue: [#1846](https://github.com/norman6464/FreStyle/issues/1846)

## 1. 基本情報

| 項目 | 内容 |
|---|---|
| タイトル | trainee 向け AI エージェント機能の利用可否トグル |
| 担当者 | @norman6464 |
| レビュアー | @norman6464 |
| ステークホルダー | 受講企業の company_admin（学校の先生 / 企業のメンター）、受講者（trainee） |
| ステータス | Draft |
| 関連 | Issue（起票後にリンク） / backend-java README §AI チャット |

## 2. 背景と目的 (Background & Objective)

### 背景
現状、AI チャット（AI エージェント機能）は全 trainee が利用できる。しかし受講企業・学校によっては、
「研修の特定フェーズでは AI に頼らせたくない」「自社の方針として AI 利用を一旦オフにしたい」という
運用ニーズがある。今は company_admin（先生 / メンター）側にこれを制御する手段がない。

### 目的（ゴール）
company_admin が、自社の trainee に対して **AI エージェント機能を有効にするかどうか** を
設定画面（歯車アイコン）から切り替えられるようにする。trainee 側は、その設定に応じて
サイドバーの「AI」メニューの表示・非表示が変わり、無効時はサーバ側でも利用が遮断される。

## 3. スコープ (Goals / Non-Goals)

### Goals
- `companies` に「trainee への AI エージェント機能を有効にするか」のフラグを追加する
- company_admin 用の設定 API（取得 / 更新）を追加する
- 設定画面（歯車）に **チェックボックス**を追加（company_admin / super_admin のみ操作可）
- trainee のサイドバー「AI」項目を、フラグが有効なときだけ表示する
- フラグが無効な会社の trainee は、AI チャット系エンドポイントを **サーバ側でも 403** で遮断する（UI 非表示だけに頼らない）
- `/auth/me` レスポンスにフラグを含め、フロントが表示判定に使えるようにする

### Non-Goals
- trainee 1 人ごとの個別 ON/OFF（今回は **会社単位**のみ）
- 時間帯・コース単位での細かい制御（将来の拡張）
- AI 利用量のクォータ / 課金制御
- company_admin 自身の AI 利用制限（対象は trainee のみ）

## 4. アーキテクチャ・設計詳細 (Architecture & Design)

### 概要設計

```
company_admin                         trainee
   │ 設定(歯車) でチェック                │ ログイン
   ▼                                    ▼
PUT /api/v2/company/settings        GET /api/v2/auth/me
   │ {aiChatEnabledForTrainees}         │ → aiChatEnabledForTrainees を含む
   ▼                                    ▼
companies.ai_chat_enabled_for_trainees  Sidebar: trainee かつ flag=true のとき「AI」表示
   │                                    │
   └──────────── 同じ会社 ───────────────┘
                  │
   trainee が AI チャット API を叩くと:
   AiAccessGuard が「role=trainee かつ 会社 flag=false」を 403 で遮断
```

### データ構造（マイグレーション）

既存 `companies` テーブル（id / name / created_at / updated_at）に列を追加する。GORM 製の既存テーブルに
backend-java(Flyway) から後付けするため、**冪等な ALTER**（`ADD COLUMN IF NOT EXISTS`）にする。

```sql
-- V13__add_ai_chat_toggle_to_companies.sql
ALTER TABLE companies
    ADD COLUMN IF NOT EXISTS ai_chat_enabled_for_trainees BOOLEAN NOT NULL DEFAULT TRUE;
```

- **既定 `TRUE`**: 現状（全 trainee が利用可）を維持し、後方互換にする（既存ユーザーから機能を奪わない）

backend-java に `Company` エンティティ + `CompanyRepository` を新設する（現状未実装）。

```java
@Entity @Table(name = "companies")
class Company {
  @Id Long id;
  String name;
  @Column(name = "ai_chat_enabled_for_trainees") boolean aiChatEnabledForTrainees;
  Instant createdAt;
  Instant updatedAt;
}
```

### API 仕様

| メソッド | パス | 認可 | 説明 |
|---|---|---|---|
| `GET` | `/api/v2/company/settings` | company_admin / super_admin | 自社の設定（`aiChatEnabledForTrainees`）を取得 |
| `PUT` | `/api/v2/company/settings` | company_admin / super_admin | `{ "aiChatEnabledForTrainees": boolean }` で更新 |

加えて `GET /api/v2/auth/me`（`MeResponse`）に `aiChatEnabledForTrainees` を追加し、フロントの表示判定に使う。

### 変更点（touch する層）

- **backend-java**:
  - `entity/Company.java`（新規）、`repository/CompanyRepository.java`（新規）
  - `controller/CompanySettingsController.java`（新規, GET/PUT）+ `service/CompanySettingsService.java`（新規）
  - `dto/MeResponse`: `aiChatEnabledForTrainees` を追加
  - **AI アクセスガード**: AI チャット系（`/api/v2/ai-chat/**`）で「role=trainee かつ 自社 flag=false」を 403 にする。
    既存の `CurrentUserProvider` + 専用 `AiChatAccessPolicy`（service）で判定し、各 AI ハンドラの入口で呼ぶ
  - `db/migration/V13__...sql`（新規）/ `MigrationTest` を V13 に更新
- **frontend**:
  - 設定ページ（歯車）に「trainee への AI エージェント機能を有効にする」チェックボックス（company_admin/super_admin のみ）
  - `hooks/useSidebar.ts` / `components/layout/Sidebar.tsx`: `mainNavItems` の `ai` を、trainee かつ `aiChatEnabledForTrainees` のときだけ表示
  - `CompanySettingsRepository`（新規, GET/PUT）

## 5. 検討した代替案 (Alternatives & Trade-offs)

- **トグルの粒度: 会社単位 vs trainee 個別**
  - 採用: **会社単位**（`companies` の 1 列）。先生 / メンターが「クラス全体」を一括制御する想定に合致し、実装も最小
  - 却下: trainee 個別。Pros: 細かい制御。Cons: UI / データモデルが複雑化、今回の要望（クラス一括）には過剰
- **既定値: TRUE vs FALSE**
  - 採用: **TRUE**（後方互換。既存 trainee の体験を変えない）
  - 却下: FALSE（明示的 opt-in）。Pros: 「意図的に有効化」が明確。Cons: リリース時に全 trainee が一斉に AI を失い混乱する
- **遮断レイヤ: UI のみ vs サーバ側でも**
  - 採用: **UI 非表示 + サーバ側 403** の二段。Cons: 実装量増。Pros: API 直叩きを防げる（UI 非表示だけは認可ではない）
- **保存先: companies 列 vs 専用 company_settings テーブル**
  - 採用: **companies に 1 列**。設定が 1 つなら過剰な正規化は不要。将来設定が増えたら `company_settings` へ切り出す

## 6. 運用・非機能要件 (Operations & Non-Functional)

- **監視・ログ**: 設定更新時に `company_id` / 変更後値 / 操作者を info ログ。403 遮断は debug ログ（濫用調査用）
- **セキュリティ**:
  - 設定 API は **company_admin / super_admin のみ**。trainee が叩くと 403（IDOR 対策で「自社のみ」操作可）
  - AI 遮断は **サーバ側で認可**。UI 非表示はあくまで体験向上で、認可の本体はバックエンド
  - super_admin は会社跨ぎ可だが、設定対象は「自社（company_id）」に限定
- **パフォーマンス**: AI ガードは 1 リクエストにつき会社フラグ 1 回参照。`/auth/me` で取得済みの値をフロントが使うため追加往復なし。バックエンドのガードは `companies` の 1 行 SELECT（PK 引き、無視できる）
- **移行・リリース計画**:
  - V13 は `ADD COLUMN IF NOT EXISTS ... DEFAULT TRUE` で冪等。AutoMigrate 相当の追加系なので無停止
  - ロールバック: 列は残してもアプリ挙動に影響しない（フラグを見なくなるだけ）。緊急時はフロントの表示条件を撤去すれば即時に従来挙動へ戻せる
  - フロント・バックは **後方互換**（`/auth/me` に項目が無くても trainee は従来どおり AI 表示、という既定を置く）ので、デプロイ順序に強い制約はない（バックエンド先行が安全）

## 7. テストプラン (Testing)

TDD（テスト先行）で実装する。

- [ ] **ユニット/結合（backend-java, `@SpringBootTest` + H2）**:
  - `CompanySettingsController`: company_admin が GET/PUT で値を取得・更新できる / trainee は 403 / 他社は操作不可
  - AI ガード: 会社 flag=false の trainee は `/api/v2/ai-chat/**` が 403 / flag=true なら通る / company_admin は flag に関わらず通る
  - `MeResponse` に `aiChatEnabledForTrainees` が含まれる
  - `MigrationTest`: V13 まで適用される
- [ ] **フロント（vitest + RTL）**:
  - 設定ページ: company_admin にはチェックボックスが見える / trainee には見えない / トグルで PUT が呼ばれる
  - `Sidebar`: trainee かつ flag=true で「AI」項目が出る / flag=false で出ない / company_admin は常に出る
- [ ] **手動・本番検証**: company_admin で OFF → trainee 再ログインで「AI」消失 + AI API が 403 / ON で復活
