-- =====================================================================
-- FreStyle 本番 DB スキーマ（PostgreSQL / Supabase）
-- =====================================================================
-- 用途: コードを読む前に DB 構造を把握するための「地図」。
--
-- 生成元 (ground truth = 実 DB):
--   supabase db dump --linked --schema public  (schema-only, データ無し)
--   project: frestyle-prod (ap-northeast-1 / Tokyo) / PostgreSQL 17.6
--   生成日: 2026-06-05
--   ※ Supabase 標準の GRANT / OWNER / ALTER DEFAULT PRIVILEGES / SET 系の
--      ノイズ行は可読性のため除去済（DDL のみ残す）。再生成時は上記コマンド。
--
-- 重要な前提:
--   - FK 制約は 1 つも無い。バックエンド(GORM)が FK を張らない方針のため、
--     テーブル間の関連はアプリ層で担保している（例: notes.user_id → users.id）。
--   - 文字列は基本 text、整数 PK / FK 相当は bigint、日時は timestamptz。
--   - AI チャットの本文メッセージは DynamoDB(fre_style_ai_chat)、画像/添付は S3。
--     この SQL には現れない（RDB に持つのはセッションメタ等のみ）。
--
-- テーブルの状態（コードが実際に使う “生きた” 表 = AutoMigrate 管理の 15 表）:
--   users / profiles / ai_chat_sessions / notes / learning_reports /
--   notifications / invitations / master_exercises / master_exercise_examples /
--   company_exercises / exercise_submissions / companies / courses /
--   teaching_materials / company_applications
--
-- ⚠️ レガシー残存（コードはもう触らない “死んだ” 表 = 要掃除の候補, 15 表）:
--   chat_rooms / chat_room_members / friendships / conversation_templates /
--   daily_goals / favorite_phrases / practice_scenarios / reminder_settings /
--   scenario_bookmarks / score_cards / score_goals / session_notes /
--   shared_sessions / weekly_challenges / weekly_challenge_progress
--   ※ これらは RDS→Supabase 移行(2026-05)で全テーブルを pg_dump/restore した
--      名残。実装側は PR-A〜PR-D で撤去済、AutoMigrate からも除外済のため、
--      残っていても再生成されず実害は無いが、いずれ DROP してよい。
-- =====================================================================


