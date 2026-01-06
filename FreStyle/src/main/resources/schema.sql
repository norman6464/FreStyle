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

-- user_profiles（ユーザーの自己設定・パーソナリティ情報）
CREATE TABLE IF NOT EXISTS user_profiles (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL UNIQUE,
    
    -- 基本的な自己紹介
    display_name VARCHAR(100),           -- 呼ばれたい名前
    self_introduction TEXT,              -- 自由形式の自己紹介
    
    -- コミュニケーションスタイル
    communication_style VARCHAR(50),     -- 例: 'casual', 'formal', 'friendly' など
    personality_traits JSON,             -- 性格特性（複数選択可能）例: ["内向的", "論理的", "共感力が高い"]
    
    -- AIフィードバック用の追加情報
    goals TEXT,                          -- コミュニケーションで改善したい点・目標
    concerns TEXT,                       -- 苦手なこと・気になっていること
    preferred_feedback_style VARCHAR(50), -- フィードバックの受け取り方 例: 'direct', 'gentle', 'detailed'
    
    -- メタ情報
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
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

-- アクセストークンを格納するテーブル（まだ検証段階なのでプロダクション環境に反映しない）
CREATE TABLE IF NOT EXISTS access_tokens (
    id INT PRIMARY KEY AUTO_INCREMENT,
    access_token TEXT(2048) NOT NULL,
    user_id INT NOT NULL,
    refresh_token TEXT(2048) NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- AIチャットセッション（会話のグループ単位）
CREATE TABLE IF NOT EXISTS ai_chat_sessions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    title VARCHAR(255),                    -- セッションのタイトル（自動生成 or ユーザー設定）
    related_room_id INT,                   -- chat_roomsとの関連（会話レビューの場合）
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (related_room_id) REFERENCES chat_rooms(id) ON DELETE SET NULL,
    INDEX idx_user_created (user_id, created_at DESC)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- AIチャットメッセージ（個々のやり取り）
CREATE TABLE IF NOT EXISTS ai_chat_messages (
    id INT PRIMARY KEY AUTO_INCREMENT,
    session_id INT NOT NULL,
    user_id INT NOT NULL,
    role ENUM('user', 'assistant') NOT NULL,  -- 発言者
    content TEXT NOT NULL,                     -- メッセージ内容
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (session_id) REFERENCES ai_chat_sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_session_created (session_id, created_at)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


-- ┌─────────────────┐     ┌──────────────────────┐
-- │     users       │────→│   ai_chat_sessions   │
-- └─────────────────┘     └──────────────────────┘
--                                   │
--                                   ↓
--                         ┌──────────────────────┐
--                         │   ai_chat_messages   │
--                         └──────────────────────┘



-- DynamoDB（NoSQL）テーブル設計メモ
-- Table Name: fre_style_ai_chat
-- sender_id (文字列)
-- timestamp (数値)
-- content
-- is_user
-- sender_id (文字列)
-- timestamp (数値)
-- content	
-- is_user
