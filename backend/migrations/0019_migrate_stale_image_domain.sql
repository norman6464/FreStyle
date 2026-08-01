-- FRESTYLE-232: 保存済み絶対 URL の旧ドメインを新ドメインへ移行する。
--
-- 画像の表示 URL は「アップロード時点の配信ドメイン + キー」の絶対 URL として
-- DB に保存される。FRESTYLE-226 で旧ドメイン normanblog.com を削除したため、
-- 移行前に保存された URL は名前解決できず、アバターと教材の図が表示できなくなった。
-- S3 のオブジェクトは無傷で新ドメインから配信できるため、文字列の差し替えで復旧する。
--
-- 教材本文には図の URL のほかに auth. / api. / staging. の説明も含まれる。これらは
-- 実際に frestyle.jp 配下へ移設済みなので、同じ置換で正しい値になる。
-- 教材の正本リポジトリ（frestyle-teaching-materials）にも同一の置換を適用済みで、
-- 次回の教材同期で旧ドメインが復活しないようにしてある。
--
-- 冪等: 置換対象が無ければ 0 行更新で正常終了する。何度実行しても結果は変わらない。

BEGIN;

-- cdn.normanblog.com は実在しないサブドメイン（教材内の例示）。実際の配信元である
-- apex に直す。この行は後段の一括置換より先に実行する必要がある
-- （先に一括置換すると cdn.frestyle.jp という実在しない値になってしまう）。
UPDATE course_chapters
SET content = replace(content, 'https://cdn.normanblog.com', 'https://frestyle.jp')
WHERE content LIKE '%cdn.normanblog.com%';

-- プロフィールのアイコン。
UPDATE profiles
SET avatar_url = replace(avatar_url, 'normanblog.com', 'frestyle.jp')
WHERE avatar_url LIKE '%normanblog.com%';

-- 教材本文（図の参照とアプリ内 URL の説明）。
UPDATE course_chapters
SET content = replace(content, 'normanblog.com', 'frestyle.jp')
WHERE content LIKE '%normanblog.com%';

-- 検証: 旧ドメインが残っていたら適用を失敗させる（部分適用のまま完了しない）。
DO $$
DECLARE
  stale_profiles bigint;
  stale_chapters bigint;
  bogus_cdn      bigint;
BEGIN
  SELECT count(*) INTO stale_profiles FROM profiles WHERE avatar_url LIKE '%normanblog.com%';
  SELECT count(*) INTO stale_chapters FROM course_chapters WHERE content LIKE '%normanblog.com%';
  SELECT count(*) INTO bogus_cdn      FROM course_chapters WHERE content LIKE '%cdn.frestyle.jp%';

  IF stale_profiles > 0 OR stale_chapters > 0 THEN
    RAISE EXCEPTION
      '旧ドメインが残存: profiles.avatar_url=% 件 / course_chapters.content=% 件',
      stale_profiles, stale_chapters;
  END IF;

  IF bogus_cdn > 0 THEN
    RAISE EXCEPTION '実在しない cdn.frestyle.jp が生成された: % 件', bogus_cdn;
  END IF;
END $$;

COMMIT;
