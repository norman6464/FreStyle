# マルチテナントデータモデル設計

## 0. 設計原則

1. **会社（テナント）境界をすべての行に固定**: ユーザー以外のあらゆるドメインテーブルは `company_id` を必須カラムとして持つ。`SuperAdmin` のみ NULL 許容（横断管理用）
2. **アプリ層 + DB 層の二重防御**: Spring Boot のクエリフィルタ（Hibernate Filter / AOP） + PostgreSQL Row-Level Security
3. **JSONB を活かす**: レッスン本体は構造が異なる種別を持つので、JSONB に格納してスキーマ追加なしに拡張可能
4. **論理削除を採用**: `deleted_at TIMESTAMPTZ` を主要テーブルに付け、メンターのコンテンツが事故で消えないように
5. **created_by / updated_by**: 監査用に作成・更新ユーザーを保持（PII リスクが低い箇所では省略）

## 1. テーブル一覧（新スキーマ）

```
┌─────────────────────────────────────────────────────────┐
│ Tenancy                                                  │
│  companies                       (テナント情報)           │
│  company_signup_applications     (利用申請)               │
└─────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────┐
│ Identity                                                 │
│  users (既存を拡張: company_id, role 追加)                │
│  user_identities (Cognito sub マッピング、既存)            │
│  invitations (CompanyAdmin → Trainee 招待)                │
└─────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────┐
│ Learning Platform (新規)                                 │
│  courses, sections, lessons                              │
│  user_course_enrollments                                 │
│  user_lesson_progress                                    │
│  user_quiz_attempts (詳細回答ログ)                         │
│  user_code_submissions (Phase 2)                         │
└─────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────┐
│ Existing (Communication training, kept)                  │
│  practice_scenarios (既存、company_id 追加)              │
│  ai_chat_session, conversation_templates, ...            │
│  scenario_bookmarks, favorite_phrases, daily_goals, ...  │
│  weekly_challenges, user_challenge_progress              │
│  session_notes, score_goals                              │
│  shared_sessions, friendships, room_members, etc.        │
└─────────────────────────────────────────────────────────┘
```

## 2. 新規テーブル DDL（PostgreSQL）

### 2.1 companies

```sql
CREATE TABLE companies (
  id              BIGSERIAL PRIMARY KEY,
  slug            VARCHAR(50)  NOT NULL UNIQUE,            -- URL や招待リンクで使う
  name            VARCHAR(200) NOT NULL,
  display_name    VARCHAR(200),
  plan            VARCHAR(20)  NOT NULL DEFAULT 'free',    -- free / starter / business / enterprise
  status          VARCHAR(20)  NOT NULL DEFAULT 'active',  -- active / suspended / archived
  contract_start  DATE,
  contract_end    DATE,
  max_seats       INTEGER,                                  -- ユーザー上限（plan に応じる）
  metadata        JSONB        NOT NULL DEFAULT '{}'::jsonb,
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_companies_status ON companies(status) WHERE deleted_at IS NULL;
```

### 2.2 company_signup_applications（申請ボックス）

```sql
CREATE TABLE company_signup_applications (
  id                BIGSERIAL PRIMARY KEY,
  company_name      VARCHAR(200) NOT NULL,
  contact_email     VARCHAR(254) NOT NULL,
  contact_name      VARCHAR(200) NOT NULL,
  expected_seats    INTEGER,
  use_case_summary  TEXT,
  status            VARCHAR(20)  NOT NULL DEFAULT 'pending',  -- pending / approved / rejected
  reviewed_by       BIGINT REFERENCES users(id),
  reviewed_at       TIMESTAMPTZ,
  created_company_id BIGINT REFERENCES companies(id),
  created_at        TIMESTAMPTZ  NOT NULL DEFAULT now()
);
```

### 2.3 users 拡張

既存 `users` テーブルに以下を追加:

```sql
ALTER TABLE users
  ADD COLUMN company_id BIGINT REFERENCES companies(id),         -- super_admin は NULL
  ADD COLUMN role       VARCHAR(30) NOT NULL DEFAULT 'trainee',  -- super_admin | company_admin | trainee
  ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_users_company_id ON users(company_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role       ON users(role);

-- 整合性: 一般ユーザー (super_admin 以外) は company_id 必須
ALTER TABLE users ADD CONSTRAINT chk_users_role_company
  CHECK (
    (role = 'super_admin' AND company_id IS NULL)
    OR (role IN ('company_admin', 'trainee') AND company_id IS NOT NULL)
  );
```

### 2.4 invitations（CompanyAdmin → Trainee 招待）

```sql
CREATE TABLE invitations (
  id          BIGSERIAL PRIMARY KEY,
  company_id  BIGINT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  email       VARCHAR(254) NOT NULL,
  role        VARCHAR(30)  NOT NULL DEFAULT 'trainee',
  token       VARCHAR(64)  NOT NULL UNIQUE,
  invited_by  BIGINT REFERENCES users(id),
  expires_at  TIMESTAMPTZ NOT NULL,
  accepted_at TIMESTAMPTZ,
  accepted_user_id BIGINT REFERENCES users(id),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_invitations_company_email ON invitations(company_id, email);
CREATE INDEX idx_invitations_token ON invitations(token);
```

