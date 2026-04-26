-- ============================================================================
-- 009_seed_test_users_and_laravel_course.sql
-- ============================================================================
-- 目的:
--   1) FreStyle 株式会社 (id=1) のテストデータ
--      - 河野拓真 (resjimkalto89890@gmail.com) を CompanyAdmin (メンター) に再分類
--      - fca2406070005@edu.fca.ac.jp を Trainee (新卒研修対象) として登録
--   2) Laravel 研修コースのサンプル教材 (Course → Section → Lesson)
--      メンターが作る教材の見本として
-- ============================================================================

BEGIN;

-- ----------------------------------------------------------------------------
-- 1. 河野拓真を CompanyAdmin (FreStyle 社のメンター) に再分類
-- ----------------------------------------------------------------------------
-- ※ super_admin（SaaS 運営）と兼任したい場合は別途 Cognito ユーザーを分けるか、
--   role に複数値を持たせる仕組みを Phase 1 で導入する。
--   当面、河野は「FreStyle 社のメンター」として教材を作る位置付け。

UPDATE users
   SET role = 'company_admin', company_id = 1
 WHERE email = 'resjimkalto89890@gmail.com';

-- ----------------------------------------------------------------------------
-- 2. fca2406070005@edu.fca.ac.jp を Trainee として登録
-- ----------------------------------------------------------------------------
-- Cognito 側は既に Admin Create User 済 (sub=17c44a88-6011-700f-b95b-53dfbe4d7daa)
-- 初回ログイン時にパスワード変更が必要 (FORCE_CHANGE_PASSWORD)

INSERT INTO users (company_id, role, username, email, is_active)
VALUES (1, 'trainee', '新卒研修生（テスト）', 'fca2406070005@edu.fca.ac.jp', true)
ON CONFLICT (email) DO UPDATE SET
  company_id = EXCLUDED.company_id,
  role       = EXCLUDED.role,
  username   = EXCLUDED.username,
  is_active  = EXCLUDED.is_active;

-- Cognito sub と紐付け (初回ログイン前に登録しておく)
INSERT INTO user_identities (user_id, provider, provider_sub)
SELECT u.id, 'cognito', '17c44a88-6011-700f-b95b-53dfbe4d7daa'
  FROM users u
 WHERE u.email = 'fca2406070005@edu.fca.ac.jp'
ON CONFLICT (provider, provider_sub) DO NOTHING;

-- ----------------------------------------------------------------------------
-- 3. Laravel コース (FreStyle 社所属、CompanyAdmin = 河野が作成)
-- ----------------------------------------------------------------------------

INSERT INTO courses (id, company_id, title, slug, description, difficulty, estimated_hours, status, created_by)
VALUES (
  1, 1,
  'Laravel 入門 - PHP フルスタック開発',
  'laravel-intro',
  '社内バックエンド開発の中核フレームワーク Laravel を、ゼロから本番運用レベルまで一気通貫で学ぶ。MVC・Eloquent ORM・認証・テスト・デプロイまで。',
  'beginner',
  20,
  'published',
  (SELECT id FROM users WHERE email = 'resjimkalto89890@gmail.com')
)
ON CONFLICT (id) DO UPDATE SET
  title           = EXCLUDED.title,
  slug            = EXCLUDED.slug,
  description     = EXCLUDED.description,
  difficulty      = EXCLUDED.difficulty,
  estimated_hours = EXCLUDED.estimated_hours,
  status          = EXCLUDED.status;

SELECT setval(pg_get_serial_sequence('courses','id'), (SELECT MAX(id) FROM courses), true);

-- ----------------------------------------------------------------------------
-- 4. セクション
-- ----------------------------------------------------------------------------

INSERT INTO sections (id, course_id, title, description, display_order) VALUES
  (1, 1, '環境構築',          'Laravel を動かすための PHP / Composer / Sail 環境を整える',     1),
  (2, 1, 'ルーティングと MVC', 'リクエスト → Controller → View の流れを理解する',              2),
  (3, 1, 'Eloquent ORM',     'マイグレーション・モデル・リレーションでデータを操作する',        3),
  (4, 1, '認証と認可',         'Breeze / Sanctum を使ったログインと権限管理',                  4),
  (5, 1, 'テストとデプロイ',   'PHPUnit でテスト、Forge / ECS でデプロイ',                     5)
