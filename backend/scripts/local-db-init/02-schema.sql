-- Add new schema named "public"
CREATE SCHEMA IF NOT EXISTS "public";
-- Set comment to schema: "public"
COMMENT ON SCHEMA "public" IS 'standard public schema';
-- Create "audit_events" table
CREATE TABLE "public"."audit_events" (
  "id" bigserial NOT NULL,
  "actor_id" bigint NOT NULL,
  "actor_email" character varying(255) NOT NULL DEFAULT '',
  "actor_role" character varying(32) NOT NULL DEFAULT '',
  "action" character varying(160) NOT NULL DEFAULT '',
  "target_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_audit_events_action" to table: "audit_events"
CREATE INDEX "idx_audit_events_action" ON "public"."audit_events" ("action");
-- Create index "idx_audit_events_actor_id" to table: "audit_events"
CREATE INDEX "idx_audit_events_actor_id" ON "public"."audit_events" ("actor_id");
-- Create index "idx_audit_events_created_at" to table: "audit_events"
CREATE INDEX "idx_audit_events_created_at" ON "public"."audit_events" ("created_at");
-- Create "blocks" table
CREATE TABLE "public"."blocks" (
  "id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  "page_id" uuid NOT NULL,
  "parent_id" uuid NULL,
  "position" text NOT NULL COLLATE "C",
  "type" character varying(32) NOT NULL,
  "attrs" jsonb NOT NULL DEFAULT '{}',
  "inline" jsonb NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_blocks_workspace_page_id" UNIQUE ("workspace_id", "page_id", "id"),
  CONSTRAINT "fk_blocks_parent" FOREIGN KEY ("workspace_id", "page_id", "parent_id") REFERENCES "public"."blocks" ("workspace_id", "page_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "ck_blocks_attrs_object" CHECK (jsonb_typeof(attrs) = 'object'::text),
  CONSTRAINT "ck_blocks_inline_array" CHECK ((inline IS NULL) OR (jsonb_typeof(inline) = 'array'::text)),
  CONSTRAINT "ck_blocks_parent_not_self" CHECK ((parent_id IS NULL) OR (parent_id <> id)),
  CONSTRAINT "ck_blocks_position_not_empty" CHECK ("position" <> ''::text)
);
-- Create index "idx_blocks_page_id" to table: "blocks"
CREATE INDEX "idx_blocks_page_id" ON "public"."blocks" ("page_id");
-- Create index "idx_blocks_parent_id" to table: "blocks"
CREATE INDEX "idx_blocks_parent_id" ON "public"."blocks" ("parent_id");
-- Create index "idx_blocks_workspace_id" to table: "blocks"
CREATE INDEX "idx_blocks_workspace_id" ON "public"."blocks" ("workspace_id");
-- Create index "uq_blocks_page_position" to table: "blocks"
CREATE UNIQUE INDEX "uq_blocks_page_position" ON "public"."blocks" ("page_id", "position") WHERE (parent_id IS NULL);
-- Create index "uq_blocks_parent_position" to table: "blocks"
CREATE UNIQUE INDEX "uq_blocks_parent_position" ON "public"."blocks" ("parent_id", "position");
-- Create "chapter_grants" table
CREATE TABLE "public"."chapter_grants" (
  "workspace_id" uuid NOT NULL,
  "chapter_id" bigint NOT NULL,
  "principal_id" uuid NOT NULL,
  "role" character varying(16) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("workspace_id", "chapter_id", "principal_id"),
  CONSTRAINT "ck_chapter_grants_role" CHECK ((role)::text = ANY (ARRAY[('admin'::character varying)::text, ('editor'::character varying)::text, ('commenter'::character varying)::text, ('viewer'::character varying)::text]))
);
-- Create index "idx_chapter_grants_principal" to table: "chapter_grants"
CREATE INDEX "idx_chapter_grants_principal" ON "public"."chapter_grants" ("workspace_id", "principal_id");
-- Create "course_chapters" table
CREATE TABLE "public"."course_chapters" (
  "id" bigserial NOT NULL,
  "course_id" bigint NOT NULL,
  "created_by_user_id" bigint NOT NULL,
  "title" text NOT NULL DEFAULT '',
  "doc" jsonb NULL,
  "revision" bigint NOT NULL DEFAULT 1,
  "schema_version" bigint NOT NULL DEFAULT 1,
  "sort_order" bigint NOT NULL DEFAULT 100,
  "is_published" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "workspace_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_course_chapters_workspace_id" UNIQUE ("workspace_id", "id")
);
-- Create index "idx_course_chapters_course_id" to table: "course_chapters"
CREATE INDEX "idx_course_chapters_course_id" ON "public"."course_chapters" ("course_id");
-- Create index "idx_course_chapters_workspace_id" to table: "course_chapters"
CREATE INDEX "idx_course_chapters_workspace_id" ON "public"."course_chapters" ("workspace_id");
-- Create "course_grants" table
CREATE TABLE "public"."course_grants" (
  "workspace_id" uuid NOT NULL,
  "course_id" bigint NOT NULL,
  "principal_id" uuid NOT NULL,
  "role" character varying(16) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("workspace_id", "course_id", "principal_id"),
  CONSTRAINT "ck_course_grants_role" CHECK ((role)::text = ANY (ARRAY[('admin'::character varying)::text, ('editor'::character varying)::text, ('commenter'::character varying)::text, ('viewer'::character varying)::text]))
);
-- Create index "idx_course_grants_principal" to table: "course_grants"
CREATE INDEX "idx_course_grants_principal" ON "public"."course_grants" ("workspace_id", "principal_id");
-- Create "courses" table
CREATE TABLE "public"."courses" (
  "id" bigserial NOT NULL,
  "created_by_user_id" bigint NOT NULL,
  "title" text NOT NULL DEFAULT '',
  "description" text NOT NULL DEFAULT '',
  "category" text NOT NULL DEFAULT '',
  "language" character varying(50) NOT NULL DEFAULT '',
  "sort_order" bigint NOT NULL DEFAULT 100,
  "is_published" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "workspace_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_courses_workspace_id" UNIQUE ("workspace_id", "id")
);
-- Create index "idx_courses_workspace_id" to table: "courses"
CREATE INDEX "idx_courses_workspace_id" ON "public"."courses" ("workspace_id");
-- Create "exercise_submissions" table
CREATE TABLE "public"."exercise_submissions" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "exercise_kind" character varying(16) NOT NULL,
  "exercise_id" bigint NOT NULL,
  "submitted_code" text NOT NULL,
  "stdout" text NULL,
  "stderr" text NULL,
  "exit_code" bigint NOT NULL DEFAULT 0,
  "is_correct" boolean NOT NULL DEFAULT false,
  "submitted_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_submissions_user_at" to table: "exercise_submissions"
CREATE INDEX "idx_submissions_user_at" ON "public"."exercise_submissions" ("user_id", "submitted_at" DESC);
-- Create "invitations" table
CREATE TABLE "public"."invitations" (
  "id" bigserial NOT NULL,
  "email" text NOT NULL DEFAULT '',
  "role" text NOT NULL DEFAULT '',
  "name" text NOT NULL DEFAULT '',
  "status" text NOT NULL DEFAULT '',
  "token" character varying(64) NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL,
  "workspace_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ck_invitations_status" CHECK (status = ANY (ARRAY['pending'::text, 'accepted'::text, 'canceled'::text]))
);
-- Create index "idx_invitations_token" to table: "invitations"
CREATE UNIQUE INDEX "idx_invitations_token" ON "public"."invitations" ("token");
-- Create index "idx_invitations_workspace_id" to table: "invitations"
CREATE INDEX "idx_invitations_workspace_id" ON "public"."invitations" ("workspace_id");
-- Create "master_exercise_examples" table
CREATE TABLE "public"."master_exercise_examples" (
  "id" bigserial NOT NULL,
  "exercise_id" bigint NOT NULL,
  "order_index" smallint NOT NULL,
  "input_text" text NOT NULL DEFAULT '',
  "expected_output" text NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_examples_exercise_order" to table: "master_exercise_examples"
CREATE UNIQUE INDEX "idx_examples_exercise_order" ON "public"."master_exercise_examples" ("exercise_id", "order_index");
-- Create "master_exercises" table
CREATE TABLE "public"."master_exercises" (
  "id" bigserial NOT NULL,
  "slug" character varying(64) NOT NULL,
  "language" character varying(32) NOT NULL,
  "sort_order" integer NOT NULL DEFAULT 0,
  "category" character varying(64) NOT NULL,
  "title" character varying(200) NOT NULL,
  "description" text NOT NULL,
  "starter_code" text NOT NULL,
  "hint_text" text NULL,
  "expected_output" text NULL,
  "mode" character varying(16) NOT NULL DEFAULT 'execute',
  "explanation" text NOT NULL DEFAULT '',
  "difficulty" smallint NOT NULL DEFAULT 1,
  "is_published" boolean NOT NULL DEFAULT true,
  "chapter_id" bigint NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_master_exercises_language" to table: "master_exercises"
CREATE INDEX "idx_master_exercises_language" ON "public"."master_exercises" ("language");
-- Create index "idx_master_exercises_slug" to table: "master_exercises"
CREATE UNIQUE INDEX "idx_master_exercises_slug" ON "public"."master_exercises" ("slug");
-- Create "notes" table
CREATE TABLE "public"."notes" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "title" text NOT NULL DEFAULT '',
  "content" text NOT NULL DEFAULT '',
  "is_public" boolean NOT NULL DEFAULT false,
  "is_pinned" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_notes_user_id" to table: "notes"
CREATE INDEX "idx_notes_user_id" ON "public"."notes" ("user_id");
-- Create "notifications" table
CREATE TABLE "public"."notifications" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "type" text NOT NULL DEFAULT '',
  "title" text NOT NULL DEFAULT '',
  "body" text NOT NULL DEFAULT '',
  "is_read" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_notifications_user_id" to table: "notifications"
CREATE INDEX "idx_notifications_user_id" ON "public"."notifications" ("user_id");
-- Create "page_grants" table
CREATE TABLE "public"."page_grants" (
  "workspace_id" uuid NOT NULL,
  "page_id" uuid NOT NULL,
  "principal_id" uuid NOT NULL,
  "role" character varying(16) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("workspace_id", "page_id", "principal_id"),
  CONSTRAINT "ck_page_grants_role" CHECK ((role)::text = ANY (ARRAY[('admin'::character varying)::text, ('editor'::character varying)::text, ('commenter'::character varying)::text, ('viewer'::character varying)::text]))
);
-- Create index "idx_page_grants_principal" to table: "page_grants"
CREATE INDEX "idx_page_grants_principal" ON "public"."page_grants" ("workspace_id", "principal_id");
-- Create "page_paths" table
CREATE TABLE "public"."page_paths" (
  "workspace_id" uuid NOT NULL,
  "page_id" uuid NOT NULL,
  "ancestor_id" uuid NOT NULL,
  "depth" integer NOT NULL,
  PRIMARY KEY ("page_id", "ancestor_id"),
  CONSTRAINT "ck_page_paths_depth" CHECK ((depth >= 0) AND ((depth = 0) = (page_id = ancestor_id)))
);
-- Create index "idx_page_paths_ancestor_id" to table: "page_paths"
CREATE INDEX "idx_page_paths_ancestor_id" ON "public"."page_paths" ("ancestor_id");
-- Create index "idx_page_paths_workspace_id" to table: "page_paths"
CREATE INDEX "idx_page_paths_workspace_id" ON "public"."page_paths" ("workspace_id");
-- Create "page_snapshots" table
CREATE TABLE "public"."page_snapshots" (
  "page_id" uuid NOT NULL,
  "doc" jsonb NOT NULL,
  "built_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("page_id"),
  CONSTRAINT "ck_page_snapshots_doc" CHECK ((jsonb_typeof(doc) = 'object'::text) AND ((doc ->> 'type'::text) = 'doc'::text))
);
-- Create "pages" table
CREATE TABLE "public"."pages" (
  "id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  "space_id" uuid NOT NULL,
  "parent_id" uuid NULL,
  "position" text NOT NULL COLLATE "C",
  "title" character varying(200) NOT NULL DEFAULT '',
  "created_by_user_id" bigint NOT NULL,
  "archived_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_pages_workspace_id" UNIQUE ("workspace_id", "id"),
  CONSTRAINT "uq_pages_workspace_space_id" UNIQUE ("workspace_id", "space_id", "id"),
  CONSTRAINT "fk_pages_parent" FOREIGN KEY ("workspace_id", "space_id", "parent_id") REFERENCES "public"."pages" ("workspace_id", "space_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "ck_pages_parent_not_self" CHECK ((parent_id IS NULL) OR (parent_id <> id)),
  CONSTRAINT "ck_pages_position_not_empty" CHECK ("position" <> ''::text)
);
-- Create index "idx_pages_archived_at" to table: "pages"
CREATE INDEX "idx_pages_archived_at" ON "public"."pages" ("archived_at");
-- Create index "idx_pages_parent_id" to table: "pages"
CREATE INDEX "idx_pages_parent_id" ON "public"."pages" ("parent_id");
-- Create index "idx_pages_space_id" to table: "pages"
CREATE INDEX "idx_pages_space_id" ON "public"."pages" ("space_id");
-- Create index "idx_pages_workspace_id" to table: "pages"
CREATE INDEX "idx_pages_workspace_id" ON "public"."pages" ("workspace_id");
-- Create index "uq_pages_parent_position" to table: "pages"
CREATE UNIQUE INDEX "uq_pages_parent_position" ON "public"."pages" ("parent_id", "position") WHERE (archived_at IS NULL);
-- Create index "uq_pages_space_position" to table: "pages"
CREATE UNIQUE INDEX "uq_pages_space_position" ON "public"."pages" ("space_id", "position") WHERE ((parent_id IS NULL) AND (archived_at IS NULL));
-- Create "principal_members" table
CREATE TABLE "public"."principal_members" (
  "workspace_id" uuid NOT NULL,
  "group_principal_id" uuid NOT NULL,
  "member_principal_id" uuid NOT NULL,
  "group_kind" character varying(16) NULL GENERATED ALWAYS AS ('group'::character varying) STORED,
  "member_kind" character varying(16) NULL GENERATED ALWAYS AS ('user'::character varying) STORED,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("group_principal_id", "member_principal_id")
);
-- Create index "idx_principal_members_member" to table: "principal_members"
CREATE INDEX "idx_principal_members_member" ON "public"."principal_members" ("workspace_id", "member_principal_id");
-- Create "principals" table
CREATE TABLE "public"."principals" (
  "id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  "kind" character varying(16) NOT NULL,
  "user_id" bigint NULL,
  "space_id" uuid NULL,
  "page_id" uuid NULL,
  "name" character varying(200) NOT NULL DEFAULT '',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_principals_workspace_id" UNIQUE ("workspace_id", "id"),
  CONSTRAINT "uq_principals_workspace_kind_id" UNIQUE ("workspace_id", "kind", "id"),
  CONSTRAINT "uq_principals_workspace_kind_page_id" UNIQUE ("workspace_id", "kind", "page_id", "id"),
  CONSTRAINT "ck_principals_kind" CHECK ((kind)::text = ANY (ARRAY[('user'::character varying)::text, ('group'::character varying)::text, ('space_all'::character varying)::text, ('share_link'::character varying)::text])),
  CONSTRAINT "ck_principals_name" CHECK (((kind)::text = 'group'::text) = ((name)::text <> (''::character varying)::text)),
  CONSTRAINT "ck_principals_page_id" CHECK (((kind)::text = 'share_link'::text) = (page_id IS NOT NULL)),
  CONSTRAINT "ck_principals_space_id" CHECK (((kind)::text = 'space_all'::text) = (space_id IS NOT NULL)),
  CONSTRAINT "ck_principals_user_id" CHECK (((kind)::text = 'user'::text) = (user_id IS NOT NULL))
);
-- Create index "idx_principals_page_id" to table: "principals"
CREATE INDEX "idx_principals_page_id" ON "public"."principals" ("page_id");
-- Create index "idx_principals_space_id" to table: "principals"
CREATE INDEX "idx_principals_space_id" ON "public"."principals" ("space_id");
-- Create index "idx_principals_user_id" to table: "principals"
CREATE INDEX "idx_principals_user_id" ON "public"."principals" ("user_id");
-- Create index "idx_principals_workspace_id" to table: "principals"
CREATE INDEX "idx_principals_workspace_id" ON "public"."principals" ("workspace_id");
-- Create index "uq_principals_group_name" to table: "principals"
CREATE UNIQUE INDEX "uq_principals_group_name" ON "public"."principals" ("workspace_id", "name") WHERE ((kind)::text = 'group'::text);
-- Create index "uq_principals_space_all" to table: "principals"
CREATE UNIQUE INDEX "uq_principals_space_all" ON "public"."principals" ("workspace_id", "space_id") WHERE ((kind)::text = 'space_all'::text);
-- Create index "uq_principals_workspace_user" to table: "principals"
CREATE UNIQUE INDEX "uq_principals_workspace_user" ON "public"."principals" ("workspace_id", "user_id") WHERE ((kind)::text = 'user'::text);
-- Create "profiles" table
CREATE TABLE "public"."profiles" (
  "user_id" bigserial NOT NULL,
  "bio" text NOT NULL DEFAULT '',
  "avatar_url" text NOT NULL DEFAULT '',
  "status_message" text NOT NULL DEFAULT '',
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("user_id")
);
-- Create "rich_documents" table
CREATE TABLE "public"."rich_documents" (
  "id" uuid NOT NULL,
  "owner_id" bigint NOT NULL,
  "kind" text NOT NULL,
  "title" text NOT NULL,
  "is_public" boolean NOT NULL DEFAULT false,
  "schema_version" bigint NOT NULL DEFAULT 1,
  "doc" jsonb NOT NULL,
  "revision" bigint NOT NULL DEFAULT 1,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "workspace_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ck_rich_documents_doc" CHECK ((jsonb_typeof(doc) = 'object'::text) AND ((doc ->> 'type'::text) = 'doc'::text)),
  CONSTRAINT "ck_rich_documents_title_len" CHECK (char_length(title) <= 200)
);
-- Create index "idx_rich_documents_owner_id" to table: "rich_documents"
CREATE INDEX "idx_rich_documents_owner_id" ON "public"."rich_documents" ("owner_id");
-- Create "score_cards" table
CREATE TABLE "public"."score_cards" (
  "id" bigserial NOT NULL,
  "user_id" bigint NULL,
  "session_id" bigint NULL,
  "overall_score" numeric NULL,
  "logical_score" numeric NULL,
  "consideration_score" numeric NULL,
  "summary_score" numeric NULL,
  "proposal_score" numeric NULL,
  "listening_score" numeric NULL,
  "feedback" text NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "share_links" table
CREATE TABLE "public"."share_links" (
  "id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  "page_id" uuid NOT NULL,
  "principal_id" uuid NOT NULL,
  "principal_kind" character varying(16) NULL GENERATED ALWAYS AS ('share_link'::character varying) STORED,
  "capability" character varying(8) NOT NULL,
  "token_hash" bytea NOT NULL,
  "password_hash" text NULL,
  "expires_at" timestamptz NULL,
  "revoked_at" timestamptz NULL,
  "created_by_user_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_share_links_principal" UNIQUE ("principal_id"),
  CONSTRAINT "uq_share_links_token_hash" UNIQUE ("token_hash"),
  CONSTRAINT "ck_share_links_capability" CHECK ((capability)::text = ANY (ARRAY[('view'::character varying)::text, ('edit'::character varying)::text])),
  CONSTRAINT "ck_share_links_password_hash" CHECK ((password_hash IS NULL) OR (password_hash <> ''::text)),
  CONSTRAINT "ck_share_links_token_hash_len" CHECK (octet_length(token_hash) = 32)
);
-- Create index "idx_share_links_created_by" to table: "share_links"
CREATE INDEX "idx_share_links_created_by" ON "public"."share_links" ("created_by_user_id");
-- Create index "idx_share_links_page" to table: "share_links"
CREATE INDEX "idx_share_links_page" ON "public"."share_links" ("workspace_id", "page_id");
-- Create "space_grants" table
CREATE TABLE "public"."space_grants" (
  "workspace_id" uuid NOT NULL,
  "space_id" uuid NOT NULL,
  "principal_id" uuid NOT NULL,
  "role" character varying(16) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("workspace_id", "space_id", "principal_id"),
  CONSTRAINT "ck_space_grants_role" CHECK ((role)::text = ANY (ARRAY[('admin'::character varying)::text, ('editor'::character varying)::text, ('commenter'::character varying)::text, ('viewer'::character varying)::text]))
);
-- Create index "idx_space_grants_principal" to table: "space_grants"
CREATE INDEX "idx_space_grants_principal" ON "public"."space_grants" ("workspace_id", "principal_id");
-- Create "spaces" table
CREATE TABLE "public"."spaces" (
  "id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  "key" character varying(64) NOT NULL,
  "name" character varying(200) NOT NULL,
  "visibility" character varying(16) NOT NULL DEFAULT 'workspace',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_spaces_workspace_id" UNIQUE ("workspace_id", "id"),
  CONSTRAINT "uq_spaces_workspace_key" UNIQUE ("workspace_id", "key"),
  CONSTRAINT "ck_spaces_key_len" CHECK ((char_length((key)::text) >= 1) AND (char_length((key)::text) <= 64)),
  CONSTRAINT "ck_spaces_visibility" CHECK ((visibility)::text = ANY (ARRAY[('workspace'::character varying)::text, ('private'::character varying)::text]))
);
-- Create index "idx_spaces_workspace_id" to table: "spaces"
CREATE INDEX "idx_spaces_workspace_id" ON "public"."spaces" ("workspace_id");
-- Create "user_chapter_progress" table
CREATE TABLE "public"."user_chapter_progress" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "chapter_id" bigint NOT NULL,
  "course_id" bigint NOT NULL,
  "completed_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_user_chapter_progress_course_id" to table: "user_chapter_progress"
CREATE INDEX "idx_user_chapter_progress_course_id" ON "public"."user_chapter_progress" ("course_id");
-- Create index "ux_user_chapter_progress" to table: "user_chapter_progress"
CREATE UNIQUE INDEX "ux_user_chapter_progress" ON "public"."user_chapter_progress" ("user_id", "chapter_id");
-- Create "user_chapter_views" table
CREATE TABLE "public"."user_chapter_views" (
  "user_id" bigint NOT NULL,
  "chapter_id" bigint NOT NULL,
  "course_id" bigint NOT NULL,
  "first_viewed_at" timestamptz NOT NULL,
  "last_viewed_at" timestamptz NOT NULL,
  "view_count" integer NOT NULL DEFAULT 1,
  PRIMARY KEY ("user_id", "chapter_id")
);
-- Create "user_daily_activities" table
CREATE TABLE "public"."user_daily_activities" (
  "user_id" bigint NOT NULL,
  "activity_date" date NOT NULL,
  "exercise_count" integer NOT NULL DEFAULT 0,
  "correct_count" integer NOT NULL DEFAULT 0,
  "chapter_count" integer NOT NULL DEFAULT 0,
  "note_count" integer NOT NULL DEFAULT 0,
  PRIMARY KEY ("user_id", "activity_date")
);
-- Create "user_oidc_identities" table
CREATE TABLE "public"."user_oidc_identities" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "provider" text NOT NULL DEFAULT 'cognito',
  "subject" text NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ck_user_oidc_identities_not_empty" CHECK ((provider <> ''::text) AND (subject <> ''::text))
);
-- Create index "uq_user_oidc_provider_subject" to table: "user_oidc_identities"
CREATE UNIQUE INDEX "uq_user_oidc_provider_subject" ON "public"."user_oidc_identities" ("provider", "subject");
-- Create index "uq_user_oidc_user_provider" to table: "user_oidc_identities"
CREATE UNIQUE INDEX "uq_user_oidc_user_provider" ON "public"."user_oidc_identities" ("user_id", "provider");
-- Create "users" table
CREATE TABLE "public"."users" (
  "id" bigserial NOT NULL,
  "email" text NOT NULL DEFAULT '',
  "password_hash" text NULL,
  "name" text NOT NULL DEFAULT '',
  "role" text NOT NULL DEFAULT 'trainee',
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "workspace_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ck_users_role" CHECK (role = ANY (ARRAY['super_admin'::text, 'company_admin'::text, 'trainee'::text]))
);
-- Create index "uq_users_email_active" to table: "users"
CREATE UNIQUE INDEX "uq_users_email_active" ON "public"."users" ((lower(btrim(email, '	
 '::text)))) WHERE ((deleted_at IS NULL) AND (btrim(email, '	
 '::text) <> ''::text));
-- Create "workspace_grants" table
CREATE TABLE "public"."workspace_grants" (
  "workspace_id" uuid NOT NULL,
  "principal_id" uuid NOT NULL,
  "role" character varying(16) NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("workspace_id", "principal_id"),
  CONSTRAINT "ck_workspace_grants_role" CHECK ((role)::text = ANY (ARRAY[('admin'::character varying)::text, ('editor'::character varying)::text, ('commenter'::character varying)::text, ('viewer'::character varying)::text]))
);
-- Create "workspaces" table
CREATE TABLE "public"."workspaces" (
  "id" uuid NOT NULL,
  "slug" character varying(64) NOT NULL,
  "name" character varying(200) NOT NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "personal_owner_user_id" bigint NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_workspaces_slug" UNIQUE ("slug"),
  CONSTRAINT "ck_workspaces_slug_len" CHECK ((char_length((slug)::text) >= 1) AND (char_length((slug)::text) <= 64))
);
-- Create index "uq_workspaces_personal_owner" to table: "workspaces"
CREATE UNIQUE INDEX "uq_workspaces_personal_owner" ON "public"."workspaces" ("personal_owner_user_id") WHERE (personal_owner_user_id IS NOT NULL);
-- Modify "blocks" table
ALTER TABLE "public"."blocks" ADD CONSTRAINT "fk_blocks_page" FOREIGN KEY ("workspace_id", "page_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "chapter_grants" table
ALTER TABLE "public"."chapter_grants" ADD CONSTRAINT "fk_chapter_grants_chapter" FOREIGN KEY ("workspace_id", "chapter_id") REFERENCES "public"."course_chapters" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_chapter_grants_principal" FOREIGN KEY ("workspace_id", "principal_id") REFERENCES "public"."principals" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "course_chapters" table
ALTER TABLE "public"."course_chapters" ADD CONSTRAINT "fk_course_chapters_course" FOREIGN KEY ("course_id") REFERENCES "public"."courses" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_course_chapters_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "course_grants" table
ALTER TABLE "public"."course_grants" ADD CONSTRAINT "fk_course_grants_course" FOREIGN KEY ("workspace_id", "course_id") REFERENCES "public"."courses" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_course_grants_principal" FOREIGN KEY ("workspace_id", "principal_id") REFERENCES "public"."principals" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "courses" table
ALTER TABLE "public"."courses" ADD CONSTRAINT "fk_courses_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "invitations" table
ALTER TABLE "public"."invitations" ADD CONSTRAINT "fk_invitations_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "page_grants" table
ALTER TABLE "public"."page_grants" ADD CONSTRAINT "fk_page_grants_page" FOREIGN KEY ("workspace_id", "page_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_page_grants_principal" FOREIGN KEY ("workspace_id", "principal_id") REFERENCES "public"."principals" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "page_paths" table
ALTER TABLE "public"."page_paths" ADD CONSTRAINT "fk_page_paths_ancestor" FOREIGN KEY ("workspace_id", "ancestor_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_page_paths_page" FOREIGN KEY ("workspace_id", "page_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "page_snapshots" table
ALTER TABLE "public"."page_snapshots" ADD CONSTRAINT "fk_page_snapshots_page" FOREIGN KEY ("page_id") REFERENCES "public"."pages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "pages" table
ALTER TABLE "public"."pages" ADD CONSTRAINT "fk_pages_space" FOREIGN KEY ("workspace_id", "space_id") REFERENCES "public"."spaces" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "principal_members" table
ALTER TABLE "public"."principal_members" ADD CONSTRAINT "fk_principal_members_group" FOREIGN KEY ("workspace_id", "group_kind", "group_principal_id") REFERENCES "public"."principals" ("workspace_id", "kind", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_principal_members_member" FOREIGN KEY ("workspace_id", "member_kind", "member_principal_id") REFERENCES "public"."principals" ("workspace_id", "kind", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "principals" table
ALTER TABLE "public"."principals" ADD CONSTRAINT "fk_principals_page" FOREIGN KEY ("workspace_id", "page_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_principals_space" FOREIGN KEY ("workspace_id", "space_id") REFERENCES "public"."spaces" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_principals_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_principals_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "rich_documents" table
ALTER TABLE "public"."rich_documents" ADD CONSTRAINT "fk_rich_documents_owner" FOREIGN KEY ("owner_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_rich_documents_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "share_links" table
ALTER TABLE "public"."share_links" ADD CONSTRAINT "fk_share_links_created_by" FOREIGN KEY ("created_by_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_share_links_page" FOREIGN KEY ("workspace_id", "page_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_share_links_principal" FOREIGN KEY ("workspace_id", "principal_kind", "page_id", "principal_id") REFERENCES "public"."principals" ("workspace_id", "kind", "page_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "space_grants" table
ALTER TABLE "public"."space_grants" ADD CONSTRAINT "fk_space_grants_principal" FOREIGN KEY ("workspace_id", "principal_id") REFERENCES "public"."principals" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_space_grants_space" FOREIGN KEY ("workspace_id", "space_id") REFERENCES "public"."spaces" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "spaces" table
ALTER TABLE "public"."spaces" ADD CONSTRAINT "fk_spaces_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "user_oidc_identities" table
ALTER TABLE "public"."user_oidc_identities" ADD CONSTRAINT "fk_user_oidc_identities_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "users" table
ALTER TABLE "public"."users" ADD CONSTRAINT "fk_users_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "workspace_grants" table
ALTER TABLE "public"."workspace_grants" ADD CONSTRAINT "fk_workspace_grants_principal" FOREIGN KEY ("workspace_id", "principal_id") REFERENCES "public"."principals" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "workspaces" table
ALTER TABLE "public"."workspaces" ADD CONSTRAINT "fk_workspaces_personal_owner" FOREIGN KEY ("personal_owner_user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