### 2.5 courses → sections → lessons

```sql
CREATE TABLE courses (
  id              BIGSERIAL PRIMARY KEY,
  company_id      BIGINT      NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  title           VARCHAR(200) NOT NULL,
  slug            VARCHAR(120) NOT NULL,
  description     TEXT,
  difficulty      VARCHAR(20),                       -- beginner / intermediate / advanced
  estimated_hours INTEGER,
  status          VARCHAR(20) NOT NULL DEFAULT 'draft',  -- draft / published / archived
  created_by      BIGINT REFERENCES users(id),
  updated_by      BIGINT REFERENCES users(id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at      TIMESTAMPTZ,
  UNIQUE (company_id, slug)
);

CREATE INDEX idx_courses_company_status ON courses(company_id, status) WHERE deleted_at IS NULL;

CREATE TABLE sections (
  id              BIGSERIAL PRIMARY KEY,
  course_id       BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  title           VARCHAR(200) NOT NULL,
  description     TEXT,
  display_order   INTEGER NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (course_id, display_order)
);

CREATE INDEX idx_sections_course ON sections(course_id);

CREATE TABLE lessons (
  id              BIGSERIAL PRIMARY KEY,
  section_id      BIGINT NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
  title           VARCHAR(200) NOT NULL,
  type            VARCHAR(30)  NOT NULL,             -- reading | quiz | coding | shadowing | communication
  content         JSONB        NOT NULL,             -- 種別ごとの本体 (後述)
  display_order   INTEGER      NOT NULL,
  estimated_minutes INTEGER,
  required_to_advance BOOLEAN  NOT NULL DEFAULT false,
  passing_score   INTEGER,                           -- quiz / coding 用 (0-100)
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (section_id, display_order)
);

CREATE INDEX idx_lessons_section_type ON lessons(section_id, type);
CREATE INDEX idx_lessons_content_gin ON lessons USING GIN (content jsonb_path_ops);
```

#### `lessons.content` の JSONB スキーマ

```jsonc
// type = 'reading'
{
  "type": "reading",
  "markdown": "# Git の基本コマンド\n...",
  "attachments": [
    { "kind": "image", "s3_key": "company/1/lessons/1/diagram.png" }
  ]
}

// type = 'quiz'  (MVP の主役)
{
  "type": "quiz",
  "questions": [
    {
      "id": "q1",
      "kind": "single_choice",        // single_choice | true_false
      "text": "git add の役割は？",
      "options": [
        "ファイルをステージングエリアに追加する",
        "リモートリポジトリにプッシュする",
        "ローカルにブランチを作る",
        "コミットを取り消す"
      ],
      "correct_index": 0,
      "explanation": "git add はコミット対象を選ぶコマンド..."
    }
  ]
}

// type = 'coding' (Phase 2)
{
  "type": "coding",
  "languages": ["python", "javascript"],
  "description": "与えられた配列の合計を返す関数 sum を実装せよ",
  "starter_code": { "python": "def sum(arr):\n    pass", "javascript": "function sum(arr){\n  // ...\n}" },
  "test_cases": [
    { "input": "[1,2,3]", "expected_output": "6" },
    { "input": "[]",      "expected_output": "0" }
  ],
  "solution": { "python": "def sum(arr):\n    return __import__('builtins').sum(arr)" },
  "time_limit_ms": 5000,
  "memory_limit_mb": 256
}

// type = 'shadowing' (Phase 2)
{
  "type": "shadowing",
  "language": "python",
  "code": "def hello():\n    print('hello')",
  "explanation_lines": [
    { "line": 1, "comment": "関数定義" },
    { "line": 2, "comment": "標準出力に hello を表示" }
  ]
}

// type = 'communication' (既存のシナリオ ID を参照する)
{
  "type": "communication",
  "scenario_id": 12,
  "evaluation_axes": ["logical_structure", "courtesy", "summary"]
}
```

### 2.6 受講・進捗

```sql
CREATE TABLE user_course_enrollments (
  id           BIGSERIAL PRIMARY KEY,
  user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  course_id    BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  enrolled_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ,
  UNIQUE (user_id, course_id)
);

CREATE TABLE user_lesson_progress (
  id                BIGSERIAL PRIMARY KEY,
  user_id           BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  lesson_id         BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
  status            VARCHAR(20) NOT NULL DEFAULT 'not_started',   -- not_started | in_progress | completed | failed
  best_score        INTEGER,
  attempts          INTEGER NOT NULL DEFAULT 0,
  started_at        TIMESTAMPTZ,
  completed_at      TIMESTAMPTZ,
  last_attempt_data JSONB,                 -- 最新の回答内容（quiz は答え、coding は提出コード）
  UNIQUE (user_id, lesson_id)
);

CREATE INDEX idx_user_lesson_progress_user_status ON user_lesson_progress(user_id, status);

CREATE TABLE user_quiz_attempts (
  id           BIGSERIAL PRIMARY KEY,
  user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  lesson_id    BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
  score        INTEGER NOT NULL,            -- 0-100
  total_count  INTEGER NOT NULL,
  correct_count INTEGER NOT NULL,
  answers      JSONB   NOT NULL,            -- [{"qid":"q1","answer":0,"correct":true}, ...]
  duration_ms  INTEGER,
  submitted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_quiz_attempts_lesson ON user_quiz_attempts(lesson_id, submitted_at DESC);
```

