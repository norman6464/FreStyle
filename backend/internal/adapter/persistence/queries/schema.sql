-- sqlc の型付け専用スキーマ。ここに書いた CREATE TABLE から Go の型を起こすだけで、
-- このファイル自体が DB へ流れることは無い（実スキーマはここでは作られない）。
-- repository を sqlc へ移行するたびに、対象テーブルの CREATE TABLE をここへ追記していく。
--
-- 実スキーマの正本がどこにあるかはテーブルによって違う。取り違えると片側だけ直して本番とずれる。
--
--   1. このファイルに並んでいるテーブル（users / roles / notes / courses など）
--      正本は domain 構造体の GORM タグ。infra/database/migrate.go の allDomainModels() に
--      並べた構造体を、起動時に AutoMigrate が適用して実スキーマを作る。AutoMigrate が
--      表現できない FK / CHECK / 部分 UNIQUE は同ファイルの ApplyXxxConstraints が明示 SQL で足す。
--      つまりここは正本ではなく写しなので、列を足す・型を変えるときは domain 構造体
--      （必要なら ApplyXxxConstraints）を先に直し、このファイルを追随させること。
--
--   2. ナレッジ基盤（骨格: workspaces / spaces / pages / blocks / page_paths / page_snapshots、
--      権限モデル: principals / principal_members / workspace_grants / space_grants /
--      page_restrictions / page_allow_lists / share_links）
--      GORM を通さない。実スキーマの正本は次の 2 ファイルそのもので、起動時に
--      ApplyKnowledgeBaseSchema（infra/database/knowledge_base_schema.go）が
--      埋め込み DDL を 1 トランザクションでこの順に流す。
--
--        internal/infra/database/schema/knowledge_base.sql              （骨格）
--        internal/infra/database/schema/knowledge_base_permissions.sql  （権限モデル）
--
--      権限側は骨格の spaces / pages と AutoMigrate が作る users を参照するので、
--      適用順（AutoMigrate → 骨格 → 権限）は崩せない。
--      sqlc へはこの 2 ファイルをそのまま入力として渡している（sqlc.yaml の schema 欄に
--      3 ファイルとも並んでいる）ため、定義が二重化しない。
--      ここへ書き写さないこと（写すと 1. と同じ二重管理に戻る）。
--
--   3. テナント統合の橋渡し列（companies.workspace_id / users.workspace_id）
--      1. のテーブルに生えているが AutoMigrate は作らない。domain 構造体に持たせず、
--      infra/database/tenant_bridge.go の tenantBridgeSchemaStatements が明示 DDL で足す
--      （FK と部分 UNIQUE は GORM タグで表現できず、読み取りも変えたくないため）。
--      ALTER TABLE ADD COLUMN は必ず列を末尾に付けるので、このファイルでも該当テーブルの
--      最後に書くこと（並びがずれると SELECT * の詰め替えが位置ずれで壊れる）。
--      移行期間だけの列で、companies を畳むときに companies 側は消える。

-- created_at / updated_at は GORM の autoCreateTime / autoUpdateTime が常に値を入れるため
-- NOT NULL とみなす（sqlc が sql.NullTime ではなく time.Time を生成し、domain への詰め替えが綺麗になる）。
CREATE TABLE master_exercise_examples (
    id              bigint PRIMARY KEY,
    exercise_id     bigint NOT NULL,
    order_index     smallint NOT NULL,
    input_text      text NOT NULL DEFAULT '',
    expected_output text NOT NULL,
    created_at      timestamptz NOT NULL,
    updated_at      timestamptz NOT NULL
);

-- ロールマスタ（FRESTYLE-311）。name はアプリのビジネス定数（super_admin 等）と一致する。
CREATE TABLE roles (
    id          smallint PRIMARY KEY,
    name        text NOT NULL,
    description text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL,
    updated_at  timestamptz NOT NULL
);

