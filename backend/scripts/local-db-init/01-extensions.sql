-- ローカル開発 DB の初期化。空のデータディレクトリで一度だけ実行される
-- (docker-entrypoint-initdb.d の仕様。既存 volume がある場合は流れない)。
--
-- pg_stat_statements: 実行されたクエリを正規化して累積統計を取る。どのクエリが
-- 遅いかを「体感」ではなく実測で特定するために入れる。compose 側で
-- shared_preload_libraries に指定済みなので、ここでは拡張を作るだけでよい。
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- 統計をリセットして計測し直すときは psql で:
--   SELECT pg_stat_statements_reset();
-- 遅いクエリの確認:
--   SELECT calls, round(mean_exec_time::numeric, 2) AS mean_ms,
--          round(total_exec_time::numeric, 2) AS total_ms, query
--     FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 20;
