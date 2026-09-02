-- =====================================================================
-- FreStyle のスキーマ定義（実スキーマの正本）
-- =====================================================================
--
-- このファイルが唯一の正本で、2 つの役割を同時に果たす:
--
--   1. **起動時に流す DDL**。infra/database が go:embed で埋め込み、ECS の起動ごとに
--      冪等に適用する。デプロイ物とスキーマ定義が必ず同じ版になる
--   2. **sqlc の型付け入力**（backend/sqlc.yaml）。同じファイルから Go の型が起きるので、
--      宣言と実体がずれない
--
-- ## 並び順を崩さないこと
--
-- 上から順に流すことを前提に書いてある。外部キーの参照先は必ず自分より上にある。
--   Ⅰ 中核    … users / roles / companies / courses / exercises …
--   Ⅱ ノートの骨格 … workspaces / spaces / pages / blocks / page_paths / page_snapshots
--   Ⅲ ノートの権限 … principals / grants / share_links
--                    （Ⅰ の users へ FK を張るので Ⅰ より後でなければならない）
--
-- ただし**適用そのものは 2 回に分かれる**（Ⅰ と Ⅱ+Ⅲ）。あいだに seed と
-- バックフィルが挟まり、それらは Ⅰ の表を必要とし、Ⅱ+Ⅲ より先に済んでいる必要が
-- あるため。詳しくは infra/database/migrate.go の Migrate を読むこと。
--
-- ## 書き方の約束
--
-- - `CREATE TABLE IF NOT EXISTS` / `CREATE INDEX IF NOT EXISTS` だけで冪等にする
-- - **既存の表への列追加も、素の `ALTER TABLE` では書かない。** 列が既に在って
--   何もしない場合でも先に ACCESS EXCLUSIVE ロックを取り、毎回の起動でその表を
--   止めてしまう。カタログを見て足りないときだけ `ALTER` する `DO` ブロックにする
--   （実例: Ⅱ の spaces.visibility）
-- - 変更したら `make sqlc` で生成物も更新する
--
-- =====================================================================

-- =====================================================================
-- Ⅰ. 中核（users / roles / companies / courses / exercises …）
-- =====================================================================

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
--     share_links）→ schema/knowledge_base*.sql が正本。
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
--   文字列・数値も NOT NULL + DEFAULT を付ける。実際に NULL を取り得るもの（未所属の workspace_id、
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

-- 利用者。deleted_at は実際に NULL になり得る。workspace_id は運営管理者等の未所属で NULL。
-- role_id は roles マスタへの参照（正規化後の正）。DEFAULT 3 = trainee は、ローリングデプロイ中の
-- 旧コード（role_id を書かない INSERT）を NOT NULL 違反で壊さないための安全弁。
-- アクティブ行の email 部分 UNIQUE（uq_users_email_active）は
-- ApplyUserNormalizationConstraints が、正規形へのバックフィル後に張る。
CREATE TABLE IF NOT EXISTS users (
    id            bigserial PRIMARY KEY,
    email         text NOT NULL DEFAULT '',
    password_hash text,
    name          text NOT NULL DEFAULT '',
    role_id       integer NOT NULL DEFAULT 3,
    is_active     boolean NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL,
    updated_at    timestamptz NOT NULL,
    deleted_at    timestamptz,
    -- workspace_id は tenant_bridge.go が末尾に足す列。実列の並びと合わせるため必ず最後に書く。
    -- 所属の正本（company_id は撤去済み）。
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

-- 教材コース。
CREATE TABLE IF NOT EXISTS courses (
    id                 bigserial PRIMARY KEY,
    created_by_user_id bigint NOT NULL,
    title              text NOT NULL DEFAULT '',
    description        text NOT NULL DEFAULT '',
    category           text NOT NULL DEFAULT '',
    language           varchar(50) NOT NULL DEFAULT '',
    sort_order         bigint NOT NULL DEFAULT 100,
    is_published       boolean NOT NULL DEFAULT false,
    created_at         timestamptz NOT NULL,
    updated_at         timestamptz NOT NULL,
    -- workspace_id は節Ⅳが末尾に足す列。実列の並びと合わせるため必ず最後に書く。
    -- 所属の正本（company_id は撤去済み）。
    workspace_id       uuid
);

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
    id           bigserial PRIMARY KEY,
    email        text NOT NULL DEFAULT '',
    role         text NOT NULL DEFAULT '',
    name         text NOT NULL DEFAULT '',
    status       text NOT NULL DEFAULT '',
    token        varchar(64),
    expires_at   timestamptz NOT NULL,
    created_at   timestamptz NOT NULL,
    -- workspace_id は節Ⅳが末尾に足す列。実列の並びと合わせるため必ず最後に書く。
    -- invitations は company_id をまだ所属の正本として読み書きしており、この列は写し
    -- （撤去は後続のチケット）。毎起動の移送がレガシー行の workspace_id を埋め続ける。
    workspace_id uuid
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_invitations_token ON invitations (token);

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
    deleted_at      timestamptz,
    -- workspace_id は節Ⅳが末尾に足す列。実列の並びと合わせるため必ず最後に書く。
    -- 所属の正本（company_id は撤去済み）。
    workspace_id    uuid
);
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
    kind           text NOT NULL,
    title          text NOT NULL,
    is_public      boolean NOT NULL DEFAULT false,
    schema_version bigint NOT NULL DEFAULT 1,
    doc            jsonb NOT NULL,
    revision       bigint NOT NULL DEFAULT 1,
    created_at     timestamptz NOT NULL,
    updated_at     timestamptz NOT NULL,
    deleted_at     timestamptz,
    -- workspace_id は節Ⅳが末尾に足す列。実列の並びと合わせるため必ず最後に書く。
    -- 所属の正本（company_id は撤去済み）。未所属の文書もあるため nullable。
    workspace_id   uuid,
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


-- コースを構成する章。本文は doc(jsonb) が正本で、未移行の章は NULL。
CREATE TABLE IF NOT EXISTS course_chapters (
    id                 bigserial PRIMARY KEY,
    course_id          bigint NOT NULL,
    created_by_user_id bigint NOT NULL,
    title              text NOT NULL DEFAULT '',
    doc                jsonb,
    revision           bigint NOT NULL DEFAULT 1,
    schema_version     bigint NOT NULL DEFAULT 1,
    sort_order         bigint NOT NULL DEFAULT 100,
    is_published       boolean NOT NULL DEFAULT false,
    created_at         timestamptz NOT NULL,
    updated_at         timestamptz NOT NULL,
    -- workspace_id は節Ⅳが末尾に足す列。実列の並びと合わせるため必ず最後に書く。
    -- 所属の正本（company_id は撤去済み）。
    workspace_id       uuid
);
CREATE INDEX IF NOT EXISTS idx_course_chapters_course_id ON course_chapters (course_id);

-- 本番に欠けている NOT NULL と既定値を、定義（このファイルの上のほう）に合わせて埋める。
--
-- なぜ要るか: これらの表は ORM が作っていた頃のもので、ORM は既存の列へ後から
-- NOT NULL を付け直さない。DDL を正本にした今も、過去に作られた列だけが緩いままで、
-- 「定義では必須なのに本番では NULL を入れられる」状態が 60 列残っていた。
--
-- 手順は 3 段。**この順序でないと落ちる**:
--   1. 既定値を付ける（以後の INSERT で NULL が入らなくなる）
--   2. 既に入っている NULL を埋める（NOT NULL は 1 行でも NULL があると付けられない）
--   3. NOT NULL を付ける
--
-- 冪等にするため、カタログを見て「まだ足りないときだけ」実行する。
-- 素の ALTER TABLE を毎回流すと、何もしない場合でも ACCESS EXCLUSIVE ロックを取り、
-- 起動のたびにその表を止めてしまう（このファイルの冒頭の約束）。
DO $fill$
DECLARE
    r record;
