-- sqlc の型付け専用スキーマ。実スキーマは GORM AutoMigrate が作る（ここは生成のための定義）。
-- repository を sqlc へ移行するたびに、対象テーブルの CREATE TABLE をここへ追記していく。
-- 列定義は docs/schema.sql（本番の実体）と一致させること。

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

-- cognito_sub / email / display_name / role はアプリが必ず値を入れるため NOT NULL とみなす
-- （sqlc が string を生成し domain への詰め替えが綺麗になる）。company_id / onboarded_at /
-- deleted_at は実際に NULL になり得る（SuperAdmin は company 無し等）ので nullable のまま。
CREATE TABLE users (
    id           bigint PRIMARY KEY,
    cognito_sub  text NOT NULL,
    email        text NOT NULL DEFAULT '',
    display_name text NOT NULL DEFAULT '',
    company_id   bigint,
    role         text NOT NULL,
    ai_chat_enabled boolean,
    is_active    boolean NOT NULL DEFAULT true,
    onboarded_at timestamptz,
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
    bio        text NOT NULL DEFAULT '',
    avatar_url text NOT NULL DEFAULT '',
    status     text NOT NULL DEFAULT '',
    updated_at timestamptz NOT NULL
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
