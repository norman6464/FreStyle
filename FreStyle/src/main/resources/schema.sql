CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    cognito_sub VARCHAR(36) UNIQUE,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(254) UNIQUE NOT NULL,
    icon_url VARCHAR(255), -- S3などに対してアイコンを保存できるようにする --
    bio TEXT,
    is_active BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP ,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS  chat_rooms (
    id INT PRIMARY KEY AUTO_INCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS room_members (
    id INT PRIMARY KEY AUTO_INCREMENT,
    room_id INT NOT NULL, 
    user_id INT NOT NULL,
    
    -- 複合ユニーク制約: 1つのルームに同じユーザーは二重参加できない
    UNIQUE KEY uk_room_member (room_id, user_id),
    
    -- 外部キー制約の定義
    FOREIGN KEY (room_id) REFERENCES chat_rooms(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- ユーザーが参加しているルームを効率的に検索するためのインデックス
-- CREATE INDEX idx_room_member_user_id ON room_members (user_id);

CREATE TABLE IF NOT EXISTS unread_counts (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    room_id INT NOT NULL,
    -- 未読メッセージ数 (今回の要件で必須となるカラム)
    count INT NOT NULL DEFAULT 0,
    
    -- 複合ユニーク制約: ユーザーとルームの組み合わせは一意
    UNIQUE KEY uk_unread_count (user_id, room_id),
    
    -- 外部キー制約の定義
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (room_id) REFERENCES chat_rooms(id) ON DELETE CASCADE
);

-- 実行する前に必ず見る
-- DROP TABLE IF EXISTS 
--     unread_counts,
--     room_members,
--     chat_rooms,
--     users;