BEGIN
    -- 1. 既定値（宣言にあるのに本番で欠けているものだけ）
    FOR r IN
        SELECT * FROM (VALUES
            ('audit_events', 'action', $$''::character varying$$),
            ('audit_events', 'actor_email', $$''::character varying$$),
            ('audit_events', 'actor_role', $$''::character varying$$),
            ('company_applications', 'message', $$''::text$$),
            ('invitations', 'email', $$''::text$$),
            ('invitations', 'name', $$''::text$$),
            ('invitations', 'role', $$''::text$$),
            ('invitations', 'status', $$''::text$$),
            ('notes', 'content', $$''::text$$),
            ('notes', 'is_public', $$false$$),
            ('notes', 'title', $$''::text$$),
            ('notifications', 'body', $$''::text$$),
            ('notifications', 'is_read', $$false$$),
            ('notifications', 'title', $$''::text$$),
            ('notifications', 'type', $$''::text$$),
            ('profiles', 'avatar_url', $$''::text$$),
            ('profiles', 'bio', $$''::text$$),
            ('users', 'email', $$''::text$$),
            ('users', 'name', $$''::text$$),
            -- workspaces は Ⅱ（ノート）が作る表なので、まっさらな DB では初回この時点でまだ無い。
            -- 下の EXISTS(table) 側の絞りで対象に上がらず no-op になる
            -- （Ⅱ の CREATE TABLE 自体が NOT NULL DEFAULT true で作るので、そもそも埋める必要が無い）。
            -- 本番は workspaces が既に存在するので、次回の起動でここが効く。
            ('workspaces', 'is_active', $$true$$)
        ) AS v(tbl, col, expr)
        WHERE EXISTS (
            -- 表そのものが無ければここで弾く。無いと NOT EXISTS(列の既定値) は「表が無い」でも
            -- 真になり、Ⅱ がまだ作っていない workspaces へ初回起動で ALTER をぶつけて落ちる
            -- （実 PostgreSQL で確認済み）。
            SELECT 1 FROM information_schema.tables t
            WHERE t.table_schema = 'public' AND t.table_name = v.tbl
        ) AND NOT EXISTS (
            SELECT 1 FROM information_schema.columns ic
            WHERE ic.table_schema = 'public' AND ic.table_name = v.tbl
              AND ic.column_name = v.col AND ic.column_default IS NOT NULL
        )
    LOOP
        EXECUTE format('ALTER TABLE %I ALTER COLUMN %I SET DEFAULT %s', r.tbl, r.col, r.expr);
    END LOOP;

    -- 2. 残っている NULL を埋めてから 3. NOT NULL を付ける
    FOR r IN
        SELECT * FROM (VALUES
            ('audit_events', 'action', $$''::character varying$$),
            ('audit_events', 'actor_email', $$''::character varying$$),
            ('audit_events', 'actor_id', $$0$$),
            ('audit_events', 'actor_role', $$''::character varying$$),
            ('audit_events', 'created_at', $$now()$$),
            ('audit_events', 'target_id', $$0$$),
            ('companies', 'created_at', $$now()$$),
            ('companies', 'updated_at', $$now()$$),
            ('company_applications', 'created_at', $$now()$$),
            ('company_applications', 'message', $$''::text$$),
            ('company_applications', 'updated_at', $$now()$$),
            ('company_exercises', 'created_at', $$now()$$),
            ('company_exercises', 'updated_at', $$now()$$),
            ('course_chapters', 'course_id', $$0$$),
            ('course_chapters', 'created_at', $$now()$$),
            ('course_chapters', 'updated_at', $$now()$$),
            ('courses', 'created_at', $$now()$$),
            ('courses', 'updated_at', $$now()$$),
            ('invitations', 'created_at', $$now()$$),
            ('invitations', 'email', $$''::text$$),
            ('invitations', 'expires_at', $$now()$$),
            ('invitations', 'name', $$''::text$$),
            ('invitations', 'role', $$''::text$$),
            ('invitations', 'status', $$''::text$$),
            ('master_exercise_examples', 'created_at', $$now()$$),
            ('master_exercise_examples', 'updated_at', $$now()$$),
            ('master_exercises', 'created_at', $$now()$$),
            ('master_exercises', 'updated_at', $$now()$$),
            ('notes', 'content', $$''::text$$),
            ('notes', 'created_at', $$now()$$),
            ('notes', 'is_pinned', $$false$$),
            ('notes', 'is_public', $$false$$),
            ('notes', 'title', $$''::text$$),
            ('notes', 'updated_at', $$now()$$),
            ('notes', 'user_id', $$0$$),
            ('notifications', 'body', $$''::text$$),
            ('notifications', 'created_at', $$now()$$),
            ('notifications', 'is_read', $$false$$),
            ('notifications', 'title', $$''::text$$),
            ('notifications', 'type', $$''::text$$),
            ('notifications', 'user_id', $$0$$),
            ('profiles', 'avatar_url', $$''::text$$),
            ('profiles', 'bio', $$''::text$$),
            ('profiles', 'status_message', $$''::text$$),
            ('profiles', 'updated_at', $$now()$$),
            ('rich_documents', 'created_at', $$now()$$),
            ('rich_documents', 'updated_at', $$now()$$),
            ('roles', 'created_at', $$now()$$),
            ('roles', 'updated_at', $$now()$$),
            ('user_chapter_progress', 'created_at', $$now()$$),
            ('user_chapter_views', 'course_id', $$0$$),
            ('user_chapter_views', 'first_viewed_at', $$now()$$),
            ('user_chapter_views', 'last_viewed_at', $$now()$$),
            ('user_oidc_identities', 'created_at', $$now()$$),
            ('user_oidc_identities', 'updated_at', $$now()$$),
            ('users', 'created_at', $$now()$$),
            ('users', 'email', $$''::text$$),
            ('users', 'name', $$''::text$$),
            ('users', 'updated_at', $$now()$$),
            ('workspaces', 'is_active', $$true$$)
        ) AS v(tbl, col, fill)
        WHERE EXISTS (
            SELECT 1 FROM information_schema.columns ic
            WHERE ic.table_schema = 'public' AND ic.table_name = v.tbl
              AND ic.column_name = v.col AND ic.is_nullable = 'YES'
        )
    LOOP
        EXECUTE format('UPDATE %I SET %I = %s WHERE %I IS NULL', r.tbl, r.col, r.fill, r.col);
        EXECUTE format('ALTER TABLE %I ALTER COLUMN %I SET NOT NULL', r.tbl, r.col);
    END LOOP;
END
$fill$;

-- invitations.status は Go 側では pending/accepted/canceled の3値だけを書くが、
-- DB 側には制約が無かった。既存データに想定外の値があれば追加せず警告に留める。
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_invitations_status') THEN
        IF EXISTS (
            SELECT 1 FROM invitations
            WHERE status NOT IN ('pending', 'accepted', 'canceled')
        ) THEN
            RAISE WARNING 'invitations.status に想定外の値があるため ck_invitations_status を作成できません';
        ELSE
            ALTER TABLE invitations ADD CONSTRAINT ck_invitations_status
                CHECK (status IN ('pending', 'accepted', 'canceled'));
        END IF;
    END IF;
END $$;

-- email をアプリと同じ正規形（lower + 前後空白除去）へ畳む。索引の式を正規形にするだけでは、
-- 生の値のまま残った既存行に対して正規形の一意性を守れない環境が残る。畳むと衝突する行だけは
-- 触らない（別人かもしれない 2 行を勝手に 1 つのアドレスへ寄せない）。btrim の文字集合は
-- Go の domain.EmailTrimCutset と同じものを明示列挙する。
UPDATE users u
   SET email = lower(btrim(u.email, E'\t\n\x0B\f\r '))
 WHERE u.deleted_at IS NULL
   AND u.email <> lower(btrim(u.email, E'\t\n\x0B\f\r '))
   AND NOT EXISTS (
       SELECT 1 FROM users o
        WHERE o.id <> u.id
          AND o.deleted_at IS NULL
          AND lower(btrim(o.email, E'\t\n\x0B\f\r ')) = lower(btrim(u.email, E'\t\n\x0B\f\r '))
   );

-- 招待の email も同じ正規形へ畳む（保留中のみ）。招待ゲートは正規形の OIDC メールで引くため、
-- 大文字混じり・空白付きのまま残った pending 行は「招待したのに見つからない」になる。
UPDATE invitations
   SET email = lower(btrim(email, E'\t\n\x0B\f\r '))
 WHERE status = 'pending'
   AND email <> lower(btrim(email, E'\t\n\x0B\f\r '));

-- 論理削除済みユーザーに紐付く identity を掃除する（SoftDelete 側でも消すが、過去データと
-- 削除処理の失敗に対する自己修復として毎起動流す）。放置すると同じ OIDC subject を占有され、
-- 再招待した本人がログインできなくなる。
DELETE FROM user_oidc_identities oi USING users u
 WHERE oi.user_id = u.id AND u.deleted_at IS NOT NULL;

-- roles.name: 空文字禁止。
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_roles_name_not_empty') THEN
        ALTER TABLE roles ADD CONSTRAINT ck_roles_name_not_empty CHECK (name <> '');
    END IF;
END $$;

-- users.role_id → roles.id。ロールマスタの行は参照されている限り消せない（RESTRICT 相当）。
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_role') THEN
        ALTER TABLE users ADD CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles(id);
    END IF;
END $$;

-- user_oidc_identities.user_id → users.id。ユーザーの物理削除で identity も消す。
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_oidc_identities_user') THEN
        ALTER TABLE user_oidc_identities
            ADD CONSTRAINT fk_user_oidc_identities_user
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- identity の provider / subject: 空文字禁止。
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_user_oidc_identities_not_empty') THEN
        ALTER TABLE user_oidc_identities
            ADD CONSTRAINT ck_user_oidc_identities_not_empty CHECK (provider <> '' AND subject <> '');
    END IF;
END $$;