ON CONFLICT (course_id, display_order) DO UPDATE SET
  title         = EXCLUDED.title,
  description   = EXCLUDED.description;

SELECT setval(pg_get_serial_sequence('sections','id'), (SELECT MAX(id) FROM sections), true);

-- ----------------------------------------------------------------------------
-- 5. レッスン (各セクション 2-3 個ずつ)
-- ----------------------------------------------------------------------------

-- セクション 1: 環境構築
INSERT INTO lessons (id, section_id, title, type, content, display_order, estimated_minutes, required_to_advance) VALUES
  (1, 1, 'Laravel とは何か', 'reading', $$
{
  "type": "reading",
  "markdown": "# Laravel とは何か\n\nLaravel は **PHP で書かれた Web アプリケーションフレームワーク** です。Ruby on Rails や Django と並ぶ「フルスタック型フレームワーク」で、認証・DB アクセス・ルーティング・テスト・キャッシュ・キューなど、Web アプリに必要な機能が一通り揃っています。\n\n## なぜ社内で Laravel を使うか\n\n- **生産性が高い**: Eloquent ORM / Blade テンプレート / Artisan コマンドで定型コードが激減\n- **学習リソースが豊富**: 日本語ドキュメント・書籍が多い\n- **エコシステム**: Forge (デプロイ), Vapor (Lambda), Nova (管理画面), Sanctum (API 認証) 等\n- **PHP の強み**: レンタルサーバーでも動く、運用知見が広い\n\n## バージョン\n\n本コースでは **Laravel 11 LTS** を扱います。PHP 8.2 以上が必須。\n\n## このコースの到達点\n\n1. Sail でローカル開発環境を立ち上げる\n2. CRUD アプリ（タスク管理）を作る\n3. PHPUnit でテストを書く\n4. ECS Fargate にデプロイする\n\n## 学習時間の目安\n\n約 **20 時間**（業務時間を確保して 1 週間で完走）"
}
$$::jsonb, 1, 30, false),

  (2, 1, '環境構築チェックリスト (クイズ)', 'quiz', $$
{
  "type": "quiz",
  "questions": [
    {
      "id": "q1",
      "kind": "single_choice",
      "text": "Laravel 11 を動かすために必要な PHP の最低バージョンは？",
      "options": ["7.4", "8.0", "8.2", "8.4"],
      "correct_index": 2,
      "explanation": "Laravel 11 は PHP 8.2 以上を要求します。"
    },
    {
      "id": "q2",
      "kind": "single_choice",
      "text": "Laravel のパッケージマネージャは？",
      "options": ["npm", "Composer", "Yarn", "pip"],
      "correct_index": 1,
      "explanation": "PHP の依存管理は Composer。フロントは Vite + npm を併用します。"
    },
    {
      "id": "q3",
      "kind": "true_false",
      "text": "Laravel Sail は Docker を使ったローカル開発環境である",
      "options": ["○", "×"],
      "correct_index": 0,
      "explanation": "Sail は docker-compose のラッパーで、PHP / MySQL / Redis 等をワンコマンドで起動できます。"
    }
  ]
}
$$::jsonb, 2, 10, true),

  (3, 1, '初プロジェクトを作る (写経)', 'reading', $$
{
  "type": "reading",
  "markdown": "# 初プロジェクトを作る\n\n## 1. 新規プロジェクト\n\n```bash\ncomposer create-project laravel/laravel my-first-app\ncd my-first-app\n```\n\n## 2. Sail で起動\n\n```bash\nphp artisan sail:install --with=mysql,redis\n./vendor/bin/sail up -d\n./vendor/bin/sail artisan migrate\n```\n\n## 3. ブラウザで確認\n\n[http://localhost](http://localhost) にアクセスして Laravel のウェルカム画面が出れば成功。\n\n## トラブル\n\n- **port 80 が使えない**: `APP_PORT=8080` を `.env` に追記\n- **MySQL 接続エラー**: `DB_HOST=mysql` (Sail の場合 localhost ではなく mysql)"
}
$$::jsonb, 3, 30, false)
ON CONFLICT (section_id, display_order) DO UPDATE SET
  title             = EXCLUDED.title,
  type              = EXCLUDED.type,
  content           = EXCLUDED.content,
  estimated_minutes = EXCLUDED.estimated_minutes;

