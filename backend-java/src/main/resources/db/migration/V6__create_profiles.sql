-- profiles: users と 1:1 のプロフィール拡張(user_id が PK = users.id)。Profile エンティティに対応。
-- 型は H2(PostgreSQL モード) と PostgreSQL の双方で通る書き方。本番は既存(GORM 製)のため no-op。
CREATE TABLE IF NOT EXISTS profiles (
    user_id    BIGINT PRIMARY KEY,
    bio        VARCHAR,
    avatar_url VARCHAR,
    status     VARCHAR DEFAULT '',
    updated_at TIMESTAMP
);