### 2.7 既存テーブルへの `company_id` 付与

主要なドメインテーブルすべてに `company_id BIGINT NOT NULL REFERENCES companies(id)` を追加:

```sql
-- 既存テーブルへ company_id を追加し、FreStyle 社（id=1）でバックフィル後 NOT NULL 化
ALTER TABLE practice_scenarios          ADD COLUMN company_id BIGINT REFERENCES companies(id);
ALTER TABLE conversation_templates      ADD COLUMN company_id BIGINT REFERENCES companies(id);
ALTER TABLE ai_chat_session             ADD COLUMN company_id BIGINT REFERENCES companies(id);
ALTER TABLE scenario_bookmarks          ADD COLUMN company_id BIGINT REFERENCES companies(id);
ALTER TABLE favorite_phrases            ADD COLUMN company_id BIGINT REFERENCES companies(id);
ALTER TABLE daily_goals                 ADD COLUMN company_id BIGINT REFERENCES companies(id);
ALTER TABLE session_notes               ADD COLUMN company_id BIGINT REFERENCES companies(id);
ALTER TABLE score_goals                 ADD COLUMN company_id BIGINT REFERENCES companies(id);
ALTER TABLE shared_sessions             ADD COLUMN company_id BIGINT REFERENCES companies(id);
ALTER TABLE chat_rooms                  ADD COLUMN company_id BIGINT REFERENCES companies(id);
ALTER TABLE weekly_challenges           ADD COLUMN company_id BIGINT REFERENCES companies(id);
ALTER TABLE user_challenge_progress     ADD COLUMN company_id BIGINT REFERENCES companies(id);
ALTER TABLE friendships                 ADD COLUMN company_id BIGINT REFERENCES companies(id);
-- 一括バックフィル
UPDATE practice_scenarios          SET company_id = 1 WHERE company_id IS NULL;
-- ... 他テーブルも同様
-- すべて埋まったら NOT NULL 化
ALTER TABLE practice_scenarios          ALTER COLUMN company_id SET NOT NULL;
-- ...
```

### 2.8 Row-Level Security ポリシー（PostgreSQL）

```sql
-- 例: courses テーブル
ALTER TABLE courses ENABLE ROW LEVEL SECURITY;

CREATE POLICY courses_tenant_read ON courses
  FOR SELECT
  USING (
    current_setting('app.current_role', true) = 'super_admin'
    OR company_id = current_setting('app.current_company_id', true)::bigint
  );

CREATE POLICY courses_tenant_write ON courses
  FOR INSERT WITH CHECK (
    current_setting('app.current_role', true) = 'super_admin'
    OR company_id = current_setting('app.current_company_id', true)::bigint
  );
```

Spring Boot 側は `JdbcTemplate` の `JdbcSession` か `Hibernate ConnectionPreparer` で、リクエスト処理開始時に `SET LOCAL app.current_company_id = ?` を発行する。実装はインターセプタで一元化。

## 3. ロール権限マトリクス

| 操作 | super_admin | company_admin | trainee |
|---|---|---|---|
| 会社一覧・作成・無効化 | ✅ | ❌ | ❌ |
| 利用申請の承認 | ✅ | ❌ | ❌ |
| 自社のメンバー招待・削除 | ❌ | ✅ | ❌ |
| 自社のコース作成・更新・削除 | ❌ | ✅ | ❌ |
| 自社のコース受講・進捗確認（自分） | ❌ | ✅ | ✅ |
| 自社チームの進捗ダッシュボード | ❌ | ✅ | ❌ |
| AI コミュニケーション練習（既存機能） | ❌ | ✅ | ✅ |

## 4. 整合性チェック（CI で自動化）

- 全 Repository クエリに `company_id` 条件があること（lint）
- すべてのドメインテーブルに `company_id` カラムがあること（schema-test）
- `super_admin` 以外の users に `company_id` があること（DB CHECK 制約 + 起動時バリデーション）
- RLS が ENABLED であること（pg_class メタデータ確認）

## 5. パフォーマンス考慮

- 全クエリに `company_id` フィルタが入る → 主要テーブルは `(company_id, ...)` の複合インデックスを優先
- `lessons.content` の JSONB は GIN index（`jsonb_path_ops`）で部分検索を高速化
- 進捗集計は実体テーブル + マテリアライズドビュー（`mv_company_user_progress`）で重い集計を高速化（Phase 1.5 以降）

## 6. 関連

- [SaaS ビジョン](./saas-vision.md)
- [DBMS 選定](./dbms-choice.md)
- [Phase 0 マイグレーション計画](../migration/0001-multitenancy-foundation.md)
