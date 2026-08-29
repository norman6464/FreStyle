-- アプリケーション中核テーブルの DDL（実スキーマの正本）。
--
-- このファイルは 2 つの役割を 1 つの定義で兼ねる。
--
--   1. 実スキーマ  … 起動時に ApplyCoreSchema（infra/database/core_schema.go）が
--                    go:embed した本文をそのまま流す。デプロイ物とスキーマ定義が必ず同じ版になる。
--   2. sqlc の型付け … sqlc.yaml の schema 欄がこのファイルを直接読む。
--
-- 宣言（sqlc）と実体（DDL）が同じ 1 ファイルなので、片方だけ直してずれることが原理的に起きない。
-- ノート（schema/knowledge_base.sql / knowledge_base_permissions.sql）が先に採っている形に揃えてある。
--
-- 適用順序（infra/database/migrate.go の Migrate が守る）:
--   このファイル → seed / バックフィル / 明示制約（ApplyXxxConstraints）→ ノート → 権限モデル
--   → テナント橋渡し（companies.workspace_id / users.workspace_id）。
--   権限モデルは users を、テナント橋渡しは workspaces を参照するため順序は崩せない。
--
-- 冪等性: CREATE TABLE / CREATE INDEX はすべて IF NOT EXISTS。何度流しても安全に通る。
--   ただし CREATE TABLE IF NOT EXISTS は既に在るテーブルへ列を足さない。既存 DB の列追加・型変更・
--   NOT NULL 化は migrations/000X_*.sql（明示 SQL）で行う。ここは「新しく作る DB の姿」の定義。
--
-- ここに置かないもの:
--   - ノート（骨格 workspaces / spaces / pages / blocks / page_paths / page_snapshots、
--     権限モデル principals / principal_members / workspace_grants / space_grants /
--     page_restrictions / page_allow_lists / share_links）→ schema/knowledge_base*.sql が正本。
--   - FK / CHECK / 部分 UNIQUE のうち、既存データの修復を伴うもの（users の正規化まわり）
--     → migrate.go の ApplyUserNormalizationConstraints が、バックフィルの後に張る。
--   - テナント橋渡し列（companies.workspace_id / users.workspace_id）
--     → tenant_bridge.go が ALTER TABLE ADD COLUMN IF NOT EXISTS で足す（既存 DB にも届く経路）。
--       ALTER は必ず列を末尾に付けるので、下の CREATE TABLE でも該当テーブルの最後に書いてある
--       （並びがずれると SELECT * の詰め替えが位置ずれで壊れる）。
--
-- 型と NULL 可否の方針:
--   created_at / updated_at はアプリが必ず値を入れるため NOT NULL とする（sqlc が sql.NullTime では
--   なく time.Time を生成し、domain への詰め替えが素直になる）。同じ理由で、アプリが必ず値を入れる
--   文字列・数値も NOT NULL + DEFAULT を付ける。実際に NULL を取り得るもの（未所属の company_id、
--   論理削除の deleted_at、任意項目の hint_text など）だけ nullable にする。

