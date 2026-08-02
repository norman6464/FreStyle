-- FRESTYLE-234: 画像の参照を配信ドメイン込みの絶対 URL からルート相対パスへ移行する。
--
-- 絶対 URL で保存していると、配信ドメインを変えるたびに過去の画像が全て参照不能になる
-- （FRESTYLE-232 で実害。アバター 4 件と教材 140 章の図が本番で表示できなくなった）。
-- 画像はアプリと同一オリジンで配信されるため、ドメインを含めない「/notes/...」形式なら
-- ブラウザが現在のドメインを補って解決し、以後ドメインが変わってもデータは無傷で済む。
--
-- 変換対象は画像のパスだけに限定する。教材本文には「このアドレスでアクセスします」と
-- 説明するためのアプリ URL（/invitations/ /login/ /auth/ /courses/ 等）も含まれており、
-- これらを相対化すると説明文として意味が壊れる。CloudFront が画像バケットへ振り分けて
-- いる /notes/ と /profiles/ のみを対象にする。
--
-- 冪等: 変換済みなら 0 行更新で正常終了する。何度実行しても結果は変わらない。
-- トランザクション: 適用は psql -f で行い外部のトランザクション管理は無いため、
-- この中で BEGIN / COMMIT して原子性を確保する（既存 migration と同じ方式）。

BEGIN;

-- プロフィールのアイコン（列の値そのものが画像 URL）。
UPDATE profiles
SET avatar_url = regexp_replace(avatar_url, '^https?://[^/]+(/profiles/)', '\1')
WHERE avatar_url ~ '^https?://[^/]+/profiles/';

-- 教材本文（Markdown に埋め込まれた画像）。
UPDATE course_chapters
SET content = regexp_replace(content, 'https?://[^/\s)"'']+(/(?:notes|profiles)/)', '\1', 'g')
WHERE content ~ 'https?://[^/\s)"'']+/(?:notes|profiles)/';

-- ノート本文（現時点で該当 0 件だが、後から増えても取りこぼさないよう同じ処理を入れる）。
UPDATE notes
SET content = regexp_replace(content, 'https?://[^/\s)"'']+(/(?:notes|profiles)/)', '\1', 'g')
WHERE content ~ 'https?://[^/\s)"'']+/(?:notes|profiles)/';

-- 検証: 画像の絶対 URL が残っていたら適用を失敗させる（部分適用のまま完了しない）。
DO $$
DECLARE
  stale_profiles bigint;
  stale_chapters bigint;
  stale_notes    bigint;
BEGIN
  SELECT count(*) INTO stale_profiles FROM profiles
   WHERE avatar_url ~ '^https?://[^/]+/profiles/';
  SELECT count(*) INTO stale_chapters FROM course_chapters
   WHERE content ~ 'https?://[^/\s)"'']+/(?:notes|profiles)/';
  SELECT count(*) INTO stale_notes FROM notes
   WHERE content ~ 'https?://[^/\s)"'']+/(?:notes|profiles)/';

  IF stale_profiles > 0 OR stale_chapters > 0 OR stale_notes > 0 THEN
    RAISE EXCEPTION
      '画像の絶対 URL が残存: profiles=% 件 / course_chapters=% 件 / notes=% 件',
      stale_profiles, stale_chapters, stale_notes;
  END IF;
END $$;

COMMIT;