-- users.email: アクティブ行（未論理削除）かつ正規形が非空に限った部分 UNIQUE。論理削除→同メール
-- 再招待と両立し、email claim の無い OIDC ユーザー（空文字）は対象外にする。キーは email その
-- ものではなく上の UPDATE と同じ正規形 lower(btrim(email, ...))。既存データに（畳んでも解決
-- できない）重複がある場合は作成せず警告に留める（起動を落とさず、修正は運用判断に委ねる）。
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_indexes WHERE indexname = 'uq_users_email_active'
          AND indexdef NOT LIKE '%btrim%'
    ) AND NOT EXISTS (
        SELECT 1 FROM users
        WHERE deleted_at IS NULL AND btrim(email, E'\t\n\x0B\f\r ') <> ''
        GROUP BY lower(btrim(email, E'\t\n\x0B\f\r ')) HAVING count(*) > 1
    ) THEN
        DROP INDEX uq_users_email_active;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'uq_users_email_active') THEN
        IF EXISTS (
            SELECT 1 FROM users
            WHERE deleted_at IS NULL AND btrim(email, E'\t\n\x0B\f\r ') <> ''
            GROUP BY lower(btrim(email, E'\t\n\x0B\f\r ')) HAVING count(*) > 1
        ) THEN
            RAISE WARNING 'users.email に（大小文字・前後空白を無視した）重複があるため uq_users_email_active を作成できません（重複を解消して再起動してください）';
        ELSE
            CREATE UNIQUE INDEX uq_users_email_active
                ON users (lower(btrim(email, E'\t\n\x0B\f\r ')))
                WHERE deleted_at IS NULL AND btrim(email, E'\t\n\x0B\f\r ') <> '';
        END IF;
    END IF;
END $$;

-- rich_documents: owner_id → users.id。ユーザーの物理削除で文書も消す（論理削除運用なので通常は
-- 発火しない）。存在判定は conrelid（テーブル）でも絞る（制約名は表単位でしか一意でないため）。
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_rich_documents_owner' AND conrelid = 'rich_documents'::regclass) THEN
        ALTER TABLE rich_documents
            ADD CONSTRAINT fk_rich_documents_owner
            FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- rich_documents.doc は tiptap のドキュメント JSON（object かつ type='doc'）に限る。
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_rich_documents_doc' AND conrelid = 'rich_documents'::regclass) THEN
        ALTER TABLE rich_documents
            ADD CONSTRAINT ck_rich_documents_doc
            CHECK (jsonb_typeof(doc) = 'object' AND doc->>'type' = 'doc');
    END IF;
END $$;

-- rich_documents.title 長の上限（アプリ側検証と二重の壁）。
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_rich_documents_title_len' AND conrelid = 'rich_documents'::regclass) THEN
        ALTER TABLE rich_documents
            ADD CONSTRAINT ck_rich_documents_title_len
            CHECK (char_length(title) <= 200);
    END IF;
END $$;

-- =====================================================================
-- Ⅱ. ノートの骨格（workspaces / spaces / pages / blocks / page_paths / page_snapshots）
-- =====================================================================

-- ノート（workspaces / spaces / pages / blocks / page_paths / page_snapshots）の DDL。
--
-- このファイルが実スキーマの正本であり、同時に sqlc の型付け入力でもある
-- （backend/sqlc.yaml の schema に登録済み。列を足したら `make sqlc` で生成物を作り直す）。
-- 起動時に database.ApplyKnowledgeBaseSchema がこの内容をそのまま流す（冪等）。
--
-- アプリ中核テーブル（users / notes / courses / rich_documents …）の schema/core.sql と同じ扱いで、
-- 「宣言（sqlc）と実体（DDL）を 1 ファイルに集約して正本を 1 つにする」という方針を共有する。
-- 適用順は core → 骨格 → 権限（権限側が users を参照するため崩せない）。
--
-- 冪等性は CREATE ... IF NOT EXISTS だけで成り立たせ、DO ブロックは書かない。
-- sqlc がこのファイルをパースして型を作るので、素の DDL に保つ必要がある
-- （手続き型の DO ブロックが混ざると sqlc がパースできない）。
--
-- 注意（開発者のローカル DB）: CREATE TABLE IF NOT EXISTS は「テーブルが無いときだけ作る」ので、
-- 既に別定義のテーブルがある DB では何もしない。古い定義が残ったローカル DB は
-- `docker compose down -v` で作り直す（このテーブル群は未リリースで本番にはまだ存在しない）。
--
-- 設計の柱は 2 つ:
--
--   (1) 境界越えを DB で塞ぐ。親子の FK は必ず「入れ物」の列を含む複合 FK にし、
--       別のテナント / スペース / ページの行を親にできないようにする。
--       木はそれぞれの入れ物の中で閉じる: ページの木はスペースの中、ブロックの木はページの中。
--       入れ物をまたぐ親子を許すと、入れ物を消したときに ON DELETE CASCADE が
--       別の入れ物に残るはずの行まで道連れにする。
--       そのために参照先へ (workspace_id, …, id) の複合 UNIQUE を張る。id 単独の PK では
--       FK の参照列に複数列を指定できないため、実データ上は冗長でも足場として要る。
--
--   (2) 並び順は分数インデックス（internal/pkg/fracindex）が採番する文字列キー。
--       同じ親の中で position が重複しないことを部分 UNIQUE で守り、既定値は置かない（採番はアプリ側）。

-- ワークスペース: ノートのテナント境界。
CREATE TABLE IF NOT EXISTS workspaces (
    id         uuid PRIMARY KEY,
    -- slug は URL に出る短い識別子。テナント内ではなくグローバルに一意。
    slug       varchar(64) NOT NULL,
    name       varchar(200) NOT NULL,
    is_active  boolean NOT NULL DEFAULT true,
    -- 個人サインアップで自動作成した、その人専用のワークスペース。1 人 1 つ
    -- （uq_workspaces_personal_owner）。列と制約は workspace_ownership_schema.go が
    -- ALTER TABLE ADD COLUMN IF NOT EXISTS で足す（本番の既存 workspaces へ列を
    -- 届ける経路がそれしか無いため）。CREATE TABLE 側にも書くのは、まっさらな DB では
    -- ここで最初から作られるようにするため。
    personal_owner_user_id bigint,
    -- GORM の autoCreateTime / autoUpdateTime は使わないため、DB 側の既定値で必ず埋まるようにする。
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT uq_workspaces_slug UNIQUE (slug),
    -- URL に出る識別子は空文字禁止・長さ上限（アプリ側検証と二重の壁）。
    CONSTRAINT ck_workspaces_slug_len CHECK (char_length(slug) BETWEEN 1 AND 64)
);

-- スペース: ワークスペース内のページの束（部門・プロジェクト単位の入れ物）。
CREATE TABLE IF NOT EXISTS spaces (
    id           uuid PRIMARY KEY,
    workspace_id uuid NOT NULL,
    -- key はワークスペース内で一意な短い識別子（例: "eng"）。
    "key"        varchar(64) NOT NULL,
    name         varchar(200) NOT NULL,
    -- visibility はワークスペース既定の grant が届くか（'workspace'）・届かないか（'private'）。
    -- 'private' のスペースにはスペース単位の付与（space_grants）だけが届く。
    -- 「プライベートかどうか」を grant の構成から導出しないための明示の印（値の正本は
    -- domain.SpaceVisibility）。実効権限の畳み方は変えず、事実の集め方（workspace_grants を
    -- 参照する各クエリ）がこの列でふるう。
    visibility   varchar(16) NOT NULL DEFAULT 'workspace',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    -- ワークスペースの物理削除で配下も消える（運用ではアーカイブを使う想定で、物理削除は例外的な操作）。
    CONSTRAINT fk_spaces_workspace FOREIGN KEY (workspace_id)
        REFERENCES workspaces (id) ON DELETE CASCADE,
    CONSTRAINT uq_spaces_workspace_key UNIQUE (workspace_id, "key"),
    -- pages からの複合 FK の参照先。id の PK があるので実データ上は冗長だが、
    -- 「テナント越えを FK で塞ぐ」ための足場として要る。
    CONSTRAINT uq_spaces_workspace_id UNIQUE (workspace_id, id),
    CONSTRAINT ck_spaces_key_len CHECK (char_length("key") BETWEEN 1 AND 64),
    CONSTRAINT ck_spaces_visibility CHECK (visibility IN ('workspace', 'private'))
);

-- 既存 DB への visibility 列と CHECK の追加（新規 DB には上の CREATE TABLE が効く）。
--
-- **カタログを見て、足りないときだけ ALTER を出す。** ALTER TABLE は
-- ADD COLUMN IF NOT EXISTS で「列が既に在るから何もしない」場合でも、判定より先に
-- ACCESS EXCLUSIVE ロックを取り、トランザクションが終わるまで手放さない。この DDL は
-- 起動時マイグレーションの 1 トランザクションで流れるので、素で書くと**毎回の起動
-- （毎デプロイ）で spaces への読み書きが後続の DDL が終わるまで止まる**。
-- 同じファイルの CREATE INDEX を「既にある索引は発行しない」形にしているのと同じ理由で、
-- 列が出揃った通常の起動ではロックを 1 つも取らずに通り抜けさせる。
DO $mig$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'spaces' AND column_name = 'visibility'
    ) THEN
        -- 既定 'workspace' で埋まるので、追加時点の既存スペースの見え方は変わらない。
        ALTER TABLE spaces ADD COLUMN visibility varchar(16) NOT NULL DEFAULT 'workspace';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'ck_spaces_visibility' AND conrelid = 'spaces'::regclass
    ) THEN
        ALTER TABLE spaces ADD CONSTRAINT ck_spaces_visibility
            CHECK (visibility IN ('workspace', 'private'));
    END IF;
END
$mig$;