-- email / name はアプリが必ず値を入れるため NOT NULL とみなす（sqlc が string を生成し
-- domain への詰め替えが綺麗になる）。company_id / deleted_at は実際に NULL になり得る
-- （SuperAdmin は company 無し等）ので nullable のまま。role_id は正規化後の正で NOT NULL。
-- OIDC プロバイダ由来のユーザー識別子（FRESTYLE-311）。Cognito の sub を users から分離。
CREATE TABLE user_oidc_identities (
    id         bigint PRIMARY KEY,
    user_id    bigint NOT NULL,
    provider   text NOT NULL DEFAULT 'cognito',
    subject    text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    -- 実制約は AutoMigrate（gorm uniqueIndex タグ）が作る。ここは sqlc の型付け用に同じ内容を明示。
    UNIQUE (user_id, provider),
    UNIQUE (provider, subject)
);

CREATE TABLE users (
    id           bigint PRIMARY KEY,
    email        text NOT NULL DEFAULT '',
    password_hash text,
    name       text NOT NULL DEFAULT '',
    company_id   bigint,
    role_id      smallint NOT NULL,
    ai_chat_enabled boolean,
    is_active    boolean NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL,
    updated_at   timestamptz NOT NULL,
    deleted_at   timestamptz,
    -- workspace_id は AutoMigrate ではなく明示 DDL（infra/database の tenant bridge）が
    -- 末尾に足す列。実列の並びと合わせるため、ここでも必ず最後に書く（SELECT * の詰め替えが
    -- 位置ずれで壊れないように）。所属の正本は当面 company_id のままで、この列は写し。
    workspace_id uuid
);

