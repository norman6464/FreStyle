-- ===========================================================================
-- 007_multitenancy_foundation.sql
-- ===========================================================================
-- 目的: B2B SaaS 化に向けた最小限のスキーマ拡張（MariaDB 用、非破壊）
--   - companies テーブル新設
--   - company_signup_applications, invitations を新設
--   - users に company_id, role カラム追加（NULL 許容で開始）
--   - 既存ドメインテーブルに company_id 列を追加（NULL 許容で開始）
--   - FreStyle 社（id=1）を INSERT、既存全データを company_id=1 にバックフィル
--   - 河野拓真（resjimkalto89890@gmail.com）を super_admin に昇格
--
-- 後続 (Phase 0c) で:
--   - PostgreSQL に切替
--   - company_id を NOT NULL 化
--   - Row-Level Security ポリシーを有効化
--
-- 冪等性: IF NOT EXISTS / INSERT IGNORE / 条件付き UPDATE で何度実行しても OK。
-- ===========================================================================

-- ----------------------------------------------------------------------------
-- 1) companies
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS companies (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    slug            VARCHAR(50)  NOT NULL UNIQUE,
    name            VARCHAR(200) NOT NULL,
    display_name    VARCHAR(200),
    plan            VARCHAR(20)  NOT NULL DEFAULT 'free',
    status          VARCHAR(20)  NOT NULL DEFAULT 'active',
    contract_start  DATE,
    contract_end    DATE,
    max_seats       INT,
    metadata        JSON,
    created_at      DATETIME     DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE INDEX idx_companies_status ON companies(status);

-- ----------------------------------------------------------------------------
-- 2) company_signup_applications（B2B 申請承認フロー用）
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS company_signup_applications (
    id                  INT AUTO_INCREMENT PRIMARY KEY,
    company_name        VARCHAR(200) NOT NULL,
    contact_email       VARCHAR(254) NOT NULL,
    contact_name        VARCHAR(200) NOT NULL,
    expected_seats      INT,
    use_case_summary    TEXT,
    status              VARCHAR(20)  NOT NULL DEFAULT 'pending',
    reviewed_by         INT,
    reviewed_at         DATETIME,
    created_company_id  INT,
    created_at          DATETIME     DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (created_company_id) REFERENCES companies(id) ON DELETE SET NULL
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE INDEX idx_signup_applications_status ON company_signup_applications(status);

-- ----------------------------------------------------------------------------
-- 3) invitations（CompanyAdmin → Trainee 招待）
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS invitations (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    company_id      INT          NOT NULL,
    email           VARCHAR(254) NOT NULL,
    role            VARCHAR(30)  NOT NULL DEFAULT 'trainee',
    token           VARCHAR(64)  NOT NULL UNIQUE,
    invited_by      INT,
    expires_at      DATETIME     NOT NULL,
    accepted_at     DATETIME,
    accepted_user_id INT,
    created_at      DATETIME     DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (accepted_user_id) REFERENCES users(id) ON DELETE SET NULL
) DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

CREATE INDEX idx_invitations_company_email ON invitations(company_id, email);

-- ----------------------------------------------------------------------------
-- 4) FreStyle 社の seed
-- ----------------------------------------------------------------------------
INSERT INTO companies (id, slug, name, display_name, plan, status, max_seats)
VALUES (1, 'frestyle', 'FreStyle 株式会社', 'FreStyle (test tenant)', 'enterprise', 'active', 1000)
ON DUPLICATE KEY UPDATE name = VALUES(name), display_name = VALUES(display_name);

-- ----------------------------------------------------------------------------
-- 5) users に company_id / role を追加 + バックフィル
-- ----------------------------------------------------------------------------
-- 列追加は冪等にできないため、ここでは "存在しなければ ALTER" 戦略を採用。
-- MariaDB はバージョンにより構文差があるので、安全のため try-catch 的に
-- ストアドプロシージャを使う。

DELIMITER $$
CREATE PROCEDURE IF NOT EXISTS _add_column_if_missing(
    IN tbl VARCHAR(64), IN col VARCHAR(64), IN col_def TEXT
)
BEGIN
    DECLARE n INT;
    SELECT COUNT(*) INTO n FROM information_schema.COLUMNS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = tbl AND COLUMN_NAME = col;
    IF n = 0 THEN
        SET @ddl = CONCAT('ALTER TABLE ', tbl, ' ADD COLUMN ', col, ' ', col_def);
        PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
    END IF;
END$$
DELIMITER ;

CALL _add_column_if_missing('users', 'company_id', 'INT NULL, ADD INDEX idx_users_company_id (company_id), ADD CONSTRAINT fk_users_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE RESTRICT');
CALL _add_column_if_missing('users', 'role',       'VARCHAR(30) NOT NULL DEFAULT ''trainee''');
CALL _add_column_if_missing('users', 'deleted_at', 'DATETIME NULL');

-- 既存ユーザーは全員 FreStyle 社の trainee 扱いでバックフィル
UPDATE users SET company_id = 1 WHERE company_id IS NULL;
UPDATE users SET role       = 'trainee' WHERE role IS NULL OR role = '';

-- ----------------------------------------------------------------------------
-- 6) 河野拓真 を super_admin に昇格
-- ----------------------------------------------------------------------------
-- user_identities.provider_sub で Cognito の sub と紐付いているはず。
-- email 一致でも昇格する（保険）。
UPDATE users
   SET role = 'super_admin', company_id = NULL
 WHERE email = 'resjimkalto89890@gmail.com';