-- ページ: ノートの 1 ページ。parent_id の自己参照で木をなす（無限入れ子）。
CREATE TABLE IF NOT EXISTS pages (
    id                 uuid PRIMARY KEY,
    workspace_id       uuid NOT NULL,
    space_id           uuid NOT NULL,
    -- parent_id が NULL ならスペース直下（ルート）。
    parent_id          uuid,
    -- position のコレーションは "C"（バイト順）に固定する。
    -- 分数インデックスは「文字列の辞書順 = 並び順」が前提で、Go 側はバイト比較で判断する。
    -- DB の既定がロケール依存のコレーション（例: en_US.utf8）だと 'a' < 'B' のように並び、
    -- ORDER BY position がアプリの認識とずれる。列の定義で最初から揃えておく。
    "position"         text COLLATE "C" NOT NULL,
    title              varchar(200) NOT NULL DEFAULT '',
    -- 作成者（users.id）。users への FK は張らない（ノートの骨格に閉じるため）。
    created_by_user_id bigint NOT NULL,
    -- archived_at が NULL の行が現役。物理削除ではなくアーカイブで隠す運用のため、
    -- position の一意性はアーカイブ済みを除外した部分 UNIQUE で守る（下の CREATE UNIQUE INDEX）。
    archived_at        timestamptz,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),

    -- ページは「同じワークスペースの space」にしか属せない。
    CONSTRAINT fk_pages_space FOREIGN KEY (workspace_id, space_id)
        REFERENCES spaces (workspace_id, id) ON DELETE CASCADE,
    -- 親は「同じワークスペースの、同じスペースの」ページだけ。親の物理削除で子孫も消える。
    -- ページの木はスペースの中で閉じる。スペースはページの入れ物であり、木がスペースをまたぐと
    -- パンくず（祖先をたどると別スペースに出る）・サブツリー一括取得・スペース単位の権限が
    -- すべて破綻するため、space_id まで一致を要求する。
    -- workspace だけの一致だと、スペース A のページがスペース B のページを親に持ててしまい、
    -- スペース B を消したときに fk_pages_space の CASCADE で B のページが消え、続けて
    -- こちらの CASCADE がスペース A に残るはずの子ページまで道連れにする。
    --
    -- parent_id は NULL 可（ルート）。複合 FK は既定の MATCH SIMPLE なので、
    -- 参照列に 1 つでも NULL があれば検査自体が行われない ＝ ルートページは素通りする。
    -- これは意図どおり: ルートの workspace_id / space_id は fk_pages_space 側で必ず検査されるため、
    -- テナント越え・スペース越えの抜け道にはならない。
    --
    -- 副作用（意図した挙動）: ページを別スペースへ移すときは、子孫の space_id も同じ文で
    -- 更新しないと FK 違反になる。木の一部だけがスペースをまたぐ「中途半端な移動」を DB が防ぐ。
    CONSTRAINT fk_pages_parent FOREIGN KEY (workspace_id, space_id, parent_id)
        REFERENCES pages (workspace_id, space_id, id) ON DELETE CASCADE,
    -- blocks / page_paths からの複合 FK の参照先。space_id を持たないテーブルから
    -- ページを参照するには (workspace_id, id) の形が要る（下の fk_blocks_page /
    -- fk_page_paths_page / fk_page_paths_ancestor の 3 本が使う）。
    CONSTRAINT uq_pages_workspace_id UNIQUE (workspace_id, id),
    -- 親ページの FK を「同じスペース」まで絞るための足場。
    CONSTRAINT uq_pages_workspace_space_id UNIQUE (workspace_id, space_id, id),
    -- 自分自身を親にできない（1 行で閉じた循環を作らせない。多段の循環はアプリ側で検出する）。
    CONSTRAINT ck_pages_parent_not_self CHECK (parent_id IS NULL OR parent_id <> id),
    -- position は空文字だと順序として意味を持たない（fracindex は空文字を返さない）。
    CONSTRAINT ck_pages_position_not_empty CHECK ("position" <> '')
);