CREATE TABLE IF NOT EXISTS "public"."ai_chat_sessions" (
    "id" bigint NOT NULL,
    "user_id" bigint,
    "title" "text",
    "session_type" "text",
    "scenario_id" bigint,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."ai_chat_sessions_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."ai_chat_sessions_id_seq" OWNED BY "public"."ai_chat_sessions"."id";

CREATE TABLE IF NOT EXISTS "public"."chat_room_members" (
    "id" bigint NOT NULL,
    "room_id" bigint,
    "user_id" bigint,
    "joined_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."chat_room_members_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."chat_room_members_id_seq" OWNED BY "public"."chat_room_members"."id";

CREATE TABLE IF NOT EXISTS "public"."chat_rooms" (
    "id" bigint NOT NULL,
    "name" "text",
    "is_group" boolean,
    "created_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."chat_rooms_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."chat_rooms_id_seq" OWNED BY "public"."chat_rooms"."id";

CREATE TABLE IF NOT EXISTS "public"."companies" (
    "id" bigint NOT NULL,
    "name" "text" NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."companies_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."companies_id_seq" OWNED BY "public"."companies"."id";

CREATE TABLE IF NOT EXISTS "public"."company_applications" (
    "id" bigint NOT NULL,
    "company_name" character varying(200) NOT NULL,
    "applicant_name" character varying(120) NOT NULL,
    "email" character varying(255) NOT NULL,
    "message" "text",
    "status" character varying(16) DEFAULT 'pending'::character varying NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."company_applications_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."company_applications_id_seq" OWNED BY "public"."company_applications"."id";

CREATE TABLE IF NOT EXISTS "public"."company_exercises" (
    "id" bigint NOT NULL,
    "company_id" bigint NOT NULL,
    "language" character varying(32) NOT NULL,
    "title" character varying(200) NOT NULL,
    "description" "text" NOT NULL,
    "starter_code" "text" NOT NULL,
    "hint_text" "text",
    "expected_output" "text",
    "difficulty" smallint DEFAULT 1 NOT NULL,
    "is_published" boolean DEFAULT false NOT NULL,
    "chapter_id" bigint,
    "created_by" bigint NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."company_exercises_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."company_exercises_id_seq" OWNED BY "public"."company_exercises"."id";

CREATE TABLE IF NOT EXISTS "public"."conversation_templates" (
    "id" bigint NOT NULL,
    "title" "text",
    "body" "text",
    "category" "text",
    "is_active" boolean,
    "created_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."conversation_templates_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."conversation_templates_id_seq" OWNED BY "public"."conversation_templates"."id";

CREATE TABLE IF NOT EXISTS "public"."courses" (
    "id" bigint NOT NULL,
    "company_id" bigint NOT NULL,
    "created_by_user_id" bigint NOT NULL,
    "title" "text" DEFAULT ''::"text" NOT NULL,
    "description" "text" DEFAULT ''::"text" NOT NULL,
    "sort_order" bigint DEFAULT 100 NOT NULL,
    "is_published" boolean DEFAULT false NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."courses_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."courses_id_seq" OWNED BY "public"."courses"."id";

CREATE TABLE IF NOT EXISTS "public"."daily_goals" (
    "id" bigint NOT NULL,
    "user_id" bigint,
    "goal_date" "date",
    "target_minutes" bigint,
    "actual_minutes" bigint,
    "is_achieved" boolean,
    "created_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."daily_goals_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."daily_goals_id_seq" OWNED BY "public"."daily_goals"."id";

CREATE TABLE IF NOT EXISTS "public"."exercise_submissions" (
    "id" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "exercise_kind" character varying(16) NOT NULL,
    "exercise_id" bigint NOT NULL,
    "submitted_code" "text" NOT NULL,
    "stdout" "text",
    "stderr" "text",
    "exit_code" bigint DEFAULT 0 NOT NULL,
    "is_correct" boolean DEFAULT false NOT NULL,
    "submitted_at" timestamp with time zone NOT NULL
);

CREATE SEQUENCE IF NOT EXISTS "public"."exercise_submissions_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."exercise_submissions_id_seq" OWNED BY "public"."exercise_submissions"."id";

CREATE TABLE IF NOT EXISTS "public"."favorite_phrases" (
    "id" bigint NOT NULL,
    "user_id" bigint,
    "phrase" "text",
    "note" "text",
    "created_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."favorite_phrases_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."favorite_phrases_id_seq" OWNED BY "public"."favorite_phrases"."id";

CREATE TABLE IF NOT EXISTS "public"."friendships" (
    "id" bigint NOT NULL,
    "requester_id" bigint,
    "addressee_id" bigint,
    "status" "text",
    "created_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."friendships_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."friendships_id_seq" OWNED BY "public"."friendships"."id";

CREATE TABLE IF NOT EXISTS "public"."invitations" (
    "id" bigint NOT NULL,
    "company_id" bigint,
    "email" "text",
    "role" "text",
    "display_name" "text",
    "status" "text",
    "expires_at" timestamp with time zone,
    "created_at" timestamp with time zone,
    "token" character varying(64)
);

CREATE SEQUENCE IF NOT EXISTS "public"."invitations_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."invitations_id_seq" OWNED BY "public"."invitations"."id";

CREATE TABLE IF NOT EXISTS "public"."learning_reports" (
    "id" bigint NOT NULL,
    "user_id" bigint,
    "period_from" timestamp with time zone,
    "period_to" timestamp with time zone,
    "status" "text",
    "s3_key" "text",
    "created_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."learning_reports_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."learning_reports_id_seq" OWNED BY "public"."learning_reports"."id";

CREATE TABLE IF NOT EXISTS "public"."master_exercise_examples" (
    "id" bigint NOT NULL,
    "exercise_id" bigint NOT NULL,
    "order_index" smallint NOT NULL,
    "input_text" "text" DEFAULT ''::"text" NOT NULL,
    "expected_output" "text" NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."master_exercise_examples_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."master_exercise_examples_id_seq" OWNED BY "public"."master_exercise_examples"."id";

CREATE TABLE IF NOT EXISTS "public"."master_exercises" (
    "id" bigint NOT NULL,
    "order_index" bigint DEFAULT 0 NOT NULL,
    "category" character varying(64) NOT NULL,
    "title" character varying(200) NOT NULL,
    "description" "text" NOT NULL,
    "starter_code" "text" NOT NULL,
    "hint_text" "text",
    "expected_output" "text",
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "language" character varying(32) NOT NULL,
    "slug" character varying(64) NOT NULL,
    "difficulty" smallint DEFAULT 1 NOT NULL,
    "is_published" boolean DEFAULT true NOT NULL,
    "chapter_id" bigint,
    "mode" character varying(16) DEFAULT 'execute'::character varying NOT NULL,
    "explanation" "text" DEFAULT ''::"text" NOT NULL
);

CREATE SEQUENCE IF NOT EXISTS "public"."master_exercises_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."master_exercises_id_seq" OWNED BY "public"."master_exercises"."id";

CREATE TABLE IF NOT EXISTS "public"."notes" (
    "id" bigint NOT NULL,
    "user_id" bigint,
    "title" "text",
    "content" "text",
    "is_public" boolean,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "is_pinned" boolean DEFAULT false
);

CREATE SEQUENCE IF NOT EXISTS "public"."notes_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."notes_id_seq" OWNED BY "public"."notes"."id";

CREATE TABLE IF NOT EXISTS "public"."notifications" (
    "id" bigint NOT NULL,
    "user_id" bigint,
    "type" "text",
    "title" "text",
    "body" "text",
    "is_read" boolean,
    "created_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."notifications_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."notifications_id_seq" OWNED BY "public"."notifications"."id";

CREATE TABLE IF NOT EXISTS "public"."practice_scenarios" (
    "id" bigint NOT NULL,
    "title" "text",
    "description" "text",
    "category" "text",
    "difficulty_level" bigint,
    "system_prompt" "text",
    "is_active" boolean DEFAULT true,
    "created_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."practice_scenarios_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."practice_scenarios_id_seq" OWNED BY "public"."practice_scenarios"."id";

CREATE TABLE IF NOT EXISTS "public"."profiles" (
    "user_id" bigint NOT NULL,
    "bio" "text",
    "avatar_url" "text",
    "updated_at" timestamp with time zone,
    "status" "text" DEFAULT ''::"text"
);

CREATE SEQUENCE IF NOT EXISTS "public"."profiles_user_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."profiles_user_id_seq" OWNED BY "public"."profiles"."user_id";

CREATE TABLE IF NOT EXISTS "public"."reminder_settings" (
    "user_id" bigint NOT NULL,
    "hour_local" bigint,
    "minute_local" bigint,
    "weekdays_mask" bigint,
    "is_active" boolean,
    "updated_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."reminder_settings_user_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."reminder_settings_user_id_seq" OWNED BY "public"."reminder_settings"."user_id";

CREATE TABLE IF NOT EXISTS "public"."scenario_bookmarks" (
    "id" bigint NOT NULL,
    "user_id" bigint,
    "scenario_id" bigint,
    "created_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."scenario_bookmarks_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."scenario_bookmarks_id_seq" OWNED BY "public"."scenario_bookmarks"."id";

CREATE TABLE IF NOT EXISTS "public"."score_cards" (
    "id" bigint NOT NULL,
    "user_id" bigint,
    "session_id" bigint,
    "overall_score" numeric,
    "logical_score" numeric,
    "consideration_score" numeric,
    "summary_score" numeric,
    "proposal_score" numeric,
    "listening_score" numeric,
    "feedback" "text",
    "created_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."score_cards_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."score_cards_id_seq" OWNED BY "public"."score_cards"."id";

CREATE TABLE IF NOT EXISTS "public"."score_goals" (
    "user_id" bigint NOT NULL,
    "target_score" numeric,
    "updated_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."score_goals_user_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."score_goals_user_id_seq" OWNED BY "public"."score_goals"."user_id";

CREATE TABLE IF NOT EXISTS "public"."session_notes" (
    "id" bigint NOT NULL,
    "session_id" bigint,
    "user_id" bigint,
    "content" "text",
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."session_notes_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."session_notes_id_seq" OWNED BY "public"."session_notes"."id";

CREATE TABLE IF NOT EXISTS "public"."shared_sessions" (
    "id" bigint NOT NULL,
    "owner_id" bigint,
    "session_id" bigint,
    "title" "text",
    "description" "text",
    "is_public" boolean,
    "created_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."shared_sessions_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."shared_sessions_id_seq" OWNED BY "public"."shared_sessions"."id";

CREATE TABLE IF NOT EXISTS "public"."teaching_materials" (
    "id" bigint NOT NULL,
    "company_id" bigint NOT NULL,
    "created_by_user_id" bigint NOT NULL,
    "title" "text" DEFAULT ''::"text" NOT NULL,
    "content" "text" DEFAULT ''::"text" NOT NULL,
    "is_published" boolean DEFAULT false NOT NULL,
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "course_id" bigint,
    "order_in_course" bigint DEFAULT 100 NOT NULL
);

CREATE SEQUENCE IF NOT EXISTS "public"."teaching_materials_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."teaching_materials_id_seq" OWNED BY "public"."teaching_materials"."id";

CREATE TABLE IF NOT EXISTS "public"."users" (
    "id" bigint NOT NULL,
    "cognito_sub" "text",
    "email" "text",
    "display_name" "text",
    "company_id" bigint,
    "role" "text",
    "created_at" timestamp with time zone,
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "onboarded_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."users_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."users_id_seq" OWNED BY "public"."users"."id";

CREATE TABLE IF NOT EXISTS "public"."weekly_challenge_progress" (
    "user_id" bigint NOT NULL,
    "challenge_id" bigint NOT NULL,
    "completed" boolean,
    "updated_at" timestamp with time zone
);

CREATE TABLE IF NOT EXISTS "public"."weekly_challenges" (
    "id" bigint NOT NULL,
    "week_start" "date",
    "title" "text",
    "description" "text",
    "is_active" boolean,
    "created_at" timestamp with time zone
);

CREATE SEQUENCE IF NOT EXISTS "public"."weekly_challenges_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE "public"."weekly_challenges_id_seq" OWNED BY "public"."weekly_challenges"."id";

ALTER TABLE ONLY "public"."ai_chat_sessions" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."ai_chat_sessions_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."chat_room_members" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."chat_room_members_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."chat_rooms" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."chat_rooms_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."companies" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."companies_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."company_applications" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."company_applications_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."company_exercises" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."company_exercises_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."conversation_templates" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."conversation_templates_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."courses" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."courses_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."daily_goals" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."daily_goals_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."exercise_submissions" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."exercise_submissions_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."favorite_phrases" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."favorite_phrases_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."friendships" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."friendships_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."invitations" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."invitations_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."learning_reports" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."learning_reports_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."master_exercise_examples" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."master_exercise_examples_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."master_exercises" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."master_exercises_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."notes" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."notes_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."notifications" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."notifications_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."practice_scenarios" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."practice_scenarios_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."profiles" ALTER COLUMN "user_id" SET DEFAULT "nextval"('"public"."profiles_user_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."reminder_settings" ALTER COLUMN "user_id" SET DEFAULT "nextval"('"public"."reminder_settings_user_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."scenario_bookmarks" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."scenario_bookmarks_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."score_cards" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."score_cards_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."score_goals" ALTER COLUMN "user_id" SET DEFAULT "nextval"('"public"."score_goals_user_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."session_notes" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."session_notes_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."shared_sessions" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."shared_sessions_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."teaching_materials" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."teaching_materials_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."users" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."users_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."weekly_challenges" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."weekly_challenges_id_seq"'::"regclass");

ALTER TABLE ONLY "public"."ai_chat_sessions"
    ADD CONSTRAINT "ai_chat_sessions_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."chat_room_members"
    ADD CONSTRAINT "chat_room_members_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."chat_rooms"
    ADD CONSTRAINT "chat_rooms_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."companies"
    ADD CONSTRAINT "companies_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."company_applications"
    ADD CONSTRAINT "company_applications_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."company_exercises"
    ADD CONSTRAINT "company_exercises_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."conversation_templates"
    ADD CONSTRAINT "conversation_templates_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."courses"
    ADD CONSTRAINT "courses_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."daily_goals"
    ADD CONSTRAINT "daily_goals_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."exercise_submissions"
    ADD CONSTRAINT "exercise_submissions_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."favorite_phrases"
    ADD CONSTRAINT "favorite_phrases_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."friendships"
    ADD CONSTRAINT "friendships_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."invitations"
    ADD CONSTRAINT "invitations_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."learning_reports"
    ADD CONSTRAINT "learning_reports_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."master_exercise_examples"
    ADD CONSTRAINT "master_exercise_examples_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."notes"
    ADD CONSTRAINT "notes_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."notifications"
    ADD CONSTRAINT "notifications_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."master_exercises"
    ADD CONSTRAINT "php_exercises_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."practice_scenarios"
    ADD CONSTRAINT "practice_scenarios_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."profiles"
    ADD CONSTRAINT "profiles_pkey" PRIMARY KEY ("user_id");

ALTER TABLE ONLY "public"."reminder_settings"
    ADD CONSTRAINT "reminder_settings_pkey" PRIMARY KEY ("user_id");

ALTER TABLE ONLY "public"."scenario_bookmarks"
    ADD CONSTRAINT "scenario_bookmarks_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."score_cards"
    ADD CONSTRAINT "score_cards_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."score_goals"
    ADD CONSTRAINT "score_goals_pkey" PRIMARY KEY ("user_id");

ALTER TABLE ONLY "public"."session_notes"
    ADD CONSTRAINT "session_notes_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."shared_sessions"
    ADD CONSTRAINT "shared_sessions_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."teaching_materials"
    ADD CONSTRAINT "teaching_materials_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."users"
    ADD CONSTRAINT "users_pkey" PRIMARY KEY ("id");

ALTER TABLE ONLY "public"."weekly_challenge_progress"
    ADD CONSTRAINT "weekly_challenge_progress_pkey" PRIMARY KEY ("user_id", "challenge_id");

ALTER TABLE ONLY "public"."weekly_challenges"
    ADD CONSTRAINT "weekly_challenges_pkey" PRIMARY KEY ("id");

CREATE INDEX "idx_ai_chat_sessions_user_id" ON "public"."ai_chat_sessions" USING "btree" ("user_id");

CREATE INDEX "idx_chat_room_members_room_id" ON "public"."chat_room_members" USING "btree" ("room_id");

CREATE INDEX "idx_chat_room_members_user_id" ON "public"."chat_room_members" USING "btree" ("user_id");

CREATE INDEX "idx_company_applications_email" ON "public"."company_applications" USING "btree" ("email");

CREATE INDEX "idx_company_applications_status" ON "public"."company_applications" USING "btree" ("status");

CREATE INDEX "idx_company_exercises_company_id" ON "public"."company_exercises" USING "btree" ("company_id");

CREATE INDEX "idx_company_exercises_deleted_at" ON "public"."company_exercises" USING "btree" ("deleted_at");

CREATE INDEX "idx_company_exercises_language" ON "public"."company_exercises" USING "btree" ("language");

CREATE INDEX "idx_courses_company_id" ON "public"."courses" USING "btree" ("company_id");

CREATE INDEX "idx_daily_goals_user_id" ON "public"."daily_goals" USING "btree" ("user_id");

CREATE UNIQUE INDEX "idx_examples_exercise_order" ON "public"."master_exercise_examples" USING "btree" ("exercise_id", "order_index");

CREATE INDEX "idx_favorite_phrases_user_id" ON "public"."favorite_phrases" USING "btree" ("user_id");

CREATE INDEX "idx_friendships_addressee_id" ON "public"."friendships" USING "btree" ("addressee_id");

CREATE INDEX "idx_friendships_requester_id" ON "public"."friendships" USING "btree" ("requester_id");

CREATE INDEX "idx_invitations_company_id" ON "public"."invitations" USING "btree" ("company_id");

CREATE UNIQUE INDEX "idx_invitations_token" ON "public"."invitations" USING "btree" ("token");

CREATE INDEX "idx_learning_reports_user_id" ON "public"."learning_reports" USING "btree" ("user_id");

CREATE INDEX "idx_master_exercises_language" ON "public"."master_exercises" USING "btree" ("language");

CREATE UNIQUE INDEX "idx_master_exercises_slug" ON "public"."master_exercises" USING "btree" ("slug");

CREATE INDEX "idx_notes_user_id" ON "public"."notes" USING "btree" ("user_id");

CREATE INDEX "idx_notifications_user_id" ON "public"."notifications" USING "btree" ("user_id");

CREATE INDEX "idx_scenario_bookmarks_scenario_id" ON "public"."scenario_bookmarks" USING "btree" ("scenario_id");

CREATE INDEX "idx_scenario_bookmarks_user_id" ON "public"."scenario_bookmarks" USING "btree" ("user_id");

CREATE INDEX "idx_score_cards_session_id" ON "public"."score_cards" USING "btree" ("session_id");

CREATE INDEX "idx_score_cards_user_id" ON "public"."score_cards" USING "btree" ("user_id");

CREATE INDEX "idx_session_notes_session_id" ON "public"."session_notes" USING "btree" ("session_id");

CREATE INDEX "idx_session_notes_user_id" ON "public"."session_notes" USING "btree" ("user_id");

CREATE INDEX "idx_shared_sessions_owner_id" ON "public"."shared_sessions" USING "btree" ("owner_id");

CREATE INDEX "idx_shared_sessions_session_id" ON "public"."shared_sessions" USING "btree" ("session_id");

CREATE INDEX "idx_submissions_user_at" ON "public"."exercise_submissions" USING "btree" ("user_id", "submitted_at" DESC);

CREATE INDEX "idx_teaching_materials_company_id" ON "public"."teaching_materials" USING "btree" ("company_id");

CREATE INDEX "idx_teaching_materials_course_id" ON "public"."teaching_materials" USING "btree" ("course_id");

CREATE UNIQUE INDEX "idx_users_cognito_sub" ON "public"."users" USING "btree" ("cognito_sub");

CREATE UNIQUE INDEX "uniq_master_exercises_slug" ON "public"."master_exercises" USING "btree" ("slug");

