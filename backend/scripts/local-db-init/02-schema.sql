-- Add new schema named "public"
CREATE SCHEMA IF NOT EXISTS "public";
-- Set comment to schema: "public"
COMMENT ON SCHEMA "public" IS 'standard public schema';
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
-- Create "comment_threads" table
CREATE TABLE "public"."comment_threads" (
  "id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  "page_id" uuid NOT NULL,
  "block_id" uuid NULL,
  "anchor_from" integer NULL,
  "anchor_to" integer NULL,
  "quote" text NULL,
  "resolved_at" timestamptz NULL,
  "resolved_by_user_id" bigint NULL,
  "created_by_user_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "ck_comment_threads_anchor_pair" CHECK ((anchor_from IS NULL) = (anchor_to IS NULL)),
  CONSTRAINT "ck_comment_threads_resolved_pair" CHECK ((resolved_at IS NULL) = (resolved_by_user_id IS NULL))
);
-- Create index "idx_comment_threads_block" to table: "comment_threads"
CREATE INDEX "idx_comment_threads_block" ON "public"."comment_threads" ("block_id");
-- Create index "idx_comment_threads_page" to table: "comment_threads"
CREATE INDEX "idx_comment_threads_page" ON "public"."comment_threads" ("workspace_id", "page_id");
-- Create "comments" table
CREATE TABLE "public"."comments" (
  "id" uuid NOT NULL,
  "thread_id" uuid NOT NULL,
  "author_user_id" bigint NOT NULL,
  "body" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "ck_comments_body_array" CHECK (jsonb_typeof(body) = 'array'::text)
);
-- Create index "idx_comments_thread" to table: "comments"
CREATE INDEX "idx_comments_thread" ON "public"."comments" ("thread_id");
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
-- Create "page_links" table
CREATE TABLE "public"."page_links" (
  "source_block_id" uuid NOT NULL,
  "target_page_id" uuid NOT NULL,
  PRIMARY KEY ("source_block_id", "target_page_id")
);
-- Create index "idx_page_links_target_page_id" to table: "page_links"
CREATE INDEX "idx_page_links_target_page_id" ON "public"."page_links" ("target_page_id");
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
-- Create "page_search" table
CREATE TABLE "public"."page_search" (
  "page_id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  "title" text NOT NULL,
  "body" text NOT NULL,
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("page_id")
);
-- Create "page_snapshots" table
CREATE TABLE "public"."page_snapshots" (
  "page_id" uuid NOT NULL,
  "doc" jsonb NOT NULL,
  "built_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("page_id"),
  CONSTRAINT "ck_page_snapshots_doc" CHECK ((jsonb_typeof(doc) = 'object'::text) AND ((doc ->> 'type'::text) = 'doc'::text))
);
-- Create "page_suggestions" table
CREATE TABLE "public"."page_suggestions" (
  "id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  "page_id" uuid NOT NULL,
  "base_seq" bigint NULL,
  "doc" jsonb NOT NULL,
  "status" text NOT NULL DEFAULT 'open',
  "author_user_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "resolved_at" timestamptz NULL,
  "resolved_by_user_id" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ck_page_suggestions_doc" CHECK ((jsonb_typeof(doc) = 'object'::text) AND ((doc ->> 'type'::text) = 'doc'::text)),
  CONSTRAINT "ck_page_suggestions_resolution_consistency" CHECK (((status = 'open'::text) AND (resolved_at IS NULL) AND (resolved_by_user_id IS NULL)) OR ((status <> 'open'::text) AND (resolved_at IS NOT NULL) AND (resolved_by_user_id IS NOT NULL))),
  CONSTRAINT "ck_page_suggestions_status" CHECK (status = ANY (ARRAY['open'::text, 'accepted'::text, 'rejected'::text]))
);
-- Create index "idx_page_suggestions_open_base_seq" to table: "page_suggestions"
CREATE INDEX "idx_page_suggestions_open_base_seq" ON "public"."page_suggestions" ("page_id", "base_seq") WHERE (status = 'open'::text);
-- Create index "idx_page_suggestions_page" to table: "page_suggestions"
CREATE INDEX "idx_page_suggestions_page" ON "public"."page_suggestions" ("workspace_id", "page_id");
-- Create "page_templates" table
CREATE TABLE "public"."page_templates" (
  "id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  "space_id" uuid NULL,
  "name" text NOT NULL,
  "icon" jsonb NULL,
  "doc" jsonb NOT NULL,
  "created_by_user_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_page_templates_workspace_name" UNIQUE ("workspace_id", "name"),
  CONSTRAINT "ck_page_templates_doc" CHECK ((jsonb_typeof(doc) = 'object'::text) AND ((doc ->> 'type'::text) = 'doc'::text)),
  CONSTRAINT "ck_page_templates_icon" CHECK ((icon IS NULL) OR (jsonb_typeof(icon) = 'object'::text))
);
-- Create index "idx_page_templates_workspace_id" to table: "page_templates"
CREATE INDEX "idx_page_templates_workspace_id" ON "public"."page_templates" ("workspace_id");
-- Create "page_versions" table
CREATE TABLE "public"."page_versions" (
  "workspace_id" uuid NOT NULL,
  "page_id" uuid NOT NULL,
  "seq" bigint NOT NULL,
  "doc" jsonb NOT NULL,
  "author_user_id" bigint NOT NULL,
  "note" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("page_id", "seq"),
  CONSTRAINT "ck_page_versions_doc" CHECK ((jsonb_typeof(doc) = 'object'::text) AND ((doc ->> 'type'::text) = 'doc'::text))
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
  "icon" jsonb NULL,
  "cover" jsonb NULL,
  "last_edited_by_user_id" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uq_pages_workspace_id" UNIQUE ("workspace_id", "id"),
  CONSTRAINT "uq_pages_workspace_space_id" UNIQUE ("workspace_id", "space_id", "id"),
  CONSTRAINT "fk_pages_parent" FOREIGN KEY ("workspace_id", "space_id", "parent_id") REFERENCES "public"."pages" ("workspace_id", "space_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "ck_pages_cover_object" CHECK ((cover IS NULL) OR ((jsonb_typeof(cover) = 'object'::text) AND (cover <> '{}'::jsonb))),
  CONSTRAINT "ck_pages_icon_object" CHECK ((icon IS NULL) OR ((jsonb_typeof(icon) = 'object'::text) AND (icon <> '{}'::jsonb))),
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
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NULL,
  "workspace_id" uuid NULL,
  PRIMARY KEY ("id")
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
-- Modify "comment_threads" table
ALTER TABLE "public"."comment_threads" ADD CONSTRAINT "fk_comment_threads_block" FOREIGN KEY ("block_id") REFERENCES "public"."blocks" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "fk_comment_threads_page" FOREIGN KEY ("workspace_id", "page_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "comments" table
ALTER TABLE "public"."comments" ADD CONSTRAINT "fk_comments_thread" FOREIGN KEY ("thread_id") REFERENCES "public"."comment_threads" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "page_grants" table
ALTER TABLE "public"."page_grants" ADD CONSTRAINT "fk_page_grants_page" FOREIGN KEY ("workspace_id", "page_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_page_grants_principal" FOREIGN KEY ("workspace_id", "principal_id") REFERENCES "public"."principals" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "page_links" table
ALTER TABLE "public"."page_links" ADD CONSTRAINT "fk_page_links_source_block" FOREIGN KEY ("source_block_id") REFERENCES "public"."blocks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_page_links_target_page" FOREIGN KEY ("target_page_id") REFERENCES "public"."pages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "page_paths" table
ALTER TABLE "public"."page_paths" ADD CONSTRAINT "fk_page_paths_ancestor" FOREIGN KEY ("workspace_id", "ancestor_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_page_paths_page" FOREIGN KEY ("workspace_id", "page_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "page_search" table
ALTER TABLE "public"."page_search" ADD CONSTRAINT "fk_page_search_page" FOREIGN KEY ("workspace_id", "page_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "page_snapshots" table
ALTER TABLE "public"."page_snapshots" ADD CONSTRAINT "fk_page_snapshots_page" FOREIGN KEY ("page_id") REFERENCES "public"."pages" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "page_suggestions" table
ALTER TABLE "public"."page_suggestions" ADD CONSTRAINT "fk_page_suggestions_base_version" FOREIGN KEY ("page_id", "base_seq") REFERENCES "public"."page_versions" ("page_id", "seq") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "fk_page_suggestions_page" FOREIGN KEY ("workspace_id", "page_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "page_templates" table
ALTER TABLE "public"."page_templates" ADD CONSTRAINT "fk_page_templates_space" FOREIGN KEY ("workspace_id", "space_id") REFERENCES "public"."spaces" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_page_templates_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "page_versions" table
ALTER TABLE "public"."page_versions" ADD CONSTRAINT "fk_page_versions_page" FOREIGN KEY ("workspace_id", "page_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "pages" table
ALTER TABLE "public"."pages" ADD CONSTRAINT "fk_pages_space" FOREIGN KEY ("workspace_id", "space_id") REFERENCES "public"."spaces" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "principal_members" table
ALTER TABLE "public"."principal_members" ADD CONSTRAINT "fk_principal_members_group" FOREIGN KEY ("workspace_id", "group_kind", "group_principal_id") REFERENCES "public"."principals" ("workspace_id", "kind", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_principal_members_member" FOREIGN KEY ("workspace_id", "member_kind", "member_principal_id") REFERENCES "public"."principals" ("workspace_id", "kind", "id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "principals" table
ALTER TABLE "public"."principals" ADD CONSTRAINT "fk_principals_page" FOREIGN KEY ("workspace_id", "page_id") REFERENCES "public"."pages" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_principals_space" FOREIGN KEY ("workspace_id", "space_id") REFERENCES "public"."spaces" ("workspace_id", "id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_principals_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "fk_principals_workspace" FOREIGN KEY ("workspace_id") REFERENCES "public"."workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
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