-- ブロック: ページ本文を構成する 1 行（段落・見出し・リスト項目・表のセル …）。
-- 入れ子（リストや表）は parent_id の自己参照で表す。
CREATE TABLE IF NOT EXISTS blocks (
    id           uuid PRIMARY KEY,
    workspace_id uuid NOT NULL,
    page_id      uuid NOT NULL,
    -- parent_id が NULL ならページ直下（トップレベル）。
    parent_id    uuid,
    -- pages.position と同じ理由でバイト順に固定する。
    "position"   text COLLATE "C" NOT NULL,
    -- ProseMirror（tiptap）のノード名。値は domain.BlockType が正。
    type         varchar(32) NOT NULL,
    -- ProseMirror の attrs（見出しの level、コードブロックの language など）。
    -- 属性が無いノードでも空オブジェクト {} を入れる（NULL と {} の二通りを作らない）。
    attrs        jsonb NOT NULL DEFAULT '{}'::jsonb,
    -- 葉ノードのインライン内容（text ノードとマークの配列）。
    -- リストや表のような容器ノードは子をブロック行として持つため NULL にする。
    inline       jsonb,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    -- ブロックは「同じワークスペースの page」にしか属せない。
    CONSTRAINT fk_blocks_page FOREIGN KEY (workspace_id, page_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    -- 親は「同じワークスペースの、同じページの」ブロックだけ。
    -- ブロックの木は 1 ページの中で閉じるものなので、page_id まで一致を要求する。
    -- workspace だけを一致させると、ページ A のブロックをページ B のブロックの親にでき、
    -- ページ A を消したときに ON DELETE CASCADE がページ B の本文まで消してしまう。
    -- MATCH SIMPLE の扱いは pages と同じで、parent_id が NULL（トップレベル）なら検査されない。
    -- その場合の workspace_id / page_id の正しさは fk_blocks_page 側で担保される。
    CONSTRAINT fk_blocks_parent FOREIGN KEY (workspace_id, page_id, parent_id)
        REFERENCES blocks (workspace_id, page_id, id) ON DELETE CASCADE,
    -- 親ブロックの FK を「同じページ」まで絞るための足場。
    CONSTRAINT uq_blocks_workspace_page_id UNIQUE (workspace_id, page_id, id),
    CONSTRAINT ck_blocks_parent_not_self CHECK (parent_id IS NULL OR parent_id <> id),
    CONSTRAINT ck_blocks_position_not_empty CHECK ("position" <> ''),
    -- attrs は ProseMirror の attrs なので必ず object（属性が無いノードでも {}）。
    CONSTRAINT ck_blocks_attrs_object CHECK (jsonb_typeof(attrs) = 'object'),
    -- inline は葉ノードの content 配列。容器ノードでは NULL にする。
    CONSTRAINT ck_blocks_inline_array CHECK (inline IS NULL OR jsonb_typeof(inline) = 'array')
);

-- page_paths: ページの祖先関係を平らに持つ派生テーブル（closure table）。自分自身も depth=0 の行として持つ。
-- pages.parent_id の連鎖だけでも木は表せるが、パンくず・サブツリー一括取得・移動時の循環検出を
-- 再帰クエリなしの 1 回の JOIN で済ませるためにこの索引を別に持つ。正本は pages.parent_id 側。
CREATE TABLE IF NOT EXISTS page_paths (
    workspace_id uuid NOT NULL,
    page_id      uuid NOT NULL,
    ancestor_id  uuid NOT NULL,
    -- 祖先までの距離。自分自身が 0、親が 1。
    depth        integer NOT NULL,

    CONSTRAINT page_paths_pkey PRIMARY KEY (page_id, ancestor_id),
    -- 1 行で「子孫」と「祖先」の 2 ページを組にするため、単独 FK を 2 本張るだけでは
    -- 別ワークスペースの 2 ページを組にした行が作れてしまう（両方の FK を通ってしまう）。
    -- 行自身の workspace_id を軸にした複合 FK にして、組になる 2 ページが同じワークスペースに
    -- 属することを DB 側で保証する。ページが消えたら派生であるこの行も一緒に消す。
    --
    -- FK で守るのは「組になる 2 ページが実在し、同じワークスペースに属すること」まで。
    -- 1 行だけで判定できる depth の不変条件は下の ck_page_paths_depth で別に塞ぐ。
    --
    -- 一方「depth が実際の親子の距離と一致するか、祖先の連鎖に抜けや余りが無いか」は DB では守らない。
    -- それは 1 行を見ても判定できず、pages の木をたどって初めて分かる複数行にまたがる不変条件で、
    -- 宣言的な制約（行ごとの CHECK / FK）では表せないため。この表は pages.parent_id から導ける
    -- 派生データなので、正本である pages 側の制約で木の形を守り、closure 全体の整合は行を書く側の
    -- 責務とする。なお page_paths は常に FK の子側で、この表の行が壊れても他の行を CASCADE で
    -- 消すことはない（壊れ方が表示の乱れに閉じ、他のデータを失わない）。
    CONSTRAINT fk_page_paths_page FOREIGN KEY (workspace_id, page_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_page_paths_ancestor FOREIGN KEY (workspace_id, ancestor_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    -- 1 行だけで判定できる不変条件: depth は祖先までの距離なので負にならず、
    -- depth=0 の行は自分自身を指す行「だけ」（逆に自己参照の行は必ず depth=0）。
    -- パンくずは ORDER BY depth で組み立てるため、ここが崩れると pages.parent_id（正本）は
    -- 正しいのに表示だけが壊れ、原因を追いにくい形で顕在化する。
    CONSTRAINT ck_page_paths_depth CHECK (depth >= 0 AND (depth = 0) = (page_id = ancestor_id))
);

-- page_snapshots: ページのブロック行を組み直した ProseMirror ドキュメント（読み取り用のキャッシュ）。
-- 表示のたびにブロック行を木に組み直すと 1 ページで数百行の取得と再帰的な組み立てが要るため、
-- 編集のたびに 1 つの jsonb へ焼き直して読み出しを 1 行の取得に落とす。
-- 正本はあくまで blocks 側で、この行は失っても blocks から再生成できる派生データ。
CREATE TABLE IF NOT EXISTS page_snapshots (
    page_id  uuid PRIMARY KEY,
    -- tiptap の getJSON() 相当（type='doc' の ProseMirror ドキュメント）。
    doc      jsonb NOT NULL,
    -- 焼き直した時刻。ブロックの更新時刻より古ければ作り直す判断に使う。
    built_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_page_snapshots_page FOREIGN KEY (page_id)
        REFERENCES pages (id) ON DELETE CASCADE,
    -- 壊れた snapshot は読み取りキャッシュとしてそのまま返り、エディタがページを開けなくなるため、
    -- rich_documents.doc と同じ形で入口を塞ぐ。
    CONSTRAINT ck_page_snapshots_doc CHECK (jsonb_typeof(doc) = 'object' AND doc->>'type' = 'doc')
);

-- --- 並び順の一意性（部分 UNIQUE は WHERE 付きなのでテーブル定義に書けず、索引として張る）---
-- 同じ親の中で position が重複しないこと。ページはアーカイブ済みを除外する
-- （アーカイブは「一覧から隠す」だけで行は残るため、現役の並びだけを守る）。
CREATE UNIQUE INDEX IF NOT EXISTS uq_pages_parent_position
    ON pages (parent_id, "position") WHERE archived_at IS NULL;
-- ルート直下（parent_id IS NULL）は上の索引では守れない。UNIQUE 索引は NULL 同士を
-- 別物として扱うため、parent_id が NULL の行同士は何度でも同じ position を持ててしまう。
-- ルートの並びはスペース単位なので、スペースを軸にした部分 UNIQUE を別に張る。
CREATE UNIQUE INDEX IF NOT EXISTS uq_pages_space_position
    ON pages (space_id, "position") WHERE parent_id IS NULL AND archived_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_blocks_parent_position
    ON blocks (parent_id, "position");
-- ブロックも同じ理由で、ページ直下（parent_id IS NULL）はページを軸に守る。
CREATE UNIQUE INDEX IF NOT EXISTS uq_blocks_page_position
    ON blocks (page_id, "position") WHERE parent_id IS NULL;

-- --- 取得経路の索引（FK の子側は自動では索引が張られないため明示する）---
-- ワークスペース / スペース単位の一覧と、親を辿る取得・CASCADE 削除の走査に効かせる。
CREATE INDEX IF NOT EXISTS idx_spaces_workspace_id ON spaces (workspace_id);
CREATE INDEX IF NOT EXISTS idx_pages_workspace_id ON pages (workspace_id);
CREATE INDEX IF NOT EXISTS idx_pages_space_id ON pages (space_id);
CREATE INDEX IF NOT EXISTS idx_pages_parent_id ON pages (parent_id);
-- アーカイブ済みの除外・アーカイブ一覧の取得に使う。
CREATE INDEX IF NOT EXISTS idx_pages_archived_at ON pages (archived_at);
CREATE INDEX IF NOT EXISTS idx_blocks_workspace_id ON blocks (workspace_id);
CREATE INDEX IF NOT EXISTS idx_blocks_page_id ON blocks (page_id);
CREATE INDEX IF NOT EXISTS idx_blocks_parent_id ON blocks (parent_id);
CREATE INDEX IF NOT EXISTS idx_page_paths_workspace_id ON page_paths (workspace_id);
-- 祖先からサブツリーを引く経路（PK は (page_id, ancestor_id) なので ancestor_id 単独では効かない）。
CREATE INDEX IF NOT EXISTS idx_page_paths_ancestor_id ON page_paths (ancestor_id);

-- =====================================================================
-- Ⅲ. ノートの権限（principals / grants / share_links）
-- =====================================================================

-- ノートの権限モデル（principals / principal_members / workspace_grants /
-- space_grants / page_grants / share_links）の DDL。
--
-- knowledge_base.sql（骨格 6 テーブル）と同じ扱い: このファイルが実スキーマの正本であり、
-- 同時に sqlc の型付け入力でもある（backend/sqlc.yaml の schema に登録済み）。
-- 起動時に database.ApplyKnowledgeBaseSchema が骨格の DDL に続けてそのまま流す（冪等）。
-- CREATE ... IF NOT EXISTS だけで冪等性を成り立たせ、DO ブロックは書かない（sqlc がパースするため）。
--
-- 骨格と別ファイルにする理由は 2 つ:
--   (1) 骨格 6 テーブルは「ページの木そのもの」で、権限は「誰がそれを触れるか」という別の関心。
--       1 枚に混ぜると、どちらを読みたいときも全部を読むことになる。
--   (2) こちらは users（schema/core.sql が作るテーブル）へ FK を張る。骨格側は
--       ノートだけで閉じており、その依存の有無をファイル境界で見えるようにする。
--       適用順は database.Migrate が core → 骨格 → 権限なので users は必ず先にある。
--
-- 設計の柱（骨格の 2 つに加えて）:
--
--   (3) 主体（principal）を 1 つの表に集める。ユーザー・グループ・スペース全員・公開リンクは
--       「権限を与える相手」という点で同じなので、grant 側から見て 1 本の FK で済む。
--       主体ごとに表を分けると grant が主体の種類だけ列（または表）を持つことになり、
--       権限を解く SQL が主体の種類だけ分岐する。
--
--   (4) 種類（kind）によって使う列が変わるので、CHECK で「その kind のときだけ非 NULL」を強制する。
--       任意の key/value に逃がす（EAV）ことはしない。列は意味を持ったまま、
--       「いつ埋まるか」だけを制約で表す。
--
--   (5) 権限は付与（grants）だけで表し、打ち消す層は持たない。
--       入れ物の階層に合わせて 3 段（workspace_grants / space_grants / page_grants）を置き、
--       届いた中で最も強い役割を採る。下の段が上の段を弱めることはない。
--
--       全ページへ ACL を展開する方式は解決が 1 行の取得で済む代わりに、ページを 1 回動かす /
--       メンバーを 1 人足すだけで数万行を書き換える。ページ移動が日常の道具である以上、
--       書き込み側の代償が大きすぎる。付与はごく少数のページにしか付かない性質を使い、
--       行を持つのは付与された段だけにして、解決は page_paths（closure）を 1 回 JOIN するだけで済ませる。

-- principals: 権限を与える相手（主体）。
--
-- **この表がワークスペース所属の唯一の表現**（「そのワークスペースに kind='user' の行がある」
-- ＝ そのユーザーはメンバー）。workspace_memberships のようなメンバーシップ専用の表は
-- 作らない・足さないこと。作ると「principal はあるがメンバーではない」「メンバーだが
-- principal が無い」の 2 通りのずれが生まれ、どちらが正かを決められなくなる。
-- 所属の追加 / 削除はこの表への 1 行の INSERT / DELETE で表す。
--
-- 「未所属」は行が無いことで表す。専用の値（0 や空文字）は置かない。既存の users.company_id が
-- NULL と 0 の 2 通りで未所属を表していて層をまたいで混在しているのと同じ轍を踏まないため。
CREATE TABLE IF NOT EXISTS principals (
    id           uuid PRIMARY KEY,
    workspace_id uuid NOT NULL,
    -- kind の値は domain.PrincipalKind が正（user / group / space_all / share_link）。
    kind         varchar(16) NOT NULL,
    -- user_id は kind='user' のときだけ埋まる（既存 users への参照）。
    user_id      bigint,
    -- space_id は kind='space_all'（そのスペースの全員）のときだけ埋まる。
    space_id     uuid,
    -- page_id は kind='share_link'（公開リンクの来訪者）のときだけ埋まる。そのリンクの対象ページ。
    -- 主体を「それが意味を持つ入れ物」に必ず結び付けるためで、こうするとページを物理削除したときに
    -- 主体もリンクも CASCADE で一緒に消える。逆向き（share_links → principals）の FK だけでは、
    -- ページを消してもリンクの行だけが消えて主体が残り、誰も指さない行が溜まる。
    page_id      uuid,
    -- name は kind='group' の表示名。ほかの kind は名前を持たない
    -- （ユーザー名は users、スペース名は spaces が正本。ここへ写すと二重管理になる）。
    name         varchar(200) NOT NULL DEFAULT '',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_principals_workspace FOREIGN KEY (workspace_id)
        REFERENCES workspaces (id) ON DELETE CASCADE,
    -- users への FK は張る。principals はノートとアプリのユーザーを結ぶ唯一の接点で、
    -- ここが緩いと「消えたユーザーの principal に権限が残る」＝ 別人が同じ id を再取得したときに
    -- 権限を引き継いでしまう。骨格側の pages.created_by_user_id が FK を持たないのは、
    -- あちらが既存テーブル（IF NOT EXISTS では後から制約を足せない）だからで、
    -- 新しく作るこの表には最初から張れる。
    CONSTRAINT fk_principals_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE,
    -- スペース全員の主体は「同じワークスペースの space」にしか結び付かない。
    CONSTRAINT fk_principals_space FOREIGN KEY (workspace_id, space_id)
        REFERENCES spaces (workspace_id, id) ON DELETE CASCADE,
    -- 公開リンクの主体は「同じワークスペースの page」にしか結び付かない。
    CONSTRAINT fk_principals_page FOREIGN KEY (workspace_id, page_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    -- grant / share_link からの複合 FK の参照先。id の PK があるので実データ上は
    -- 冗長だが、「別ワークスペースの principal に権限を張れない」を FK で塞ぐ足場として要る。
    CONSTRAINT uq_principals_workspace_id UNIQUE (workspace_id, id),
    -- kind まで含めた足場。参照側が「この列は group の principal でなければならない」を
    -- FK で言えるようにする（principal_members / share_links が使う）。
    CONSTRAINT uq_principals_workspace_kind_id UNIQUE (workspace_id, kind, id),
    -- share_links からの複合 FK の参照先。リンクが持つ page_id と、その主体が持つ page_id が
    -- 必ず同じページを指すことを FK で言えるようにする（2 か所に同じ値を持つ以上、
    -- 食い違わないことは制約で担保する）。
    CONSTRAINT uq_principals_workspace_kind_page_id UNIQUE (workspace_id, kind, page_id, id),
    CONSTRAINT ck_principals_kind CHECK (kind IN ('user', 'group', 'space_all', 'share_link')),
    -- 使う列は kind で決まる。「その kind のときだけ非 NULL」を等式で書き、
    -- 片方向（NOT NULL なのに kind が違う）も同時に塞ぐ。
    CONSTRAINT ck_principals_user_id CHECK ((kind = 'user') = (user_id IS NOT NULL)),
    CONSTRAINT ck_principals_space_id CHECK ((kind = 'space_all') = (space_id IS NOT NULL)),
    CONSTRAINT ck_principals_page_id CHECK ((kind = 'share_link') = (page_id IS NOT NULL)),
    CONSTRAINT ck_principals_name CHECK ((kind = 'group') = (name <> ''))
);

-- principal_members: グループの所属（group principal ↔ member principal）。
--
-- グループの入れ子は許さない。member 側を kind='user' に固定することで、
-- 「あるユーザーの主体の集合」を再帰なしの 1 回の UNION で出せる
-- （入れ子を許すと権限解決に再帰 CTE が要るうえ、グループ同士の循環を防ぐ手当ても要る）。
--
-- kind の固定は生成列（GENERATED ALWAYS AS ... STORED）で行う。定数なので INSERT / UPDATE から
-- 値を渡せず、書き手が間違えようがない。CHECK 付きの普通の列にすると「書けるが必ず同じ値」に
-- なり、実質使われない列を 2 つ抱えることになる。ここは足場であって属性ではない。
--
-- これが無いと、たとえば member_principal_id に他人のユーザー主体ではなくグループ主体を
-- 入れることでグループを入れ子にでき、解決 SQL（1 段しか辿らない）が黙って権限を取りこぼす。
-- 「取りこぼす」＝ 見えるはずのページが見えないだけなので、権限の穴ではないが原因を追いにくい。
CREATE TABLE IF NOT EXISTS principal_members (
    workspace_id        uuid NOT NULL,
    group_principal_id  uuid NOT NULL,
    member_principal_id uuid NOT NULL,
    -- FK の足場（定数の生成列）。テーブルの属性ではない。
    group_kind          varchar(16) GENERATED ALWAYS AS ('group') STORED,
    member_kind         varchar(16) GENERATED ALWAYS AS ('user') STORED,
    created_at          timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT principal_members_pkey PRIMARY KEY (group_principal_id, member_principal_id),
    CONSTRAINT fk_principal_members_group FOREIGN KEY (workspace_id, group_kind, group_principal_id)
        REFERENCES principals (workspace_id, kind, id) ON DELETE CASCADE,
    CONSTRAINT fk_principal_members_member FOREIGN KEY (workspace_id, member_kind, member_principal_id)
        REFERENCES principals (workspace_id, kind, id) ON DELETE CASCADE
);

-- workspace_grants: ワークスペース全体での既定の役割。配下の全スペースに効く。
--
-- スペース単位の grant だけでは「テナント全体の管理者」を表すのにスペースの数だけ grant を
-- 張って回ることになり、スペースが増えるたびに漏れる。入れ物の階層が workspace ⊃ space である
-- 以上、既定も同じ 2 段で持つ。
--
-- scope_type / scope_id を持つ 1 枚の汎用 grants 表にまとめる案は採らない。scope_id が
-- workspaces と spaces の 2 つの表を指すことになり、FK で参照先の実在もテナントの一致も
-- 守れなくなる。このスキーマは一貫して「境界を FK で塞ぐ」ことを優先しており、
-- 表が 1 枚増える代わりに両方とも複合 FK で守れる形を選ぶ。
--
-- workspaces への直接の FK は張らない。principals への複合 FK が (workspace_id, principal_id) で
-- 実在する principal との一致を要求し、その principal 自身が workspaces へ FK を持つため、
-- ワークスペースの実在も削除時の CASCADE も推移的に効く。
CREATE TABLE IF NOT EXISTS workspace_grants (
    workspace_id uuid NOT NULL,
    principal_id uuid NOT NULL,
    -- "role" の値は domain.GrantRole が正（admin / editor / commenter / viewer）。
    "role"       varchar(16) NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT workspace_grants_pkey PRIMARY KEY (workspace_id, principal_id),
    CONSTRAINT fk_workspace_grants_principal FOREIGN KEY (workspace_id, principal_id)
        REFERENCES principals (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT ck_workspace_grants_role CHECK ("role" IN ('admin', 'editor', 'commenter', 'viewer'))
);

-- space_grants: そのスペースでの既定の役割。ページに例外が無いときは、これと workspace_grants の
-- うち強い方が実効権限になる（domain.GrantRole.Rank 参照。弱い方を採る規則にすると
-- スペースを 1 つ作って viewer を張るだけでワークスペース管理者を締め出せてしまう）。
--
-- 1 つの principal がひとつのスペースで持つ役割は 1 つなので、代理キーを置かず自然キーを PK にする
-- （代理キーを置くと「同じ principal に viewer と editor の 2 行」が作れてしまい、
-- どちらが正かをアプリで決める羽目になる）。
CREATE TABLE IF NOT EXISTS space_grants (
    workspace_id uuid NOT NULL,
    space_id     uuid NOT NULL,
    principal_id uuid NOT NULL,
    -- "role" の値は domain.GrantRole が正（admin / editor / commenter / viewer）。
    "role"       varchar(16) NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT space_grants_pkey PRIMARY KEY (workspace_id, space_id, principal_id),
    CONSTRAINT fk_space_grants_space FOREIGN KEY (workspace_id, space_id)
        REFERENCES spaces (workspace_id, id) ON DELETE CASCADE,
    -- workspace_id を含めることで「別ワークスペースの principal への grant」を DB が弾く。
    -- principal_id 単独の FK だと、行の workspace_id と principal の workspace_id が
    -- 食い違っていても両方の FK を通ってしまう（テナント越えの権限昇格になる）。
    CONSTRAINT fk_space_grants_principal FOREIGN KEY (workspace_id, principal_id)
        REFERENCES principals (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT ck_space_grants_role CHECK ("role" IN ('admin', 'editor', 'commenter', 'viewer'))
);

-- page_grants: そのページ以下での既定の役割。workspace_grants / space_grants に続く 3 段目で、
-- 意味も合成の仕方も上の 2 つと同じ（配下へ降りる・最も強いものを採る）。
--
-- これが要るのは「この人にこのページだけ編集を渡す」を書くため。
--
-- 経路は page_paths を辿る。祖先のページに editor を張れば、その子孫は既定が editor 以上に
-- なる（親に渡したら配下も編集できる、という素直な形）。
--
-- **弱める手段はこの層にも、どの層にも無い。** 権限は 3 段の付与を足し合わせて
-- 「届いた中で最も強いもの」で決まり、下の段が上の段を打ち消すことはない。
-- 「親は共有、この子だけ隠す」は書けない — 狭めたい内容は private のスペースへ置く。
CREATE TABLE IF NOT EXISTS page_grants (
    workspace_id uuid NOT NULL,
    page_id      uuid NOT NULL,
    principal_id uuid NOT NULL,
    -- "role" の値は domain.GrantRole が正（admin / editor / commenter / viewer）。
    "role"       varchar(16) NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT page_grants_pkey PRIMARY KEY (workspace_id, page_id, principal_id),
    CONSTRAINT fk_page_grants_page FOREIGN KEY (workspace_id, page_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    -- space_grants と同じ理由で workspace_id を含む複合 FK にする（別テナントの principal へ
    -- 付与できてしまうと、そのままテナント越えの権限昇格になる）。
    CONSTRAINT fk_page_grants_principal FOREIGN KEY (workspace_id, principal_id)
        REFERENCES principals (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT ck_page_grants_role CHECK ("role" IN ('admin', 'editor', 'commenter', 'viewer'))
);
-- 経路をさかのぼって「祖先に張られた付与」を引く向きの索引。主キーは (workspace_id, page_id,
-- principal_id) なので page_id 先頭では principal から引けない。
CREATE INDEX IF NOT EXISTS idx_page_grants_principal ON page_grants (workspace_id, principal_id);

-- share_links: ログイン不要の公開 URL。
--
-- 来訪者は kind='share_link' の principal として扱う。主体の種類を 1 本に揃えておくと、
-- 権限解決の入口が主体ごとに分岐しない。
--
-- ただし既定（そのリンクで何ができるか）は grants ではなくこの表の capability で決める。
-- リンクの来訪者はワークスペースに所属しないので、付与の 3 段はそもそも届かない。
--
-- **共有リンクは広げる方向にしか働かない。** ログインしていない相手へ「見せる」を足すだけで、
-- すでに見えている人から取り上げることはない。
--
-- token は平文で持たない。DB が漏れた時点で全リンクが開けるのを避けるため、SHA-256 の
-- ダイジェストだけを保存して照合はハッシュ同士で行う（トークンは十分な長さの乱数なので
-- 総当たりに強く、bcrypt のような遅いハッシュは要らない）。パスワードは人が選ぶ値なので
-- 逆に総当たりに弱く、こちらは bcrypt で持つ。
CREATE TABLE IF NOT EXISTS share_links (
    id                 uuid PRIMARY KEY,
    workspace_id       uuid NOT NULL,
    -- page_id はリンクの対象ページ。このページとその子孫が対象になる。
    page_id            uuid NOT NULL,
    principal_id       uuid NOT NULL,
    -- FK の足場（定数の生成列）。principal_members と同じ理由でここも生成列にする。
    principal_kind     varchar(16) GENERATED ALWAYS AS ('share_link') STORED,
    -- capability の値は domain.Capability が正（view / edit）。
    capability         varchar(8) NOT NULL,
    -- token_hash は共有 URL に載るトークンの SHA-256（32 バイト固定）。
    token_hash         bytea NOT NULL,
    -- password_hash は bcrypt。NULL ならパスワード無しで開ける。
    password_hash      text,
    -- expires_at が NULL なら無期限。
    expires_at         timestamptz,
    -- revoked_at が NULL なら有効。失効は行を消さず日付で残す（誰がいつ止めたかを追えるように）。
    revoked_at         timestamptz,
    created_by_user_id bigint NOT NULL,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_share_links_page FOREIGN KEY (workspace_id, page_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    -- principal は「同じワークスペースの、kind='share_link' の、同じページに結び付いた」主体だけ。
    -- page_id まで参照列に含めることで、リンクと主体が別々のページを指す状態を作れなくする。
    CONSTRAINT fk_share_links_principal FOREIGN KEY (workspace_id, principal_kind, page_id, principal_id)
        REFERENCES principals (workspace_id, kind, page_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_share_links_created_by FOREIGN KEY (created_by_user_id)
        REFERENCES users (id) ON DELETE CASCADE,
    -- トークンからリンクを 1 件引く経路。UNIQUE はその索引も兼ねる。
    CONSTRAINT uq_share_links_token_hash UNIQUE (token_hash),
    -- 1 つの share_link principal は 1 本のリンクだけを表す（使い回すと失効が効かなくなる）。
    CONSTRAINT uq_share_links_principal UNIQUE (principal_id),
    CONSTRAINT ck_share_links_capability CHECK (capability IN ('view', 'edit')),
    -- SHA-256 以外（平文トークンをそのまま入れた等）を入口で弾く。
    CONSTRAINT ck_share_links_token_hash_len CHECK (octet_length(token_hash) = 32),
    CONSTRAINT ck_share_links_password_hash CHECK (password_hash IS NULL OR password_hash <> '')
);

-- --- 一意性（部分 UNIQUE は WHERE 付きなのでテーブル定義に書けず、索引として張る）---
-- 1 ユーザー 1 ワークスペースにつき主体は 1 つ（重複メンバーを作らない）。
CREATE UNIQUE INDEX IF NOT EXISTS uq_principals_workspace_user
    ON principals (workspace_id, user_id) WHERE kind = 'user';
-- 1 スペースにつき「全員」の主体は 1 つ。
CREATE UNIQUE INDEX IF NOT EXISTS uq_principals_space_all
    ON principals (workspace_id, space_id) WHERE kind = 'space_all';
-- グループ名はワークスペース内で一意（同名グループが 2 つあると権限を張る先を人が選べない）。
CREATE UNIQUE INDEX IF NOT EXISTS uq_principals_group_name
    ON principals (workspace_id, name) WHERE kind = 'group';

-- --- 取得経路の索引（FK の子側は自動では索引が張られないため明示する）---
CREATE INDEX IF NOT EXISTS idx_principals_workspace_id ON principals (workspace_id);
CREATE INDEX IF NOT EXISTS idx_principals_user_id ON principals (user_id);
CREATE INDEX IF NOT EXISTS idx_principals_space_id ON principals (space_id);
CREATE INDEX IF NOT EXISTS idx_principals_page_id ON principals (page_id);
-- 「このユーザーが属するグループ」を引く経路（PK は group 側が先頭なので member 単独では効かない）。
CREATE INDEX IF NOT EXISTS idx_principal_members_member
    ON principal_members (workspace_id, member_principal_id);
-- 「この principal の grant」を引く経路（PK は入れ物側が先頭）。
-- principal を消すときの CASCADE 走査にも効く。
-- workspace_grants は PK が (workspace_id, principal_id) そのものなので追加の索引は要らない。
CREATE INDEX IF NOT EXISTS idx_space_grants_principal ON space_grants (workspace_id, principal_id);
CREATE INDEX IF NOT EXISTS idx_share_links_page ON share_links (workspace_id, page_id);
CREATE INDEX IF NOT EXISTS idx_share_links_created_by ON share_links (created_by_user_id);

-- =====================================================================
-- Ⅳ. テナント橋渡し（companies.workspace_id / users.workspace_id）と
--     個人ワークスペースの所有者（workspaces.personal_owner_user_id）
-- =====================================================================
--
-- companies / users は Ⅰ（中核）が作る表だが、workspaces を参照する列なので
-- workspaces（Ⅱ）より後に置く。CREATE TABLE ではなく ALTER TABLE ADD COLUMN
-- IF NOT EXISTS で足すのは、CREATE TABLE IF NOT EXISTS が既存の表へ列を追加しない
-- ため（本番に届く経路がこれしかない）。カタログを見て未作成のときだけ ALTER する
-- のは、素の ALTER TABLE が列が既に在ってスキップする場合でも先に
-- AccessExclusiveLock を取り、トランザクションが終わるまで手放さないため
-- （列が出揃っている通常の起動で companies / users / workspaces を掴まないようにする）。

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_schema = current_schema()
           AND table_name = 'companies' AND column_name = 'workspace_id'
    ) THEN
        ALTER TABLE companies ADD COLUMN workspace_id uuid;
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_schema = current_schema()
           AND table_name = 'users' AND column_name = 'workspace_id'
    ) THEN
        ALTER TABLE users ADD COLUMN workspace_id uuid;
    END IF;
END $$;

-- 会社とワークスペースは 1:1。移行中に 2 つの会社が同じワークスペースを指す状態を作らない。
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'uq_companies_workspace_id') THEN
        CREATE UNIQUE INDEX uq_companies_workspace_id ON companies (workspace_id) WHERE workspace_id IS NOT NULL;
    END IF;
END $$;

-- 存在しないワークスペースを指せないようにする（company_id には FK が無く、同じ轍を踏まない）。
-- 参照されている workspaces の行は消せない（既定の NO ACTION）。
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_companies_workspace' AND conrelid = 'companies'::regclass) THEN
        ALTER TABLE companies
            ADD CONSTRAINT fk_companies_workspace
            FOREIGN KEY (workspace_id) REFERENCES workspaces (id);
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_workspace' AND conrelid = 'users'::regclass) THEN
        ALTER TABLE users
            ADD CONSTRAINT fk_users_workspace
            FOREIGN KEY (workspace_id) REFERENCES workspaces (id);
    END IF;
END $$;

-- 個人サインアップで自動作成した、その人専用のワークスペースの持ち主。
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_schema = current_schema()
           AND table_name = 'workspaces' AND column_name = 'personal_owner_user_id'
    ) THEN
        ALTER TABLE workspaces ADD COLUMN personal_owner_user_id bigint;
    END IF;
END $$;

-- 作った人を物理削除しても中身は消さない。持ち主のいない箱として残り、招かれた他の
-- メンバーはそのまま使い続けられる。
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_workspaces_personal_owner' AND conrelid = 'workspaces'::regclass) THEN
        ALTER TABLE workspaces
            ADD CONSTRAINT fk_workspaces_personal_owner
            FOREIGN KEY (personal_owner_user_id) REFERENCES users (id) ON DELETE SET NULL;
    END IF;
END $$;

-- 1 人につき個人ワークスペースは 1 つ。サインアップの再送・並行実行でも 2 つ目が作れない
-- （check-then-act をアプリに書かずに済む。ON CONFLICT の推論先にもなる）。
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'uq_workspaces_personal_owner') THEN
        CREATE UNIQUE INDEX uq_workspaces_personal_owner
            ON workspaces (personal_owner_user_id) WHERE personal_owner_user_id IS NOT NULL;
    END IF;
END $$;

-- =====================================================================
-- Ⅴ. テナント統合（courses / course_chapters / company_exercises /
--     invitations / rich_documents への workspace_id 列追加）
-- =====================================================================
--
-- workspace_id が唯一の所属参照（company_id は全表から撤去済み）。列を足して FK を張る。
-- FK は workspace_id 側にだけ張る（companies で既に採った方針と同じ）。
-- 新規に作る DB では上の CREATE TABLE で最初から workspace_id を持つため、この節は
-- 既存 DB（起動時点でまだ列が無い環境）へ届かせるための ALTER TABLE ADD COLUMN
-- IF NOT EXISTS 経路。

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_schema = current_schema()
           AND table_name = 'courses' AND column_name = 'workspace_id'
    ) THEN
        ALTER TABLE courses ADD COLUMN workspace_id uuid;
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_courses_workspace' AND conrelid = 'courses'::regclass) THEN
        ALTER TABLE courses
            ADD CONSTRAINT fk_courses_workspace
            FOREIGN KEY (workspace_id) REFERENCES workspaces (id);
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_schema = current_schema()
           AND table_name = 'course_chapters' AND column_name = 'workspace_id'
    ) THEN
        ALTER TABLE course_chapters ADD COLUMN workspace_id uuid;
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_course_chapters_workspace' AND conrelid = 'course_chapters'::regclass) THEN
        ALTER TABLE course_chapters
            ADD CONSTRAINT fk_course_chapters_workspace
            FOREIGN KEY (workspace_id) REFERENCES workspaces (id);
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_schema = current_schema()
           AND table_name = 'company_exercises' AND column_name = 'workspace_id'
    ) THEN
        ALTER TABLE company_exercises ADD COLUMN workspace_id uuid;
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_company_exercises_workspace' AND conrelid = 'company_exercises'::regclass) THEN
        ALTER TABLE company_exercises
            ADD CONSTRAINT fk_company_exercises_workspace
            FOREIGN KEY (workspace_id) REFERENCES workspaces (id);
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_schema = current_schema()
           AND table_name = 'invitations' AND column_name = 'workspace_id'
    ) THEN
        ALTER TABLE invitations ADD COLUMN workspace_id uuid;
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_invitations_workspace' AND conrelid = 'invitations'::regclass) THEN
        ALTER TABLE invitations
            ADD CONSTRAINT fk_invitations_workspace
            FOREIGN KEY (workspace_id) REFERENCES workspaces (id);
    END IF;
END $$;

-- rich_documents は他の 4 テーブルと違い未所属のドキュメントがあるため、
-- workspace_id は NULL を許容する。列追加・FK の作法自体は他と同じ。
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_schema = current_schema()
           AND table_name = 'rich_documents' AND column_name = 'workspace_id'
    ) THEN
        ALTER TABLE rich_documents ADD COLUMN workspace_id uuid;
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_rich_documents_workspace' AND conrelid = 'rich_documents'::regclass) THEN
        ALTER TABLE rich_documents
            ADD CONSTRAINT fk_rich_documents_workspace
            FOREIGN KEY (workspace_id) REFERENCES workspaces (id);
    END IF;
