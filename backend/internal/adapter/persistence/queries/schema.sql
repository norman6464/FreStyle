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
--   2. ナレッジ基盤（workspaces / spaces / pages / blocks / page_paths / page_snapshots）
--      GORM を通さない。infra/database/schema/knowledge_base.sql が実スキーマの正本そのもので、
--      起動時に ApplyKnowledgeBaseSchema が埋め込み DDL をそのまま流す。sqlc へも同じファイルを
--      入力として渡している（sqlc.yaml の schema 欄）ため定義が二重化しない。
--      ここへ書き写さないこと（写すと 1. と同じ二重管理に戻る）。

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
    deleted_at   timestamptz
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
    updated_at                   timestamptz NOT NULL
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
