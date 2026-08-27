-- 0023 (Contract): AI チャット機能の全廃止に伴い、関連テーブルと列を削除する。
--
-- 【何を消すか】
--   1. session_notes      … AI チャットのセッションに紐づくメモ（親と一緒に廃止）
--   2. ai_chat_sessions   … 会話の器
--   3. companies.ai_chat_enabled_for_trainees … 会社単位の AI 有効化設定
--   4. users.ai_chat_enabled                  … 個人単位の上書き設定
--   5. workspaces.ai_chat_enabled_for_trainees … 会社設定の写し（テナント移行の中間列）
--   6. user_daily_activities.ai_chat_count     … 日次活動量の AI 回数
--
-- 【なぜ消すか】
--   機能そのものを廃止した（FRESTYLE-391）。コードからの参照は同じ PR で全て撤去済みで、
--   残すと「読む人が効いていると誤読する死んだ表・列」になる。
--
-- 【順序の理由】
--   session_notes.session_id は ai_chat_sessions を指す運用だが FK は張られていない。
--   とはいえ「子 → 親」の順で消すのが安全（将来 FK が張られていても壊れない）。
--
-- 【不可逆性】
--   このファイルの適用でデータは消える。適用は frestyle-infrastructure の
--   make apply-migration-supabase で行い、適用前に最終確認すること。
--   会話履歴の正本はもともと DynamoDB（fre_style_ai_chat）で、そちらの削除は
--   infra リポの terraform 作業として別途行う。S3 の ai-chat/ prefix も同様。
--
-- 冪等: すべて IF EXISTS。二度流しても安全。

DROP TABLE IF EXISTS session_notes;
DROP TABLE IF EXISTS ai_chat_sessions;

ALTER TABLE companies  DROP COLUMN IF EXISTS ai_chat_enabled_for_trainees;
ALTER TABLE users      DROP COLUMN IF EXISTS ai_chat_enabled;
ALTER TABLE workspaces DROP COLUMN IF EXISTS ai_chat_enabled_for_trainees;

-- 学習カレンダー・連続日数・会社別の学習状況は、この列を含めない合計へコード側を変更済み。
-- 列を消すことで過去日の合計も AI 分を含まなくなる（AI だけ学習した日は「学習なし」に変わる）。
-- これは仕様変更として FRESTYLE-391 に明記してある。
ALTER TABLE user_daily_activities DROP COLUMN IF EXISTS ai_chat_count;