END $$;

-- workspace_id の索引は列を足したあとに作る。節Ⅰ（CREATE TABLE 群）へ置くと、既存 DB では
-- まだ列が無い時点で CREATE INDEX が走って落ちる（IF NOT EXISTS は索引の有無しか見ず、
-- 列の不在は防げない）。
CREATE INDEX IF NOT EXISTS idx_courses_workspace_id ON courses (workspace_id);
CREATE INDEX IF NOT EXISTS idx_course_chapters_workspace_id ON course_chapters (workspace_id);
CREATE INDEX IF NOT EXISTS idx_company_exercises_workspace_id ON company_exercises (workspace_id);
CREATE INDEX IF NOT EXISTS idx_invitations_workspace_id ON invitations (workspace_id);

-- ── コースと章の骨格を締める（対象ごとに権限を張るための足場）──
--
-- コース・教材の編集可否を「対象ごと」に決めるには、権限の行から対象を
-- **テナントごと**指せる必要がある。ノート側（page_grants）と同じ形にするので、
-- 参照される側に (workspace_id, id) の一意制約が要る。
--
-- id は既に主キーなので、この一意制約が新しく防ぐ重複は無い。**要るのは複合外部キーの
-- 参照先として**で、これが無いと権限の行から「同じワークスペースのコース」を指せず、
-- テナントを跨いだ付与を DB で塞げない（アプリの検査だけが頼りになる）。

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conname = 'uq_courses_workspace_id' AND conrelid = 'courses'::regclass
    ) THEN
        ALTER TABLE courses ADD CONSTRAINT uq_courses_workspace_id UNIQUE (workspace_id, id);
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conname = 'uq_course_chapters_workspace_id' AND conrelid = 'course_chapters'::regclass
    ) THEN
        ALTER TABLE course_chapters
            ADD CONSTRAINT uq_course_chapters_workspace_id UNIQUE (workspace_id, id);
    END IF;