-- ----------------------------------------------------------------------------
-- 7) 主要ドメインテーブルに company_id を追加 + バックフィル
-- ----------------------------------------------------------------------------
-- 全テーブルを 1 つずつ NULL 許容で追加 → 全行 company_id=1 でバックフィル
-- 後続マイグレーションで NOT NULL 化する。

CALL _add_column_if_missing('practice_scenarios',     'company_id', 'INT NULL, ADD INDEX idx_practice_scenarios_company (company_id)');
CALL _add_column_if_missing('conversation_templates', 'company_id', 'INT NULL, ADD INDEX idx_conv_templates_company (company_id)');
CALL _add_column_if_missing('ai_chat_sessions',       'company_id', 'INT NULL, ADD INDEX idx_ai_chat_sessions_company (company_id)');
CALL _add_column_if_missing('scenario_bookmarks',     'company_id', 'INT NULL, ADD INDEX idx_scenario_bookmarks_company (company_id)');
CALL _add_column_if_missing('favorite_phrases',       'company_id', 'INT NULL, ADD INDEX idx_favorite_phrases_company (company_id)');
CALL _add_column_if_missing('daily_goals',            'company_id', 'INT NULL, ADD INDEX idx_daily_goals_company (company_id)');
CALL _add_column_if_missing('session_notes',          'company_id', 'INT NULL, ADD INDEX idx_session_notes_company (company_id)');
CALL _add_column_if_missing('score_goals',            'company_id', 'INT NULL, ADD INDEX idx_score_goals_company (company_id)');
CALL _add_column_if_missing('shared_sessions',        'company_id', 'INT NULL, ADD INDEX idx_shared_sessions_company (company_id)');
CALL _add_column_if_missing('chat_rooms',             'company_id', 'INT NULL, ADD INDEX idx_chat_rooms_company (company_id)');
CALL _add_column_if_missing('weekly_challenges',      'company_id', 'INT NULL, ADD INDEX idx_weekly_challenges_company (company_id)');
CALL _add_column_if_missing('user_challenge_progress','company_id', 'INT NULL, ADD INDEX idx_user_challenge_progress_company (company_id)');
CALL _add_column_if_missing('friendships',            'company_id', 'INT NULL, ADD INDEX idx_friendships_company (company_id)');
CALL _add_column_if_missing('learning_reports',       'company_id', 'INT NULL, ADD INDEX idx_learning_reports_company (company_id)');
CALL _add_column_if_missing('communication_scores',   'company_id', 'INT NULL, ADD INDEX idx_communication_scores_company (company_id)');
CALL _add_column_if_missing('reminder_settings',      'company_id', 'INT NULL, ADD INDEX idx_reminder_settings_company (company_id)');
CALL _add_column_if_missing('notifications',          'company_id', 'INT NULL, ADD INDEX idx_notifications_company (company_id)');

-- 全行を FreStyle 社にバックフィル
UPDATE practice_scenarios     SET company_id = 1 WHERE company_id IS NULL;
UPDATE conversation_templates SET company_id = 1 WHERE company_id IS NULL;
UPDATE ai_chat_sessions       SET company_id = 1 WHERE company_id IS NULL;
UPDATE scenario_bookmarks     SET company_id = 1 WHERE company_id IS NULL;
UPDATE favorite_phrases       SET company_id = 1 WHERE company_id IS NULL;
UPDATE daily_goals            SET company_id = 1 WHERE company_id IS NULL;
UPDATE session_notes          SET company_id = 1 WHERE company_id IS NULL;
UPDATE score_goals            SET company_id = 1 WHERE company_id IS NULL;
UPDATE shared_sessions        SET company_id = 1 WHERE company_id IS NULL;
UPDATE chat_rooms             SET company_id = 1 WHERE company_id IS NULL;
UPDATE weekly_challenges      SET company_id = 1 WHERE company_id IS NULL;
UPDATE user_challenge_progress SET company_id = 1 WHERE company_id IS NULL;
UPDATE friendships            SET company_id = 1 WHERE company_id IS NULL;
UPDATE learning_reports       SET company_id = 1 WHERE company_id IS NULL;
UPDATE communication_scores   SET company_id = 1 WHERE company_id IS NULL;
UPDATE reminder_settings      SET company_id = 1 WHERE company_id IS NULL;
UPDATE notifications          SET company_id = 1 WHERE company_id IS NULL;

-- 後始末
DROP PROCEDURE IF EXISTS _add_column_if_missing;

-- ----------------------------------------------------------------------------
-- 確認用 SELECT（実行後に手動チェック）
-- ----------------------------------------------------------------------------
-- SELECT id, slug, name, plan, status FROM companies;
-- SELECT id, username, email, role, company_id FROM users;
-- SELECT 'practice_scenarios' AS tbl, COUNT(*) c, SUM(company_id IS NULL) nulls FROM practice_scenarios
--   UNION ALL SELECT 'conversation_templates', COUNT(*), SUM(company_id IS NULL) FROM conversation_templates
--   UNION ALL SELECT 'ai_chat_sessions', COUNT(*), SUM(company_id IS NULL) FROM ai_chat_sessions;
