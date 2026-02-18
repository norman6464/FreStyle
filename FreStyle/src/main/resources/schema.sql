-- users
CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(254) UNIQUE NOT NULL,
    icon_url VARCHAR(255),
    bio TEXT,
    status VARCHAR(100),
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

-- アクセストークンを格納するテーブル（まだ検証段階なのでプロダクション環境に反映しない）
CREATE TABLE IF NOT EXISTS access_tokens (
    id INT PRIMARY KEY AUTO_INCREMENT,
    access_token TEXT NOT NULL,
    user_id INT NOT NULL,
    refresh_token TEXT NOT NULL,
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
    scene VARCHAR(50) DEFAULT NULL,        -- フィードバックシーン（meeting, one_on_one, email, presentation, negotiation）
    session_type VARCHAR(20) DEFAULT 'normal', -- セッション種別（normal, practice）
    scenario_id INT DEFAULT NULL,          -- 練習モードのシナリオID
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


-- コミュニケーションスコア（AIフィードバック時の評価軸スコア）
CREATE TABLE IF NOT EXISTS communication_scores (
    id INT PRIMARY KEY AUTO_INCREMENT,
    session_id INT NOT NULL,
    user_id INT NOT NULL,
    axis_name VARCHAR(50) NOT NULL,
    score INT NOT NULL,
    comment TEXT,
    scene VARCHAR(50),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES ai_chat_sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_created (user_id, created_at DESC)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- 練習シナリオ（ビジネスコミュニケーション練習モード用）
CREATE TABLE IF NOT EXISTS practice_scenarios (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    role_name VARCHAR(100) NOT NULL,
    difficulty VARCHAR(20) DEFAULT 'intermediate',
    system_prompt TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- 初期シナリオデータ（ITエンジニア向け12シナリオ）

-- カテゴリ1: 顧客折衝（5シナリオ）
INSERT IGNORE INTO practice_scenarios (id, name, description, category, role_name, difficulty, system_prompt) VALUES
(1, '本番障害の緊急報告', '本番環境で重大な障害が発生。顧客への第一報と状況説明を行う', 'customer', '怒っている顧客（SIer企業のPM）', 'intermediate', 'デプロイ直後に本番環境で障害が発生しました。顧客のシステムが30分以上停止しており、顧客PMから緊急連絡が入っています。顧客は非常に怒っており、原因究明と復旧見込みを求めています。'),
(2, '要件変更の影響説明', 'スプリント中盤での要件変更依頼。スケジュール・コストへの影響を説明する', 'customer', '要件追加を求める顧客（事業部長）', 'intermediate', '開発中のWebアプリケーションについて、顧客の事業部長から「やっぱりこの機能も追加してほしい」と要望がありました。追加するとスケジュールが2週間延びる見込みです。顧客は「簡単でしょ？」と考えています。'),
(3, '見積もり提案・交渉', '新規開発案件の見積もりを提示し、予算内に収まるよう交渉する', 'customer', '予算を削りたい顧客（情シス部長）', 'advanced', '新規システム開発の見積もりを提示する場面です。顧客の予算は1000万円ですが、こちらの見積もりは1500万円です。顧客は「もっと安くならないか」と値引きを求めています。'),
(4, 'リリース延期の報告', 'テストで重大バグが発見され、予定通りのリリースが困難になった状況を報告する', 'customer', '納期を重視する顧客（プロジェクトオーナー）', 'advanced', 'リリース予定日の1週間前に重大なセキュリティバグが発見されました。修正には最低2週間必要です。顧客はこのリリースに合わせてマーケティング施策を準備しており、延期は大きな損失になります。'),
(5, '技術提案プレゼン', '顧客にクラウド移行を提案するプレゼンテーション', 'customer', '技術に詳しくない経営層（CEO）', 'intermediate', '顧客企業のオンプレミスシステムをAWSに移行する提案を行います。CEOは技術には詳しくないですが、コスト削減と事業のスピードアップに関心があります。');

-- カテゴリ2: シニアエンジニア・上司とのコミュニケーション（5シナリオ）
INSERT IGNORE INTO practice_scenarios (id, name, description, category, role_name, difficulty, system_prompt) VALUES
(6, '設計レビューでの意見対立', 'マイクロサービス vs モノリスで意見が分かれた設計レビュー', 'senior', '経験豊富なテックリード', 'advanced', '新規プロジェクトのアーキテクチャ設計レビューです。あなたはマイクロサービスを推していますが、テックリードはモノリスファーストを主張しています。チームの規模は5人で、まだMVP段階です。テックリードは10年以上の経験があり、論理的に反論してきます。'),
(7, 'コードレビューのフィードバック受け入れ', '厳しいコードレビューコメントへの対応', 'senior', '厳格なシニアエンジニア', 'beginner', 'あなたが書いたPull Requestに対して、シニアエンジニアから厳しいコードレビューコメントが20件以上付きました。設計パターンの選択、命名規則、テストカバレッジなど多岐にわたる指摘です。'),
(8, '進捗遅延の上司への報告', 'タスクが予定より大幅に遅れていることを上司に報告する', 'senior', '進捗を気にするエンジニアリングマネージャー', 'intermediate', '見積もり3日のタスクが1週間経っても完了していません。技術的な難しさを事前に見積もれなかったことが原因です。マネージャーは今週中の完了を期待しています。'),
(9, '技術負債の改善提案', 'レガシーコードのリファクタリングを上司に提案する', 'senior', 'ビジネス優先のプロダクトマネージャー', 'intermediate', 'プロダクトのコードベースが肥大化し、新機能の開発速度が低下しています。あなたは2スプリント分のリファクタリング期間を確保したいと考えていますが、PMは新機能のリリースを優先したいと考えています。'),
(10, '1on1での成長相談', 'キャリアの方向性について上司と相談する', 'senior', '面倒見の良いエンジニアリングマネージャー', 'beginner', '入社2年目のあなたが、今後のキャリアパスについてマネージャーと1on1で相談する場面です。バックエンドを中心にやってきましたが、インフラやフロントエンドにも興味があり、どう成長していくべきか悩んでいます。');

-- カテゴリ3: チーム内コミュニケーション（2シナリオ）
INSERT IGNORE INTO practice_scenarios (id, name, description, category, role_name, difficulty, system_prompt) VALUES
(11, '新メンバーへの技術オンボーディング', '新しくチームに入ったジュニアエンジニアに技術スタックを説明する', 'team', '入社したばかりのジュニアエンジニア', 'beginner', '新卒で入社したジュニアエンジニアがあなたのチームに配属されました。プロジェクトで使用しているReact + Spring Boot + AWSの技術スタックについて説明し、最初のタスクをアサインする場面です。本人はJavaの基礎はありますが、実務経験はありません。'),
(12, 'チーム間の仕様調整', 'フロントエンドチームとバックエンドチームのAPI仕様のすり合わせ', 'team', 'フロントエンドチームのリーダー', 'intermediate', 'あなたはバックエンドチームのエンジニアです。フロントエンドチームのリーダーと新機能のAPI仕様を調整する必要があります。フロントエンド側はGraphQLを希望していますが、バックエンドチームはREST APIで統一したいと考えています。');

-- 学習レポート（月次レポート）
CREATE TABLE IF NOT EXISTS learning_reports (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    year INT NOT NULL,
    month INT NOT NULL,
    total_sessions INT NOT NULL DEFAULT 0,
    average_score DOUBLE NOT NULL DEFAULT 0.0,
    previous_average_score DOUBLE,
    best_axis VARCHAR(50),
    worst_axis VARCHAR(50),
    practice_days INT NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_user_year_month (user_id, year, month),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- フレンド（フォロー/フォロワー）
CREATE TABLE IF NOT EXISTS friendships (
    id INT PRIMARY KEY AUTO_INCREMENT,
    follower_id INT NOT NULL,
    following_id INT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_follower_following (follower_id, following_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_follower (follower_id),
    INDEX idx_following (following_id)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- デイリー目標
CREATE TABLE IF NOT EXISTS daily_goals (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    goal_date DATE NOT NULL,
    target INT NOT NULL,
    completed INT NOT NULL DEFAULT 0,

    UNIQUE KEY uk_user_goal_date (user_id, goal_date),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- シナリオブックマーク
CREATE TABLE IF NOT EXISTS scenario_bookmarks (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    scenario_id INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_user_scenario (user_id, scenario_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (scenario_id) REFERENCES practice_scenarios(id) ON DELETE CASCADE
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- スコア目標
CREATE TABLE IF NOT EXISTS score_goals (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    goal_score DOUBLE NOT NULL,
    updated_at DATETIME NOT NULL,

    UNIQUE KEY uk_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- お気に入りフレーズ
CREATE TABLE IF NOT EXISTS favorite_phrases (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    original_text TEXT NOT NULL,
    rephrased_text TEXT NOT NULL,
    pattern VARCHAR(50) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- 通知
CREATE TABLE IF NOT EXISTS notifications (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    related_id INT,
    created_at DATETIME NOT NULL,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_created (user_id, created_at DESC)
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- セッションノート
CREATE TABLE IF NOT EXISTS session_notes (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    session_id INT NOT NULL,
    note TEXT NOT NULL,
    updated_at DATETIME NOT NULL,

    UNIQUE KEY uk_user_session (user_id, session_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES ai_chat_sessions(id) ON DELETE CASCADE
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