-- 学習メモ。全列アプリが必ず値を入れる（user_id / title / content / is_public / is_pinned）ため
-- NOT NULL とみなす。domain.Note も全フィールド非ポインタなので 1:1 に詰め替えられる。
CREATE TABLE notes (
    id         bigint PRIMARY KEY,
    user_id    bigint NOT NULL,
    title      text NOT NULL DEFAULT '',
    content    text NOT NULL DEFAULT '',
    is_public  boolean NOT NULL DEFAULT false,
    is_pinned  boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- users とは別管理のプロフィール拡張（user_id が PK）。全列 domain.Profile と 1:1。
CREATE TABLE profiles (
    user_id    bigint PRIMARY KEY,
    bio            text NOT NULL DEFAULT '',
    avatar_url     text NOT NULL DEFAULT '',
    status_message text NOT NULL DEFAULT '',
    updated_at     timestamptz NOT NULL
);

-- アプリ内通知。全列 domain.Notification と 1:1。
CREATE TABLE notifications (
    id         bigint PRIMARY KEY,
    user_id    bigint NOT NULL,
    type       text NOT NULL DEFAULT '',
    title      text NOT NULL DEFAULT '',
    body       text NOT NULL DEFAULT '',
    is_read    boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL
);

-- AI チャットセッション固有のメモ。全列 domain.SessionNote と 1:1。
CREATE TABLE session_notes (
    id         bigint PRIMARY KEY,
    session_id bigint NOT NULL,
    user_id    bigint NOT NULL,
    content    text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- 企業。ai_chat_enabled_for_trainees は AutoMigrate が追加する列（domain と 1:1）。
CREATE TABLE companies (
    id                           bigint PRIMARY KEY,
    name                         text NOT NULL,
    ai_chat_enabled_for_trainees boolean NOT NULL DEFAULT true,
    is_active                    boolean NOT NULL DEFAULT true,
    created_at                   timestamptz NOT NULL,
    updated_at                   timestamptz NOT NULL,
    -- 対応する workspaces 行への橋渡し。テナントの正本を workspaces へ移す移行期間だけの列で、
    -- companies そのものを畳むときに列ごと消える（恒久的な 1:1 の関連として設計していない）。
    -- 実列は明示 DDL が末尾に足すので、ここでも必ず最後に書く（SELECT * の位置ずれ防止）。
    workspace_id                 uuid
);

-- 公開フォームからの利用申請。message は空文字許容（NULL は来ない想定で NOT NULL とみなす）。
CREATE TABLE company_applications (
    id             bigint PRIMARY KEY,
    company_name   text NOT NULL,
    applicant_name text NOT NULL,
    email          text NOT NULL,
    message        text NOT NULL DEFAULT '',
    status         text NOT NULL DEFAULT 'pending',
    created_at     timestamptz NOT NULL,
    updated_at     timestamptz NOT NULL
);

-- 教材コース。domain.Course の全フィールドが非ポインタで必ず値を持つため NOT NULL とみなす
-- 列定義は internal/domain/course.go を正とする。
CREATE TABLE courses (
    id                 bigint PRIMARY KEY,
    company_id         bigint NOT NULL,
    created_by_user_id bigint NOT NULL,
    title              text NOT NULL DEFAULT '',
    description        text NOT NULL DEFAULT '',
    category           text NOT NULL DEFAULT '',
    language           varchar(50) NOT NULL DEFAULT '',
    sort_order         integer NOT NULL DEFAULT 100,
    is_published       boolean NOT NULL DEFAULT false,
    created_at         timestamptz NOT NULL,
    updated_at         timestamptz NOT NULL
);

CREATE TABLE audit_events (
    id          bigint PRIMARY KEY,
    actor_id    bigint NOT NULL,
    actor_email varchar(255) NOT NULL DEFAULT '',
    actor_role  varchar(32) NOT NULL DEFAULT '',
    action      varchar(160) NOT NULL DEFAULT '',
    target_id   bigint NOT NULL,
    created_at  timestamptz NOT NULL
);

-- ---------------------------------------------------------------------
-- ここから下は AutoMigrate 管理なのに未定義だったテーブル群。上と同じく
-- domain 構造体の GORM タグを正本として写す（型付け専用・DB へは流れない）。
-- 型の対応: uint64→bigint / uint16→smallint / plain int→bigint /
--   type:integer→integer / int16(type:smallint)→smallint / string(size:N)→varchar(N) /
--   string(type:text)→text / bare string→text / type:jsonb→jsonb / type:uuid→uuid /
--   *T→nullable。created_at / updated_at は autoTime が必ず値を入れるので NOT NULL とみなし
--   DB DEFAULT は付けない（既存テーブルと同じ方針）。
-- ---------------------------------------------------------------------

-- AI チャットの 1 セッション。domain.AiChatSession の GORM タグを正とする。
-- title / session_type はアプリが必ず値を入れるため NOT NULL とみなす。scenario_id は *uint64 で nullable。
CREATE TABLE ai_chat_sessions (
    id           bigint PRIMARY KEY,
    user_id      bigint NOT NULL,
    title        text NOT NULL DEFAULT '',
    session_type text NOT NULL DEFAULT '',
    scenario_id  bigint,
    created_at   timestamptz NOT NULL,
    updated_at   timestamptz NOT NULL
);

-- 招待（マジックリンク）。domain.AdminInvitation（TableName = invitations）の GORM タグを正とする。
-- token は *string（未設定を NULL にして UNIQUE を避けるため）なので nullable。updated_at は持たない。
CREATE TABLE invitations (
    id         bigint PRIMARY KEY,
    company_id bigint NOT NULL,
    email      text NOT NULL DEFAULT '',
    role       text NOT NULL DEFAULT '',
    name       text NOT NULL DEFAULT '',
    status     text NOT NULL DEFAULT '',
    token      varchar(64),
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    -- 実制約は AutoMigrate（gorm uniqueIndex タグ）が作る。ここは sqlc の型付け用に同じ内容を明示。
    UNIQUE (token)
);

-- 運営が用意した練習問題マスタ。domain.MasterExercise の GORM タグを正とする。
-- sort_order は plain int（type:integer 指定なし）なので AutoMigrate 既定の bigint。
-- hint_text / expected_output は not null 指定が無く nullable。chapter_id は *uint64 で nullable。
CREATE TABLE master_exercises (
    id              bigint PRIMARY KEY,
    slug            varchar(64) NOT NULL,
    language        varchar(32) NOT NULL,
    sort_order      bigint NOT NULL DEFAULT 0,
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
    updated_at      timestamptz NOT NULL,
    -- 実制約は AutoMigrate（gorm uniqueIndex タグ）が作る。ここは sqlc の型付け用に同じ内容を明示。
    UNIQUE (slug)
);

-- 会社独自の演習。domain.CompanyExercise の GORM タグを正とする。
-- hint_text / expected_output は nullable。deleted_at は *time.Time で nullable。
CREATE TABLE company_exercises (
    id              bigint PRIMARY KEY,
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

-- コード演習の提出履歴（append-only）。domain.ExerciseSubmission の GORM タグを正とする。
-- stdout / stderr は nullable。exit_code は plain int なので bigint。
CREATE TABLE exercise_submissions (
    id             bigint PRIMARY KEY,
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

-- コースを構成する章。domain.TeachingMaterial（TableName = course_chapters）の GORM タグを正とする。
-- doc は *string type:jsonb（未移行の章は NULL）なので nullable jsonb。旧 content 列は撤去済で持たない。
-- revision / schema_version / sort_order は plain int なので bigint。
CREATE TABLE course_chapters (
    id                 bigint PRIMARY KEY,
    company_id         bigint NOT NULL,
    course_id          bigint NOT NULL,
    created_by_user_id bigint NOT NULL,
    title              text NOT NULL DEFAULT '',
    doc                jsonb,
    revision           bigint NOT NULL DEFAULT 1,
    schema_version     bigint NOT NULL DEFAULT 1,
    sort_order         bigint NOT NULL DEFAULT 100,
    is_published       boolean NOT NULL DEFAULT false,
    created_at         timestamptz NOT NULL,
    updated_at         timestamptz NOT NULL
);

-- 学習レポート（週次・月次集計を非同期生成）。domain.LearningReport の GORM タグを正とする。
-- updated_at は持たない。
CREATE TABLE learning_reports (
    id          bigint PRIMARY KEY,
    user_id     bigint NOT NULL,
    period_from timestamptz NOT NULL,
    period_to   timestamptz NOT NULL,
    status      text NOT NULL DEFAULT '',
    s3_key      text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL
);

-- 章の完了記録。domain.UserLessonProgress（TableName = user_chapter_progress）の GORM タグを正とする。
-- 列名は改名後（chapter_id）。(user_id, chapter_id) は複合 UNIQUE。
CREATE TABLE user_chapter_progress (
    id           bigint PRIMARY KEY,
    user_id      bigint NOT NULL,
    chapter_id   bigint NOT NULL,
    course_id    bigint NOT NULL,
    completed_at timestamptz NOT NULL,
    created_at   timestamptz NOT NULL,
    -- 実制約は AutoMigrate（gorm uniqueIndex タグ）が作る。ここは sqlc の型付け用に同じ内容を明示。
    UNIQUE (user_id, chapter_id)
);

-- 章の閲覧記録。domain.UserChapterView の GORM タグを正とする。PK = (user_id, chapter_id)。
-- 列名は改名後（chapter_id）。view_count は type:integer 指定のため integer（bigint にしない）。
CREATE TABLE user_chapter_views (
    user_id         bigint NOT NULL,
    chapter_id      bigint NOT NULL,
    course_id       bigint NOT NULL,
    first_viewed_at timestamptz NOT NULL,
    last_viewed_at  timestamptz NOT NULL,
    view_count      integer NOT NULL DEFAULT 1,
    PRIMARY KEY (user_id, chapter_id)
);

-- リッチテキスト文書（tiptap JSON を jsonb で保持）。domain.RichDocument の GORM タグを正とし、
-- FK / CHECK は ApplyRichDocumentConstraints（infra/database/migrate.go）を写す。
--   - id は uuid（アプリ採番。bigint ではない）
--   - doc は jsonb NOT NULL、object かつ type='doc' に限る CHECK
--   - owner_id → users(id) ON DELETE CASCADE
--   - company_id は *uint64 で nullable、deleted_at も nullable
CREATE TABLE rich_documents (
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

-- 日次の学習活動サマリー。domain.UserDailyActivity の GORM タグを正とする。
-- activity_date は date 型（PK の一部）。各 *_count は type:integer 指定のため integer。
-- chapter_count は改名後の列名（旧 lesson_count）。
CREATE TABLE user_daily_activities (
    user_id        bigint NOT NULL,
    activity_date  date NOT NULL,
    exercise_count integer NOT NULL DEFAULT 0,
    correct_count  integer NOT NULL DEFAULT 0,
    chapter_count  integer NOT NULL DEFAULT 0,
    ai_chat_count  integer NOT NULL DEFAULT 0,
    note_count     integer NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, activity_date)
);
