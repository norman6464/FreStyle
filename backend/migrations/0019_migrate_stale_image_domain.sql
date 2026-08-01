-- FRESTYLE-232: 保存済み絶対 URL の旧ドメインを新ドメインへ移行する。
--
-- 画像の表示 URL は「アップロード時点の配信ドメイン + キー」の絶対 URL として
-- DB に保存される。FRESTYLE-226 で旧ドメイン normanblog.com を削除したため、
-- 移行前に保存された URL は名前解決できず、アバターと教材の図が表示できなくなった。
-- S3 のオブジェクトは無傷で新ドメインから配信できるため、文字列の差し替えで復旧する。
--
-- 教材本文には図の URL のほかに auth. / api. / staging. / 送信元メールの説明も含まれる。
-- これらは実際に frestyle.jp 配下へ移設済みなので、同じ置換で正しい値になる。
-- 教材の正本リポジトリ（frestyle-teaching-materials）にも同一の置換を適用済みで、
-- 次回の教材同期で旧ドメインが復活しないようにしてある。
--
-- 置換はドメイン境界に限定する。単純な部分文字列置換だと、旧ドメインを含むだけの
-- 別ドメイン（notnormanblog.com / normanblog.company.com 等）まで壊してしまう。
-- 前後が英数字・ハイフンでないことを先読み / 後読みで確認し、境界の文字を消費しない
-- ため連続出現も取りこぼさない。
--
-- 冪等: 置換対象が無ければ 0 行更新で正常終了する。何度実行しても結果は変わらない。
-- トランザクション: 適用は psql -f で行い外部のトランザクション管理は無いため、
-- この中で BEGIN / COMMIT して原子性を確保する（既存 migration と同じ方式）。

BEGIN;

-- cdn.normanblog.com は実在しないサブドメイン（教材内の例示）。実際の配信元である
-- apex に直す。この行は後段の一括置換より先に実行する必要がある
-- （先に一括置換すると cdn.frestyle.jp という実在しない値になってしまう）。
UPDATE course_chapters
SET content = regexp_replace(
  content,
  'https://cdn\.normanblog\.com(?![A-Za-z0-9-])',
  'https://frestyle.jp',
  'g'
)
WHERE content ~ 'https://cdn\.normanblog\.com(?![A-Za-z0-9-])';

-- プロフィールのアイコン。
UPDATE profiles
SET avatar_url = regexp_replace(
  avatar_url,
  '(?<![A-Za-z0-9-])normanblog\.com(?![A-Za-z0-9-])',
  'frestyle.jp',
  'g'
)
WHERE avatar_url ~ '(?<![A-Za-z0-9-])normanblog\.com(?![A-Za-z0-9-])';

-- 教材本文（図の参照とアプリ内 URL の説明）。
UPDATE course_chapters
SET content = regexp_replace(
  content,
  '(?<![A-Za-z0-9-])normanblog\.com(?![A-Za-z0-9-])',
  'frestyle.jp',
  'g'
)
WHERE content ~ '(?<![A-Za-z0-9-])normanblog\.com(?![A-Za-z0-9-])';

-- 検証: 旧ドメインが残っていたら適用を失敗させる（部分適用のまま完了しない）。
DO $$
DECLARE
  stale_profiles bigint;
  stale_chapters bigint;
  bogus_cdn      bigint;
  corrupted      bigint;
BEGIN
  SELECT count(*) INTO stale_profiles FROM profiles
   WHERE avatar_url ~ '(?<![A-Za-z0-9-])normanblog\.com(?![A-Za-z0-9-])';
  SELECT count(*) INTO stale_chapters FROM course_chapters
   WHERE content ~ '(?<![A-Za-z0-9-])normanblog\.com(?![A-Za-z0-9-])';
  SELECT count(*) INTO bogus_cdn FROM course_chapters WHERE content LIKE '%cdn.frestyle.jp%';

  -- 境界を無視した置換が起きると英数字の直後に新ドメインが現れる（例: notfrestyle.jp）。
  SELECT count(*) INTO corrupted FROM course_chapters WHERE content ~ '[A-Za-z0-9-]frestyle\.jp';

  IF stale_profiles > 0 OR stale_chapters > 0 THEN
    RAISE EXCEPTION
      '旧ドメインが残存: profiles.avatar_url=% 件 / course_chapters.content=% 件',
      stale_profiles, stale_chapters;
  END IF;

  IF bogus_cdn > 0 THEN
    RAISE EXCEPTION '実在しない cdn.frestyle.jp が生成された: % 件', bogus_cdn;
  END IF;

  IF corrupted > 0 THEN
    RAISE EXCEPTION 'ドメイン境界を跨いだ置換が起きた疑い: % 件', corrupted;
  END IF;
END $$;

COMMIT;
