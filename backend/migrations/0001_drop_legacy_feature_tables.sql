-- =====================================================================
-- 0001 — Legacy feature tables DROP migration
-- =====================================================================
-- 目的:
--   PR-A 〜 PR-D で実装側を削除した機能（練習モード / スコア / お気に入り /
--   会話テンプレート / 練習リマインダー / みんなの会話 / ランキング /
--   ウィークリーチャレンジ / 達成バッジ / 練習レベル / 今日の目標 / etc）の
--   PostgreSQL テーブルを物理削除する。
--
-- 想定 DB: AWS RDS PostgreSQL (frestyle-prod-rds-postgres)
-- 実行手順: EC2 踏み台 (test) 経由で psql 接続 → 本ファイルを実行。
--   詳細は frestyle-pdm/docs/migration/0003-drop-legacy-feature-tables.md を参照。
--
-- 冪等性: すべて DROP TABLE IF EXISTS / DROP SEQUENCE IF EXISTS で
--   既に削除済みの環境でも安全に再実行できる。
--
-- ロールバック: PostgreSQL のスナップショット (frestyle-prod-rds-postgres) を
--   作成してから実行することを strongly 推奨。万一誤実行した場合、論理復元の
--   手段はないのでスナップショットからの巻き戻しが唯一の経路。
--
-- 関連:
--   - 実装側の削除は norman6464/FreStyle PR-D (#1644) で merged
--   - AutoMigrate からも該当 domain は除外済（再 deploy しても再作成されない）
-- =====================================================================

BEGIN;

-- 練習モード関連
DROP TABLE IF EXISTS practice_scenarios          CASCADE;
DROP TABLE IF EXISTS scenario_bookmarks          CASCADE;
DROP TABLE IF EXISTS practice_sessions           CASCADE;

-- スコア関連
DROP TABLE IF EXISTS score_cards                 CASCADE;
DROP TABLE IF EXISTS score_goals                 CASCADE;
DROP TABLE IF EXISTS score_trends                CASCADE;

-- 会話テンプレート / お気に入り
DROP TABLE IF EXISTS conversation_templates      CASCADE;
DROP TABLE IF EXISTS favorite_phrases            CASCADE;

-- ゲーミフィケーション系
DROP TABLE IF EXISTS daily_goals                 CASCADE;
DROP TABLE IF EXISTS weekly_challenges           CASCADE;
DROP TABLE IF EXISTS weekly_challenge_progresses CASCADE;
DROP TABLE IF EXISTS rankings                    CASCADE;

-- リマインダー / 共有セッション
DROP TABLE IF EXISTS reminder_settings           CASCADE;
DROP TABLE IF EXISTS shared_sessions             CASCADE;

-- AI チャット附帯（セッションノートは廃止対象、メインの ai_chat_sessions は残す）
DROP TABLE IF EXISTS session_notes               CASCADE;

-- GORM が自動生成した SEQUENCE が残っているケースのクリーンアップ。
-- テーブル削除時に CASCADE で消えるはずだが、過去に手動 INSERT されたものが
-- 残らないよう明示的に DROP する（冪等）。
DROP SEQUENCE IF EXISTS practice_scenarios_id_seq          CASCADE;
DROP SEQUENCE IF EXISTS scenario_bookmarks_id_seq          CASCADE;
DROP SEQUENCE IF EXISTS practice_sessions_id_seq           CASCADE;
DROP SEQUENCE IF EXISTS score_cards_id_seq                 CASCADE;
DROP SEQUENCE IF EXISTS score_goals_id_seq                 CASCADE;
DROP SEQUENCE IF EXISTS score_trends_id_seq                CASCADE;
DROP SEQUENCE IF EXISTS conversation_templates_id_seq      CASCADE;
DROP SEQUENCE IF EXISTS favorite_phrases_id_seq            CASCADE;
DROP SEQUENCE IF EXISTS daily_goals_id_seq                 CASCADE;
DROP SEQUENCE IF EXISTS weekly_challenges_id_seq           CASCADE;
DROP SEQUENCE IF EXISTS weekly_challenge_progresses_id_seq CASCADE;
DROP SEQUENCE IF EXISTS rankings_id_seq                    CASCADE;
DROP SEQUENCE IF EXISTS reminder_settings_id_seq           CASCADE;
DROP SEQUENCE IF EXISTS shared_sessions_id_seq             CASCADE;
DROP SEQUENCE IF EXISTS session_notes_id_seq               CASCADE;

COMMIT;

-- 検証クエリ（実行後の確認用、別途手動で実行）
-- SELECT table_name FROM information_schema.tables
--  WHERE table_schema='public' AND table_type='BASE TABLE'
--  ORDER BY table_name;
-- 期待される残存テーブル: ai_chat_sessions / companies / invitations /
--   learning_reports / notes / notifications / php_exercises / profiles / users
