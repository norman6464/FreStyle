-- user_daily_activities: ユーザーの1日分の学習活動サマリー
-- PK = (user_id, activity_date)。upsert-on-write で各カウンタを加算する。
CREATE TABLE IF NOT EXISTS user_daily_activities (
    user_id        BIGINT  NOT NULL,
    activity_date  DATE    NOT NULL,
    exercise_count INTEGER NOT NULL DEFAULT 0,
    correct_count  INTEGER NOT NULL DEFAULT 0,
    lesson_count   INTEGER NOT NULL DEFAULT 0,
    ai_chat_count  INTEGER NOT NULL DEFAULT 0,
    note_count     INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, activity_date)
);

CREATE INDEX IF NOT EXISTS idx_user_daily_activities_user_date
    ON user_daily_activities (user_id, activity_date DESC);

-- user_chapter_views: ユーザーが章（教材）を開いた記録
-- PK = (user_id, teaching_material_id)。upsert により last_viewed_at と view_count を更新する。
CREATE TABLE IF NOT EXISTS user_chapter_views (
    user_id              BIGINT      NOT NULL,
    teaching_material_id BIGINT      NOT NULL,
    course_id            BIGINT      NOT NULL,
    first_viewed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_viewed_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    view_count           INTEGER     NOT NULL DEFAULT 1,
    PRIMARY KEY (user_id, teaching_material_id)
);

CREATE INDEX IF NOT EXISTS idx_user_chapter_views_user_last_viewed
    ON user_chapter_views (user_id, last_viewed_at DESC);