-- セクション 2: ルーティングと MVC
INSERT INTO lessons (id, section_id, title, type, content, display_order, estimated_minutes, required_to_advance) VALUES
  (4, 2, 'ルートと Controller', 'reading', $$
{
  "type": "reading",
  "markdown": "# ルートと Controller\n\n## ルート定義\n\n`routes/web.php`:\n\n```php\nuse App\\Http\\Controllers\\TaskController;\n\nRoute::get('/tasks', [TaskController::class, 'index']);\nRoute::post('/tasks', [TaskController::class, 'store']);\nRoute::get('/tasks/{task}', [TaskController::class, 'show']);\n```\n\n## Controller 作成\n\n```bash\nphp artisan make:controller TaskController --resource\n```\n\n`app/Http/Controllers/TaskController.php` に 7 メソッド (index/create/store/show/edit/update/destroy) のテンプレが生成される。\n\n## ルートと Controller の対応\n\n| HTTP | URL | Controller method | 用途 |\n|---|---|---|---|\n| GET  | /tasks         | index   | 一覧 |\n| GET  | /tasks/create  | create  | 作成フォーム |\n| POST | /tasks         | store   | 作成処理 |\n| GET  | /tasks/{task}  | show    | 詳細 |\n| PUT  | /tasks/{task}  | update  | 更新処理 |\n| DELETE | /tasks/{task} | destroy | 削除 |"
}
$$::jsonb, 1, 25, false),

  (5, 2, 'MVC の理解度チェック (クイズ)', 'quiz', $$
{
  "type": "quiz",
  "questions": [
    {
      "id": "q1",
      "kind": "single_choice",
      "text": "Laravel の M, V, C はそれぞれ何を表すか？",
      "options": [
        "Manager, Viewer, Controller",
        "Model, View, Controller",
        "Module, Vendor, Container",
        "Migration, Validation, Configuration"
      ],
      "correct_index": 1,
      "explanation": "Model (DB) → Controller (ビジネスロジック) → View (Blade テンプレート) の流れ"
    },
    {
      "id": "q2",
      "kind": "single_choice",
      "text": "RESTful な「タスクの一覧表示」に対応する Controller method は？",
      "options": ["create", "index", "show", "store"],
      "correct_index": 1,
      "explanation": "index は一覧、show は単一レコードの詳細"
    }
  ]
}
$$::jsonb, 2, 8, true)
ON CONFLICT (section_id, display_order) DO UPDATE SET
  title             = EXCLUDED.title,
  type              = EXCLUDED.type,
  content           = EXCLUDED.content,
  estimated_minutes = EXCLUDED.estimated_minutes;

