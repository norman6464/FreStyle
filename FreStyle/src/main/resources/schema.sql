-- users
CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(254) UNIQUE NOT NULL,
    icon_url VARCHAR(255),
    bio TEXT,
    is_active BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- user_identities
CREATE TABLE IF NOT EXISTS user_identities (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    provider VARCHAR(255) NOT NULL,
    provider_sub VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uq_provider_sub (provider, provider_sub),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- chat_rooms
CREATE TABLE IF NOT EXISTS chat_rooms (
    id INT PRIMARY KEY AUTO_INCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- room_members
CREATE TABLE IF NOT EXISTS room_members (
    id INT PRIMARY KEY AUTO_INCREMENT,
    room_id INT NOT NULL, 
    user_id INT NOT NULL,

    UNIQUE KEY uk_room_member (room_id, user_id),

    FOREIGN KEY (room_id) REFERENCES chat_rooms(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- chat_messages
CREATE TABLE IF NOT EXISTS chat_messages (
    id INT PRIMARY KEY AUTO_INCREMENT,
    room_id INT NOT NULL,
    sender_id INT NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (room_id) REFERENCES chat_rooms(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- unread_counts
CREATE TABLE IF NOT EXISTS unread_counts (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    room_id INT NOT NULL,
    count INT NOT NULL DEFAULT 0,

    UNIQUE KEY uk_unread_count (user_id, room_id),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (room_id) REFERENCES chat_rooms(id) ON DELETE CASCADE
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;




-- ユーザーが参加しているルームを効率的に検索するためのインデックス
-- CREATE INDEX idx_room_member_user_id ON room_members (user_id);

-- 実行する前に必ず見る
-- DROP TABLE IF EXISTS 
--     unread_counts,
--     room_members,
--     chat_rooms,
--     use
--     users;

-- 文字コードの設定をする
-- Database changed
-- MariaDB [fre_style]> ALTER TABLE users CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
-- Query OK, 0 rows affected (0.06 sec)
-- Records: 0  Duplicates: 0  Warnings: 0