END $$;

-- 章は必ず実在するコースにぶら下がる。
--
-- ORM が作っていた頃からこの FK は無く、コースを消しても章を残せる状態だった
-- （消す側のコードが明示的に消していただけで、DB は何も守っていない）。対象ごとの権限を
-- 張ると、親の居ない章に権限だけが残り、誰の目にも触れないまま生き続ける。
--
-- **複合（workspace_id を含む）にはしない。** workspace_id はまだ NULL を許すので、
-- 複合にすると NULL の行では検査そのものが飛ぶ（MATCH SIMPLE の既定）。course_id は
-- NOT NULL なので、単純な FK なら全行で必ず効く。テナントの一致は上の一意制約と、
-- 権限側から張る複合 FK が受け持つ。
--
-- workspace_id を NOT NULL にしていないのは、教材の投入 SQL（別リポで生成する）が
-- この列を書いているかをここから確かめられないため。確かめてから別途締める。
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conname = 'fk_course_chapters_course' AND conrelid = 'course_chapters'::regclass
    ) THEN
        ALTER TABLE course_chapters
            ADD CONSTRAINT fk_course_chapters_course
            FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE;
    END IF;
END $$;

-- ── コース / 教材の権限（対象ごとの編集）──
--
-- 既定は「コース → 章」の 2 段で届き、最も強いものが実効になる。ノートと同じ合成規則だが、
-- **ワークスペースの grant は届かない**。あれはノートの木に対する既定で、教材は別の入れ物。
-- 届かせると、いまノートの editor である人が教材の編集権まで一度に得てしまう
-- （実際に本番でその状態の人が居る）。
--
-- 唯一の例外がワークスペースの admin で、配下すべてを管理できる。そうしないと、
-- 付与された人が居なくなった教材の権限を誰も変えられなくなる（ノート側で最後の admin を
-- 守っているのと同じ理由）。この 1 つだけが例外であることは domain 側の解決に書いてある。
--
-- 読むことには付与を要求しない。公開済みの教材はワークスペースの一員なら誰でも読める
-- （学ぶための場なので、読む側に権限を持たせない）。下書きだけが編集できる人に限られる。

