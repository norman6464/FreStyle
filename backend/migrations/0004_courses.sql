-- 0004_courses.sql
-- PR-A: コース概念の導入
--
-- 1. courses テーブルを作成（GORM AutoMigrate と同じスキーマ）
-- 2. teaching_materials に course_id / order_in_course カラムを追加
-- 3. 各 company ごとに「Web 基礎」コースを作成し、 既存教材をそこに移管
-- 4. course_id を NOT NULL に確定
--
-- 適用先: 本番 (RDS) は EC2 踏み台経由で psql -f
-- 冪等性: IF NOT EXISTS / DO ブロックで guard

BEGIN;

-- 1. courses テーブル
CREATE TABLE IF NOT EXISTS courses (
  id                   BIGSERIAL PRIMARY KEY,
  company_id           BIGINT NOT NULL,
  created_by_user_id   BIGINT NOT NULL,
  title                VARCHAR(200) NOT NULL DEFAULT '',
  description          TEXT NOT NULL DEFAULT '',
  sort_order           INT NOT NULL DEFAULT 100,
  is_published         BOOLEAN NOT NULL DEFAULT FALSE,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_courses_company_id ON courses(company_id);

-- 2. teaching_materials に course_id / order_in_course カラムを追加（nullable で先に追加）
ALTER TABLE teaching_materials
  ADD COLUMN IF NOT EXISTS course_id        BIGINT,
  ADD COLUMN IF NOT EXISTS order_in_course  INT NOT NULL DEFAULT 100;

CREATE INDEX IF NOT EXISTS idx_teaching_materials_course_id ON teaching_materials(course_id);

-- 3. 各 company ごとに「Web 基礎」コースを作成（既存教材がある company のみ）
--    教材作成者の中で最も古い user_id を created_by_user_id に流用する。
DO $$
DECLARE
  rec RECORD;
  new_course_id BIGINT;
BEGIN
  FOR rec IN
    SELECT company_id, MIN(created_by_user_id) AS owner_user_id
    FROM teaching_materials
    WHERE course_id IS NULL
    GROUP BY company_id
  LOOP
    INSERT INTO courses (company_id, created_by_user_id, title, description, sort_order, is_published, created_at, updated_at)
    VALUES (
      rec.company_id,
      rec.owner_user_id,
      'Web 基礎',
      'Web の基本的な知識（HTTP / URI / HTML / REST）を学ぶコース。',
      10,
      TRUE,
      NOW(),
      NOW()
    )
    RETURNING id INTO new_course_id;

    UPDATE teaching_materials
       SET course_id = new_course_id
     WHERE company_id = rec.company_id AND course_id IS NULL;
  END LOOP;
END $$;

-- 4. course_id を NOT NULL に確定（残った NULL があるなら ALTER は失敗するので手当てが必要）
ALTER TABLE teaching_materials
  ALTER COLUMN course_id SET NOT NULL;

COMMIT;
