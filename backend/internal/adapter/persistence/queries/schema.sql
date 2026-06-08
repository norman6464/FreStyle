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