-- セクション 3: Eloquent ORM
INSERT INTO lessons (id, section_id, title, type, content, display_order, estimated_minutes, required_to_advance) VALUES
  (6, 3, 'Migration と Model', 'reading', $$
{
  "type": "reading",
  "markdown": "# Migration と Model\n\n## Migration を作る\n\n```bash\nphp artisan make:migration create_tasks_table\n```\n\n`database/migrations/xxxx_create_tasks_table.php`:\n\n```php\npublic function up(): void {\n    Schema::create('tasks', function (Blueprint $table) {\n        $table->id();\n        $table->string('title');\n        $table->text('description')->nullable();\n        $table->boolean('is_done')->default(false);\n        $table->foreignId('user_id')->constrained()->onDelete('cascade');\n        $table->timestamps();\n    });\n}\n```\n\n## 適用\n\n```bash\nphp artisan migrate\n```\n\n## Model を作る\n\n```bash\nphp artisan make:model Task\n```\n\n`app/Models/Task.php`:\n\n```php\nclass Task extends Model {\n    protected $fillable = ['title', 'description', 'is_done', 'user_id'];\n    \n    public function user() {\n        return $this->belongsTo(User::class);\n    }\n}\n```\n\n## 操作例\n\n```php\n// 作成\nTask::create(['title' => '買い物', 'user_id' => 1]);\n\n// 取得\n$tasks = Task::where('is_done', false)->orderBy('created_at', 'desc')->get();\n\n// リレーション\n$user = User::find(1);\nforeach ($user->tasks as $task) { echo $task->title; }\n```"
}
$$::jsonb, 1, 40, false),

  (7, 3, 'Eloquent クイズ', 'quiz', $$
{
  "type": "quiz",
  "questions": [
    {
      "id": "q1",
      "kind": "single_choice",
      "text": "Laravel の Migration は何のためのファイル？",
      "options": [
        "本番環境にコードをデプロイする",
        "DB のスキーマ変更をバージョン管理する",
        "ページのリダイレクト設定",
        "テストデータを投入する"
      ],
      "correct_index": 1,
      "explanation": "Migration は DB スキーマの DDL を PHP で書き、git 管理することでチーム間の DB 構造を統一する仕組み"
    },
    {
      "id": "q2",
      "kind": "single_choice",
      "text": "1 ユーザーが多数のタスクを持つ関係を Eloquent で表現すると？",
      "options": [
        "$this->hasOne(Task::class)",
        "$this->hasMany(Task::class)",
        "$this->belongsTo(Task::class)",
        "$this->morphMany(Task::class)"
      ],
      "correct_index": 1,
      "explanation": "1 対多は hasMany、多対 1 は belongsTo"
    },
    {
      "id": "q3",
      "kind": "true_false",
      "text": "Eloquent の $fillable は Mass Assignment 脆弱性対策の機能である",
      "options": ["○", "×"],
      "correct_index": 0,
      "explanation": "$fillable に書かれていないフィールドは create / update で一括代入されない"
    }
  ]
}
$$::jsonb, 2, 12, true)
ON CONFLICT (section_id, display_order) DO UPDATE SET
  title             = EXCLUDED.title,
  type              = EXCLUDED.type,
  content           = EXCLUDED.content,
  estimated_minutes = EXCLUDED.estimated_minutes;

-- セクション 4: 認証と認可
INSERT INTO lessons (id, section_id, title, type, content, display_order, estimated_minutes, required_to_advance) VALUES
  (8, 4, 'Laravel Breeze で認証を 5 分で', 'reading', $$
{
  "type": "reading",
  "markdown": "# Laravel Breeze で認証を 5 分で\n\n## インストール\n\n```bash\ncomposer require laravel/breeze --dev\nphp artisan breeze:install blade\nnpm install && npm run build\nphp artisan migrate\n```\n\nこれだけで:\n\n- ログイン / 新規登録 / パスワードリセット画面\n- Email 認証フロー\n- ミドルウェア (`auth`, `verified`)\n- ユーザー設定画面\n\nが揃います。\n\n## 認証保護\n\n`routes/web.php`:\n\n```php\nRoute::middleware('auth')->group(function () {\n    Route::resource('tasks', TaskController::class);\n});\n```\n\n## 認可 (Policy)\n\n```bash\nphp artisan make:policy TaskPolicy --model=Task\n```\n\n```php\npublic function update(User $user, Task $task): bool {\n    return $user->id === $task->user_id;\n}\n```\n\nController で:\n\n```php\n$this->authorize('update', $task);\n```\n\n他人のタスクを編集しようとすると 403。"
}
$$::jsonb, 1, 30, false)
ON CONFLICT (section_id, display_order) DO UPDATE SET
  title             = EXCLUDED.title,
  type              = EXCLUDED.type,
  content           = EXCLUDED.content,
  estimated_minutes = EXCLUDED.estimated_minutes;