-- 演習の入力例 / 期待出力例。(exercise_id, order_index) は同一問題内で一意。
CREATE TABLE IF NOT EXISTS master_exercise_examples (
    id              bigserial PRIMARY KEY,
    exercise_id     bigint NOT NULL,
    order_index     smallint NOT NULL,
    input_text      text NOT NULL DEFAULT '',
    expected_output text NOT NULL,
    created_at      timestamptz NOT NULL,
    updated_at      timestamptz NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_examples_exercise_order
    ON master_exercise_examples (exercise_id, order_index);

-- ロールマスタ。name はアプリのビジネス定数（super_admin 等）と一致する。
-- id は固定採番（1: super_admin / 2: company_admin / 3: trainee）で、採番列にしない。
-- 型は integer（本番の実列も integer）。
CREATE TABLE IF NOT EXISTS roles (
    id          integer PRIMARY KEY,
    name        text NOT NULL,
    description text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL,
    updated_at  timestamptz NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_name ON roles (name);

-- OIDC プロバイダ由来のユーザー識別子（Cognito の sub を users から分離）。
-- FK / CHECK は ApplyUserNormalizationConstraints が張る（既存データの修復を伴うため）。
CREATE TABLE IF NOT EXISTS user_oidc_identities (
    id         bigserial PRIMARY KEY,
    user_id    bigint NOT NULL,
    provider   text NOT NULL DEFAULT 'cognito',
    subject    text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_user_oidc_user_provider
    ON user_oidc_identities (user_id, provider);
CREATE UNIQUE INDEX IF NOT EXISTS uq_user_oidc_provider_subject
    ON user_oidc_identities (provider, subject);

-- 利用者。company_id / deleted_at は実際に NULL になり得る（運営管理者は会社無し等）。
-- role_id は roles マスタへの参照（正規化後の正）。DEFAULT 3 = trainee は、ローリングデプロイ中の
-- 旧コード（role_id を書かない INSERT）を NOT NULL 違反で壊さないための安全弁。
-- アクティブ行の email 部分 UNIQUE（uq_users_email_active）は
-- ApplyUserNormalizationConstraints が、正規形へのバックフィル後に張る。
CREATE TABLE IF NOT EXISTS users (
    id            bigserial PRIMARY KEY,
    email         text NOT NULL DEFAULT '',
    password_hash text,
    name          text NOT NULL DEFAULT '',
    company_id    bigint,
    role_id       integer NOT NULL DEFAULT 3,
    is_active     boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL,
    updated_at    timestamptz NOT NULL,
    deleted_at    timestamptz,
    -- workspace_id は tenant_bridge.go が末尾に足す列。実列の並びと合わせるため必ず最後に書く。
    -- 所属の正本は当面 company_id のままで、この列は写し。
    workspace_id  uuid
);

-- 学習メモ。
CREATE TABLE IF NOT EXISTS notes (
    id         bigserial PRIMARY KEY,
    user_id    bigint NOT NULL,
    title      text NOT NULL DEFAULT '',
    content    text NOT NULL DEFAULT '',
    is_public  boolean NOT NULL DEFAULT false,
    is_pinned  boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes (user_id);

-- users とは別管理のプロフィール拡張（user_id が PK）。
CREATE TABLE IF NOT EXISTS profiles (
    user_id        bigserial PRIMARY KEY,
    bio            text NOT NULL DEFAULT '',
    avatar_url     text NOT NULL DEFAULT '',
    status_message text NOT NULL DEFAULT '',
    updated_at     timestamptz NOT NULL
);

-- アプリ内通知。
CREATE TABLE IF NOT EXISTS notifications (
    id         bigserial PRIMARY KEY,
    user_id    bigint NOT NULL,
    type       text NOT NULL DEFAULT '',
    title      text NOT NULL DEFAULT '',
    body       text NOT NULL DEFAULT '',
    is_read    boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications (user_id);

-- 企業（現行のテナント）。
CREATE TABLE IF NOT EXISTS companies (
    id                           bigserial PRIMARY KEY,
    name                         text NOT NULL,
    is_active                    boolean NOT NULL DEFAULT true,
    created_at                   timestamptz NOT NULL,
    updated_at                   timestamptz NOT NULL,
    -- 対応する workspaces 行への橋渡し。tenant_bridge.go が末尾に足す列なので必ず最後に書く。
    -- テナントの正本を workspaces へ移す移行期間だけの列で、companies を畳むときに列ごと消える。
    workspace_id                 uuid
);

-- 公開フォームからの利用申請。
CREATE TABLE IF NOT EXISTS company_applications (
    id             bigserial PRIMARY KEY,
    company_name   varchar(200) NOT NULL,
    applicant_name varchar(120) NOT NULL,
    email          varchar(255) NOT NULL,
    message        text NOT NULL DEFAULT '',
    status         varchar(16) NOT NULL DEFAULT 'pending',
    created_at     timestamptz NOT NULL,
    updated_at     timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_company_applications_email ON company_applications (email);
CREATE INDEX IF NOT EXISTS idx_company_applications_status ON company_applications (status);

-- 教材コース。
CREATE TABLE IF NOT EXISTS courses (
    id                 bigserial PRIMARY KEY,
    company_id         bigint NOT NULL,
    created_by_user_id bigint NOT NULL,
    title              text NOT NULL DEFAULT '',
    description        text NOT NULL DEFAULT '',
    category           text NOT NULL DEFAULT '',
    language           varchar(50) NOT NULL DEFAULT '',
    sort_order         bigint NOT NULL DEFAULT 100,
    is_published       boolean NOT NULL DEFAULT false,
    created_at         timestamptz NOT NULL,
    updated_at         timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_courses_company_id ON courses (company_id);

-- 運営 / 管理者の重要操作の監査記録。
CREATE TABLE IF NOT EXISTS audit_events (
    id          bigserial PRIMARY KEY,
    actor_id    bigint NOT NULL,
    actor_email varchar(255) NOT NULL DEFAULT '',
    actor_role  varchar(32) NOT NULL DEFAULT '',
    action      varchar(160) NOT NULL DEFAULT '',
    target_id   bigint NOT NULL,
    created_at  timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_audit_events_actor_id ON audit_events (actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_action ON audit_events (action);
CREATE INDEX IF NOT EXISTS idx_audit_events_created_at ON audit_events (created_at);

-- 招待（マジックリンク）。token は未設定を NULL にして一意制約を避けるため nullable。
-- updated_at は持たない。
CREATE TABLE IF NOT EXISTS invitations (
    id         bigserial PRIMARY KEY,
    company_id bigint NOT NULL,
    email      text NOT NULL DEFAULT '',
    role       text NOT NULL DEFAULT '',
    name       text NOT NULL DEFAULT '',
    status     text NOT NULL DEFAULT '',
    token      varchar(64),
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_invitations_token ON invitations (token);
CREATE INDEX IF NOT EXISTS idx_invitations_company_id ON invitations (company_id);

-- 運営が用意した練習問題マスタ。
-- sort_order は migration 0011 が ALTER ADD COLUMN で integer として作った列（本番の実列も integer）。
-- hint_text / expected_output は任意項目で nullable。chapter_id も NULL を取り得る。
CREATE TABLE IF NOT EXISTS master_exercises (
    id              bigserial PRIMARY KEY,
    slug            varchar(64) NOT NULL,
    language        varchar(32) NOT NULL,
    sort_order      integer NOT NULL DEFAULT 0,
    category        varchar(64) NOT NULL,
    title           varchar(200) NOT NULL,
    description     text NOT NULL,
    starter_code    text NOT NULL,
    hint_text       text,
    expected_output text,
    mode            varchar(16) NOT NULL DEFAULT 'execute',
    explanation     text NOT NULL DEFAULT '',
    difficulty      smallint NOT NULL DEFAULT 1,
    is_published    boolean NOT NULL DEFAULT true,
    chapter_id      bigint,
    created_at      timestamptz NOT NULL,
    updated_at      timestamptz NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_master_exercises_slug ON master_exercises (slug);
CREATE INDEX IF NOT EXISTS idx_master_exercises_language ON master_exercises (language);

-- 会社独自の演習（論理削除あり）。
CREATE TABLE IF NOT EXISTS company_exercises (
    id              bigserial PRIMARY KEY,
    company_id      bigint NOT NULL,
    language        varchar(32) NOT NULL,
    title           varchar(200) NOT NULL,
    description     text NOT NULL,
    starter_code    text NOT NULL,
    hint_text       text,
    expected_output text,
    difficulty      smallint NOT NULL DEFAULT 1,
    is_published    boolean NOT NULL DEFAULT false,
    chapter_id      bigint,
    created_by      bigint NOT NULL,
    created_at      timestamptz NOT NULL,
    updated_at      timestamptz NOT NULL,
    deleted_at      timestamptz
);
CREATE INDEX IF NOT EXISTS idx_company_exercises_company_id ON company_exercises (company_id);
CREATE INDEX IF NOT EXISTS idx_company_exercises_language ON company_exercises (language);
CREATE INDEX IF NOT EXISTS idx_company_exercises_deleted_at ON company_exercises (deleted_at);

-- コード演習の提出履歴（append-only）。stdout / stderr は未取得のとき NULL。
CREATE TABLE IF NOT EXISTS exercise_submissions (
    id             bigserial PRIMARY KEY,
    user_id        bigint NOT NULL,
    exercise_kind  varchar(16) NOT NULL,
    exercise_id    bigint NOT NULL,
    submitted_code text NOT NULL,
    stdout         text,
    stderr         text,
    exit_code      bigint NOT NULL DEFAULT 0,
    is_correct     boolean NOT NULL DEFAULT false,
    submitted_at   timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_submissions_user_at
    ON exercise_submissions (user_id, submitted_at DESC);

-- 章の完了記録。(user_id, chapter_id) は複合 UNIQUE（同じ章の二重記録を防ぐ）。
CREATE TABLE IF NOT EXISTS user_chapter_progress (
    id           bigserial PRIMARY KEY,
    user_id      bigint NOT NULL,
    chapter_id   bigint NOT NULL,
    course_id    bigint NOT NULL,
    completed_at timestamptz NOT NULL,
    created_at   timestamptz NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_user_chapter_progress
    ON user_chapter_progress (user_id, chapter_id);
CREATE INDEX IF NOT EXISTS idx_user_chapter_progress_course_id
    ON user_chapter_progress (course_id);

-- 章の閲覧記録。PK = (user_id, chapter_id)。upsert で last_viewed_at / view_count を更新する。
-- view_count は migration 0005 の実テーブル定義に合わせて integer。
CREATE TABLE IF NOT EXISTS user_chapter_views (
    user_id         bigint NOT NULL,
    chapter_id      bigint NOT NULL,
    course_id       bigint NOT NULL,
    first_viewed_at timestamptz NOT NULL,
    last_viewed_at  timestamptz NOT NULL,
    view_count      integer NOT NULL DEFAULT 1,
    PRIMARY KEY (user_id, chapter_id)
);

-- リッチテキスト文書（tiptap JSON を jsonb で保持）。id はアプリ採番の uuid。
CREATE TABLE IF NOT EXISTS rich_documents (
    id             uuid PRIMARY KEY,
    owner_id       bigint NOT NULL,
    company_id     bigint,
    kind           text NOT NULL,
    title          text NOT NULL,
    is_public      boolean NOT NULL DEFAULT false,
    schema_version bigint NOT NULL DEFAULT 1,
    doc            jsonb NOT NULL,
    revision       bigint NOT NULL DEFAULT 1,
    created_at     timestamptz NOT NULL,
    updated_at     timestamptz NOT NULL,
    deleted_at     timestamptz,
    CONSTRAINT fk_rich_documents_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT ck_rich_documents_doc CHECK (jsonb_typeof(doc) = 'object' AND doc->>'type' = 'doc'),
    CONSTRAINT ck_rich_documents_title_len CHECK (char_length(title) <= 200)
);
CREATE INDEX IF NOT EXISTS idx_rich_documents_owner_id ON rich_documents (owner_id);

-- 日次の学習活動サマリー。PK = (user_id, activity_date)。書き込み時に upsert (+= delta)。
-- 各 *_count は migration 0005 の実テーブル定義に合わせて integer。
CREATE TABLE IF NOT EXISTS user_daily_activities (
    user_id        bigint NOT NULL,
    activity_date  date NOT NULL,
    exercise_count integer NOT NULL DEFAULT 0,
    correct_count  integer NOT NULL DEFAULT 0,
    chapter_count  integer NOT NULL DEFAULT 0,
    note_count     integer NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, activity_date)
);

-- score_cards は AI 評価スコアの旧テーブル（対応する domain 構造体は撤去済）。
-- user_stats_repository が集計元として user_id / overall_score を読むため定義を残す。
-- 列定義は撤去前の domain.ScoreCard と本番 pg_dump（いずれも git 履歴）から復元したもので、
-- id 以外は当時の実列に合わせ nullable。
CREATE TABLE IF NOT EXISTS score_cards (
    id                  bigserial PRIMARY KEY,
    user_id             bigint,
    session_id          bigint,
    overall_score       numeric,
    logical_score       numeric,
    consideration_score numeric,
    summary_score       numeric,
    proposal_score      numeric,
    listening_score     numeric,
    feedback            text,
    created_at          timestamptz
);