CREATE TABLE IF NOT EXISTS course_grants (
    workspace_id uuid NOT NULL,
    course_id    bigint NOT NULL,
    principal_id uuid NOT NULL,
    "role"       varchar(16) NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT course_grants_pkey PRIMARY KEY (workspace_id, course_id, principal_id),
    -- 複合にすることで「同じワークスペースのコース」しか指せない（テナント跨ぎの付与を塞ぐ）。
    CONSTRAINT fk_course_grants_course FOREIGN KEY (workspace_id, course_id)
        REFERENCES courses (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_course_grants_principal FOREIGN KEY (workspace_id, principal_id)
        REFERENCES principals (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT ck_course_grants_role CHECK ("role" IN ('admin', 'editor', 'commenter', 'viewer'))
);
CREATE INDEX IF NOT EXISTS idx_course_grants_principal ON course_grants (workspace_id, principal_id);

-- 章 1 つだけに効く既定の権限（「この教材だけ編集してよい」）。
-- コースの付与より弱い役割をここに張っても下がらない（合成は最も強いものを採る）。
CREATE TABLE IF NOT EXISTS chapter_grants (
    workspace_id uuid NOT NULL,
    chapter_id   bigint NOT NULL,
    principal_id uuid NOT NULL,
    "role"       varchar(16) NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT chapter_grants_pkey PRIMARY KEY (workspace_id, chapter_id, principal_id),
    CONSTRAINT fk_chapter_grants_chapter FOREIGN KEY (workspace_id, chapter_id)
        REFERENCES course_chapters (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_chapter_grants_principal FOREIGN KEY (workspace_id, principal_id)
        REFERENCES principals (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT ck_chapter_grants_role CHECK ("role" IN ('admin', 'editor', 'commenter', 'viewer'))
);
CREATE INDEX IF NOT EXISTS idx_chapter_grants_principal ON chapter_grants (workspace_id, principal_id);