-- セクション 5: テストとデプロイ
INSERT INTO lessons (id, section_id, title, type, content, display_order, estimated_minutes, required_to_advance) VALUES
  (9, 5, 'PHPUnit でテストを書く', 'reading', $$
{
  "type": "reading",
  "markdown": "# PHPUnit でテストを書く\n\n## Feature Test\n\n```bash\nphp artisan make:test TaskCreateTest\n```\n\n`tests/Feature/TaskCreateTest.php`:\n\n```php\nuse Tests\\TestCase;\nuse App\\Models\\User;\nuse Illuminate\\Foundation\\Testing\\RefreshDatabase;\n\nclass TaskCreateTest extends TestCase {\n    use RefreshDatabase;\n\n    public function test_user_can_create_task(): void {\n        $user = User::factory()->create();\n        $response = $this->actingAs($user)->post('/tasks', [\n            'title' => '買い物に行く',\n        ]);\n        $response->assertRedirect('/tasks');\n        $this->assertDatabaseHas('tasks', [\n            'title' => '買い物に行く',\n            'user_id' => $user->id,\n        ]);\n    }\n}\n```\n\n## 実行\n\n```bash\nphp artisan test\n```\n\n## カバレッジ\n\n```bash\nphp artisan test --coverage --min=80\n```"
}
$$::jsonb, 1, 35, false),

  (10, 5, 'コース修了テスト', 'quiz', $$
{
  "type": "quiz",
  "questions": [
    {
      "id": "q1",
      "kind": "single_choice",
      "text": "Feature Test と Unit Test の違いは？",
      "options": [
        "Feature は HTTP / DB を含む E2E、Unit は単一クラス内",
        "Feature は速い、Unit は遅い",
        "違いはなく書き方の好み",
        "Feature は本番、Unit は開発"
      ],
      "correct_index": 0,
      "explanation": "Feature は actingAs / assertDatabaseHas など全層を通すテスト、Unit は依存をモックして単一ロジックをテスト"
    },
    {
      "id": "q2",
      "kind": "single_choice",
      "text": "RefreshDatabase trait の役割は？",
      "options": [
        "テスト前に毎回 DB を migrate し直す",
        "本番 DB をリセット",
        "Cache をクリアする",
        "ログをクリアする"
      ],
      "correct_index": 0,
      "explanation": "RefreshDatabase はテスト用 DB をクリーンな状態で migrate し、テスト間の状態漏れを防ぐ"
    },
    {
      "id": "q3",
      "kind": "true_false",
      "text": "User::factory()->create() はテスト用にダミーユーザーを DB に作る",
      "options": ["○", "×"],
      "correct_index": 0,
      "explanation": "Factory は seed 用のフェイクデータ生成。create で DB に永続化、make でメモリのみ"
    }
  ]
}
$$::jsonb, 2, 10, true)
ON CONFLICT (section_id, display_order) DO UPDATE SET
  title             = EXCLUDED.title,
  type              = EXCLUDED.type,
  content           = EXCLUDED.content,
  estimated_minutes = EXCLUDED.estimated_minutes;

SELECT setval(pg_get_serial_sequence('lessons','id'), (SELECT MAX(id) FROM lessons), true);

-- ----------------------------------------------------------------------------
-- 6. Trainee を Laravel コースに enroll
-- ----------------------------------------------------------------------------

INSERT INTO user_course_enrollments (user_id, course_id)
SELECT u.id, 1
  FROM users u
 WHERE u.email = 'fca2406070005@edu.fca.ac.jp'
ON CONFLICT (user_id, course_id) DO NOTHING;

COMMIT;

-- ----------------------------------------------------------------------------
-- 確認用クエリ (手動実行)
-- ----------------------------------------------------------------------------
-- SELECT id, email, role, company_id FROM users ORDER BY id;
-- SELECT id, title, status FROM courses;
-- SELECT s.id, s.title, COUNT(l.id) lessons
--   FROM sections s LEFT JOIN lessons l ON l.section_id = s.id
--  GROUP BY s.id, s.title ORDER BY s.id;
-- SELECT user_id, course_id, enrolled_at FROM user_course_enrollments;
