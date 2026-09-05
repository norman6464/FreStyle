# FreStyle のスキーマ正本（Atlas 宣言的スキーマ）。
#
# このファイルが唯一の正本。schema.gen.sql はこのファイルから
# `make schema-gen`（backend/Makefile）が機械生成する、DO NOT EDIT な副産物で、
# CI が drift（手で直したまま make し忘れ）を検査する。
#
# schema.gen.sql の使いどころ:
#   - sqlc の型付け入力（backend/sqlc.yaml）
#   - go:embed して結合テスト用 DB / ローカルの docker-entrypoint-initdb.d へ適用
#     （どちらも「まだ何も無い空の DB」に対してだけ使う。CREATE 文のみで DO ブロックも
#     IF NOT EXISTS も持たないので、既存 DB へ直接流すと衝突する）
#
# 既存 DB（本番・書き換え済みのローカル DB）への適用は必ずこの schema.hcl を正本にした
# `make schema-apply TARGET=<DSN>` を使う。Atlas が実 DB の現在の姿を見て差分だけを計算するため、
# 何度実行しても収束する（宣言のみ・履歴ファイルは持たない）。
#
# コメントの置き場所: DB のメタデータとして残したい説明は `comment = "…"`（COMMENT ON になり、
# 差分の対象にもなる）。ファイルの都合・設計の背景など DB に残す必要のないものは、
# この `#` 行コメントのように書く。
schema "public" {
  comment = "standard public schema"
}

# =====================================================================
# 中核（users / courses / exercises …）
# =====================================================================

# 利用者。deleted_at は実際に NULL になり得る。workspace_id は未所属で NULL。
#
# アプリ全体のロール（かつての users.role）は撤去済み。権限は per-workspace の
# grant（workspace_grants / space_grants / page_grants / course_grants / chapter_grants、
# domain.GrantRole）だけで表現する。
#
# workspace_id → workspaces.id は、users と workspaces が互いを参照する真の循環依存
# （workspaces.personal_owner_user_id が users を参照する）。Atlas は FK をまとめて
# 末尾の ALTER で張るため、宣言側は普通の foreign_key ブロックのままで済む
# （かつては「workspaces の CREATE TABLE より後でなければ張れない」ために DO ブロックで
# 追加していた唯一の例外だった）。
table "users" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "email" {
    null    = false
    type    = text
    default = ""
  }
  column "password_hash" {
    null = true
    type = text
  }
  column "name" {
    null    = false
    type    = text
    default = ""
  }
  column "is_active" {
    null    = false
    type    = boolean
    default = true
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  # 所属の正本（company_id は撤去済み）。
  column "workspace_id" {
    null = true
    type = uuid
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_users_workspace" {
    columns     = [column.workspace_id]
    ref_columns = [table.workspaces.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  # アクティブ行（未論理削除）かつ正規形が非空に限った部分 UNIQUE。論理削除→同メール再招待と
  # 両立し、email claim の無い OIDC ユーザー（空文字）は対象外にする。
  #
  # 重複データが既にある DB へこの索引を宣言的に適用すると、Atlas は作成に失敗する
  # （かつては DO ブロックで重複を検知し、警告に留めて起動は落とさない実行時分岐を持っていた。
  # 宣言的スキーマでは「重複があれば黙って作らない」という分岐そのものを表現できないため、
  # 適用前に重複を解消しておくことが前提になる。2026-09 時点で本番の重複は解消済みで、
  # ローカル / CI は毎回まっさらな DB から始まるため、以後どの環境でも重複には当たらない）。
  index "uq_users_email_active" {
    unique  = true
    on {
      expr = "lower(btrim(email, '\t\n\u000b\u000c\r '::text))"
    }
    where = "((deleted_at IS NULL) AND (btrim(email, '\t\n\u000b\u000c\r '::text) <> ''::text))"
  }
}

# OIDC プロバイダ由来のユーザー識別子（Cognito の sub を users から分離）。
table "user_oidc_identities" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "provider" {
    null    = false
    type    = text
    default = "cognito"
  }
  column "subject" {
    null = false
    type = text
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_user_oidc_identities_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "uq_user_oidc_user_provider" {
    unique  = true
    columns = [column.user_id, column.provider]
  }
  index "uq_user_oidc_provider_subject" {
    unique  = true
    columns = [column.provider, column.subject]
  }
  check "ck_user_oidc_identities_not_empty" {
    expr = "(provider <> ''::text) AND (subject <> ''::text)"
  }
}

# ワークスペース: テナント境界。ノート（spaces 以下）と、業務データ
# （courses / course_chapters / rich_documents）がどちらもこの表を指す。
table "workspaces" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  # slug は URL に出る短い識別子。テナント内ではなくグローバルに一意。
  column "slug" {
    null = false
    type = character_varying(64)
  }
  column "name" {
    null = false
    type = character_varying(200)
  }
  column "is_active" {
    null    = false
    type    = boolean
    default = true
  }
  # 個人サインアップで自動作成した、その人専用のワークスペース。1 人 1 つ
  # （uq_workspaces_personal_owner）。作った人を物理削除しても中身は消さない
  # （持ち主のいない箱として残り、招かれた他のメンバーはそのまま使い続けられる）。
  column "personal_owner_user_id" {
    null = true
    type = bigint
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_workspaces_personal_owner" {
    columns     = [column.personal_owner_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }
  # 1 人につき個人ワークスペースは 1 つ。サインアップの再送・並行実行でも 2 つ目が作れない
  # （check-then-act をアプリに書かずに済む。ON CONFLICT の推論先にもなる）。
  index "uq_workspaces_personal_owner" {
    unique = true
    columns = [column.personal_owner_user_id]
    where   = "(personal_owner_user_id IS NOT NULL)"
  }
  unique "uq_workspaces_slug" {
    columns = [column.slug]
  }
  # URL に出る識別子は空文字禁止・長さ上限（アプリ側検証と二重の壁）。
  check "ck_workspaces_slug_len" {
    expr = "(char_length((slug)::text) >= 1) AND (char_length((slug)::text) <= 64)"
  }
}

# 学習メモ。
table "notes" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "title" {
    null    = false
    type    = text
    default = ""
  }
  column "content" {
    null    = false
    type    = text
    default = ""
  }
  column "is_public" {
    null    = false
    type    = boolean
    default = false
  }
  column "is_pinned" {
    null    = false
    type    = boolean
    default = false
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_notes_user_id" {
    columns = [column.user_id]
  }
}

# users とは別管理のプロフィール拡張（user_id が PK）。
table "profiles" {
  schema = schema.public
  column "user_id" {
    null = false
    type = bigserial
  }
  column "bio" {
    null    = false
    type    = text
    default = ""
  }
  column "avatar_url" {
    null    = false
    type    = text
    default = ""
  }
  column "status_message" {
    null    = false
    type    = text
    default = ""
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.user_id]
  }
}

# アプリ内通知。
table "notifications" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "type" {
    null    = false
    type    = text
    default = ""
  }
  column "title" {
    null    = false
    type    = text
    default = ""
  }
  column "body" {
    null    = false
    type    = text
    default = ""
  }
  column "is_read" {
    null    = false
    type    = boolean
    default = false
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_notifications_user_id" {
    columns = [column.user_id]
  }
}

# 運営が用意した練習問題マスタ。
# hint_text / expected_output は任意項目で nullable。chapter_id も NULL を取り得る。
table "master_exercises" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "slug" {
    null = false
    type = character_varying(64)
  }
  column "language" {
    null = false
    type = character_varying(32)
  }
  column "sort_order" {
    null    = false
    type    = integer
    default = 0
  }
  column "category" {
    null = false
    type = character_varying(64)
  }
  column "title" {
    null = false
    type = character_varying(200)
  }
  column "description" {
    null = false
    type = text
  }
  column "starter_code" {
    null = false
    type = text
  }
  column "hint_text" {
    null = true
    type = text
  }
  column "expected_output" {
    null = true
    type = text
  }
  column "mode" {
    null    = false
    type    = character_varying(16)
    default = "execute"
  }
  column "explanation" {
    null    = false
    type    = text
    default = ""
  }
  column "difficulty" {
    null    = false
    type    = smallint
    default = 1
  }
  column "is_published" {
    null    = false
    type    = boolean
    default = true
  }
  column "chapter_id" {
    null = true
    type = bigint
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_master_exercises_slug" {
    unique  = true
    columns = [column.slug]
  }
  index "idx_master_exercises_language" {
    columns = [column.language]
  }
}

# 演習の入力例 / 期待出力例。(exercise_id, order_index) は同一問題内で一意。
table "master_exercise_examples" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "exercise_id" {
    null = false
    type = bigint
  }
  column "order_index" {
    null = false
    type = smallint
  }
  column "input_text" {
    null    = false
    type    = text
    default = ""
  }
  column "expected_output" {
    null = false
    type = text
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_examples_exercise_order" {
    unique  = true
    columns = [column.exercise_id, column.order_index]
  }
}

# コード演習の提出履歴（append-only）。stdout / stderr は未取得のとき NULL。
table "exercise_submissions" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "exercise_kind" {
    null = false
    type = character_varying(16)
  }
  column "exercise_id" {
    null = false
    type = bigint
  }
  column "submitted_code" {
    null = false
    type = text
  }
  column "stdout" {
    null = true
    type = text
  }
  column "stderr" {
    null = true
    type = text
  }
  column "exit_code" {
    null    = false
    type    = bigint
    default = 0
  }
  column "is_correct" {
    null    = false
    type    = boolean
    default = false
  }
  column "submitted_at" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_submissions_user_at" {
    on {
      column = column.user_id
    }
    on {
      desc   = true
      column = column.submitted_at
    }
  }
}

# リッチテキスト文書（tiptap JSON を jsonb で保持）。id はアプリ採番の uuid。
table "rich_documents" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "owner_id" {
    null = false
    type = bigint
  }
  column "kind" {
    null = false
    type = text
  }
  column "title" {
    null = false
    type = text
  }
  column "is_public" {
    null    = false
    type    = boolean
    default = false
  }
  column "schema_version" {
    null    = false
    type    = bigint
    default = 1
  }
  column "doc" {
    null = false
    type = jsonb
  }
  column "revision" {
    null    = false
    type    = bigint
    default = 1
  }
  column "created_at" {
    null = false
    type = timestamptz
  }
  column "updated_at" {
    null = false
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  # 所属の正本（company_id は撤去済み）。未所属の文書もあるため nullable。
  column "workspace_id" {
    null = true
    type = uuid
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_rich_documents_owner" {
    columns     = [column.owner_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  foreign_key "fk_rich_documents_workspace" {
    columns     = [column.workspace_id]
    ref_columns = [table.workspaces.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_rich_documents_owner_id" {
    columns = [column.owner_id]
  }
  check "ck_rich_documents_doc" {
    expr = "(jsonb_typeof(doc) = 'object'::text) AND ((doc ->> 'type'::text) = 'doc'::text)"
  }
  check "ck_rich_documents_title_len" {
    expr = "char_length(title) <= 200"
  }
}

# 日次の学習活動サマリー。PK = (user_id, activity_date)。書き込み時に upsert (+= delta)。
table "user_daily_activities" {
  schema = schema.public
  column "user_id" {
    null = false
    type = bigint
  }
  column "activity_date" {
    null = false
    type = date
  }
  column "exercise_count" {
    null    = false
    type    = integer
    default = 0
  }
  column "correct_count" {
    null    = false
    type    = integer
    default = 0
  }
  column "chapter_count" {
    null    = false
    type    = integer
    default = 0
  }
  column "note_count" {
    null    = false
    type    = integer
    default = 0
  }
  primary_key {
    columns = [column.user_id, column.activity_date]
  }
}


# =====================================================================
# ノートの骨格（spaces / pages / blocks / page_paths / page_snapshots）
# =====================================================================
#
# 設計の柱は 2 つ:
#
#   (1) 境界越えを DB で塞ぐ。親子の FK は必ず「入れ物」の列を含む複合 FK にし、
#       別のテナント / スペース / ページの行を親にできないようにする。
#       木はそれぞれの入れ物の中で閉じる: ページの木はスペースの中、ブロックの木はページの中。
#       入れ物をまたぐ親子を許すと、入れ物を消したときに ON DELETE CASCADE が
#       別の入れ物に残るはずの行まで道連れにする。
#       そのために参照先へ (workspace_id, …, id) の複合 UNIQUE を張る。id 単独の PK では
#       FK の参照列に複数列を指定できないため、実データ上は冗長でも足場として要る。
#
#   (2) 並び順は分数インデックス（internal/pkg/fracindex）が採番する文字列キー。
#       同じ親の中で position が重複しないことを部分 UNIQUE で守り、既定値は置かない（採番はアプリ側）。

# スペース: ワークスペース内のページの束（部門・プロジェクト単位の入れ物）。
table "spaces" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "workspace_id" {
    null = false
    type = uuid
  }
  # key はワークスペース内で一意な短い識別子（例: "eng"）。
  column "key" {
    null = false
    type = character_varying(64)
  }
  column "name" {
    null = false
    type = character_varying(200)
  }
  # visibility はワークスペース既定の grant が届くか（'workspace'）・届かないか（'private'）。
  # 'private' のスペースにはスペース単位の付与（space_grants）だけが届く。
  # 「プライベートかどうか」を grant の構成から導出しないための明示の印（値の正本は
  # domain.SpaceVisibility）。実効権限の畳み方は変えず、事実の集め方がこの列でふるう。
  column "visibility" {
    null    = false
    type    = character_varying(16)
    default = "workspace"
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  # ワークスペースの物理削除で配下も消える（運用ではアーカイブを使う想定で、物理削除は例外的な操作）。
  foreign_key "fk_spaces_workspace" {
    columns     = [column.workspace_id]
    ref_columns = [table.workspaces.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "idx_spaces_workspace_id" {
    columns = [column.workspace_id]
  }
  unique "uq_spaces_workspace_key" {
    columns = [column.workspace_id, column.key]
  }
  # pages からの複合 FK の参照先。id の PK があるので実データ上は冗長だが、
  # 「テナント越えを FK で塞ぐ」ための足場として要る。
  unique "uq_spaces_workspace_id" {
    columns = [column.workspace_id, column.id]
  }
  check "ck_spaces_key_len" {
    expr = "(char_length((key)::text) >= 1) AND (char_length((key)::text) <= 64)"
  }
  check "ck_spaces_visibility" {
    expr = "(visibility)::text = ANY (ARRAY[('workspace'::character varying)::text, ('private'::character varying)::text])"
  }
}

# ページ: ノートの 1 ページ。parent_id の自己参照で木をなす（無限入れ子）。
table "pages" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "workspace_id" {
    null = false
    type = uuid
  }
  column "space_id" {
    null = false
    type = uuid
  }
  # parent_id が NULL ならスペース直下（ルート）。
  column "parent_id" {
    null = true
    type = uuid
  }
  # position のコレーションは "C"（バイト順）に固定する。
  # 分数インデックスは「文字列の辞書順 = 並び順」が前提で、Go 側はバイト比較で判断する。
  # DB の既定がロケール依存のコレーション（例: en_US.utf8）だと 'a' < 'B' のように並び、
  # ORDER BY position がアプリの認識とずれる。列の定義で最初から揃えておく。
  column "position" {
    null      = false
    type      = text
    collate = "C"
  }
  column "title" {
    null    = false
    type    = character_varying(200)
    default = ""
  }
  # 作成者（users.id）。users への FK は張らない（ノートの骨格に閉じるため）。
  column "created_by_user_id" {
    null = false
    type = bigint
  }
  # archived_at が NULL の行が現役。物理削除ではなくアーカイブで隠す運用のため、
  # position の一意性はアーカイブ済みを除外した部分 UNIQUE で守る（下の index）。
  column "archived_at" {
    null = true
    type = timestamptz
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  # ページは「同じワークスペースの space」にしか属せない。
  foreign_key "fk_pages_space" {
    columns     = [column.workspace_id, column.space_id]
    ref_columns = [table.spaces.column.workspace_id, table.spaces.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  # 親は「同じワークスペースの、同じスペースの」ページだけ。親の物理削除で子孫も消える。
  # ページの木はスペースの中で閉じる。スペースはページの入れ物であり、木がスペースをまたぐと
  # パンくず（祖先をたどると別スペースに出る）・サブツリー一括取得・スペース単位の権限が
  # すべて破綻するため、space_id まで一致を要求する。workspace だけの一致だと、スペース A の
  # ページがスペース B のページを親に持ててしまい、スペース B を消したときに fk_pages_space の
  # CASCADE で B のページが消え、続けてこちらの CASCADE がスペース A に残るはずの子ページまで
  # 道連れにする。
  #
  # parent_id は NULL 可（ルート）。複合 FK は既定の MATCH SIMPLE なので、参照列に 1 つでも
  # NULL があれば検査自体が行われない ＝ ルートページは素通りする。これは意図どおり:
  # ルートの workspace_id / space_id は fk_pages_space 側で必ず検査されるため、テナント越え・
  # スペース越えの抜け道にはならない。
  #
  # 副作用（意図した挙動）: ページを別スペースへ移すときは、子孫の space_id も同じ文で
  # 更新しないと FK 違反になる。木の一部だけがスペースをまたぐ「中途半端な移動」を DB が防ぐ。
  foreign_key "fk_pages_parent" {
    columns     = [column.workspace_id, column.space_id, column.parent_id]
    ref_columns = [table.pages.column.workspace_id, table.pages.column.space_id, table.pages.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "idx_pages_workspace_id" {
    columns = [column.workspace_id]
  }
  index "idx_pages_space_id" {
    columns = [column.space_id]
  }
  index "idx_pages_parent_id" {
    columns = [column.parent_id]
  }
  # アーカイブ済みの除外・アーカイブ一覧の取得に使う。
  index "idx_pages_archived_at" {
    columns = [column.archived_at]
  }
  # 同じ親の中で position が重複しないこと。アーカイブ済みを除外する
  # （アーカイブは「一覧から隠す」だけで行は残るため、現役の並びだけを守る）。
  index "uq_pages_parent_position" {
    unique  = true
    columns = [column.parent_id, column.position]
    where   = "(archived_at IS NULL)"
  }
  # ルート直下（parent_id IS NULL）は上の索引では守れない。UNIQUE 索引は NULL 同士を
  # 別物として扱うため、parent_id が NULL の行同士は何度でも同じ position を持ててしまう。
  # ルートの並びはスペース単位なので、スペースを軸にした部分 UNIQUE を別に張る。
  index "uq_pages_space_position" {
    unique  = true
    columns = [column.space_id, column.position]
    where   = "((parent_id IS NULL) AND (archived_at IS NULL))"
  }
  # blocks / page_paths からの複合 FK の参照先。space_id を持たないテーブルからページを
  # 参照するには (workspace_id, id) の形が要る（fk_blocks_page / fk_page_paths_page /
  # fk_page_paths_ancestor が使う）。
  unique "uq_pages_workspace_id" {
    columns = [column.workspace_id, column.id]
  }
  # 親ページの FK を「同じスペース」まで絞るための足場。
  unique "uq_pages_workspace_space_id" {
    columns = [column.workspace_id, column.space_id, column.id]
  }
  # 自分自身を親にできない（1 行で閉じた循環を作らせない。多段の循環はアプリ側で検出する）。
  check "ck_pages_parent_not_self" {
    expr = "(parent_id IS NULL) OR (parent_id <> id)"
  }
  # position は空文字だと順序として意味を持たない（fracindex は空文字を返さない）。
  check "ck_pages_position_not_empty" {
    expr = "position <> ''::text"
  }
}

# ブロック: ページ本文を構成する 1 行（段落・見出し・リスト項目・表のセル …）。
# 入れ子（リストや表）は parent_id の自己参照で表す。
table "blocks" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "workspace_id" {
    null = false
    type = uuid
  }
  column "page_id" {
    null = false
    type = uuid
  }
  # parent_id が NULL ならページ直下（トップレベル）。
  column "parent_id" {
    null = true
    type = uuid
  }
  # pages.position と同じ理由でバイト順に固定する。
  column "position" {
    null      = false
    type      = text
    collate = "C"
  }
  # ProseMirror（tiptap）のノード名。値は domain.BlockType が正。
  column "type" {
    null = false
    type = character_varying(32)
  }
  # ProseMirror の attrs（見出しの level、コードブロックの language など）。
  # 属性が無いノードでも空オブジェクト {} を入れる（NULL と {} の二通りを作らない）。
  column "attrs" {
    null    = false
    type    = jsonb
    default = sql("'{}'::jsonb")
  }
  # 葉ノードのインライン内容（text ノードとマークの配列）。
  # リストや表のような容器ノードは子をブロック行として持つため NULL にする。
  column "inline" {
    null = true
    type = jsonb
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  # ブロックは「同じワークスペースの page」にしか属せない。
  foreign_key "fk_blocks_page" {
    columns     = [column.workspace_id, column.page_id]
    ref_columns = [table.pages.column.workspace_id, table.pages.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  # 親は「同じワークスペースの、同じページの」ブロックだけ。ブロックの木は 1 ページの中で
  # 閉じるものなので、page_id まで一致を要求する。workspace だけを一致させると、ページ A の
  # ブロックをページ B のブロックの親にでき、ページ A を消したときに ON DELETE CASCADE が
  # ページ B の本文まで消してしまう。MATCH SIMPLE の扱いは pages と同じで、parent_id が NULL
  # （トップレベル）なら検査されない。その場合の workspace_id / page_id の正しさは
  # fk_blocks_page 側で担保される。
  foreign_key "fk_blocks_parent" {
    columns     = [column.workspace_id, column.page_id, column.parent_id]
    ref_columns = [table.blocks.column.workspace_id, table.blocks.column.page_id, table.blocks.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "idx_blocks_workspace_id" {
    columns = [column.workspace_id]
  }
  index "idx_blocks_page_id" {
    columns = [column.page_id]
  }
  index "idx_blocks_parent_id" {
    columns = [column.parent_id]
  }
  index "uq_blocks_parent_position" {
    unique  = true
    columns = [column.parent_id, column.position]
  }
  # ブロックも同じ理由で、ページ直下（parent_id IS NULL）はページを軸に守る。
  index "uq_blocks_page_position" {
    unique  = true
    columns = [column.page_id, column.position]
    where   = "(parent_id IS NULL)"
  }
  # 親ブロックの FK を「同じページ」まで絞るための足場。
  unique "uq_blocks_workspace_page_id" {
    columns = [column.workspace_id, column.page_id, column.id]
  }
  check "ck_blocks_parent_not_self" {
    expr = "(parent_id IS NULL) OR (parent_id <> id)"
  }
  check "ck_blocks_position_not_empty" {
    expr = "position <> ''::text"
  }
  # attrs は ProseMirror の attrs なので必ず object（属性が無いノードでも {}）。
  check "ck_blocks_attrs_object" {
    expr = "jsonb_typeof(attrs) = 'object'::text"
  }
  # inline は葉ノードの content 配列。容器ノードでは NULL にする。
  check "ck_blocks_inline_array" {
    expr = "(inline IS NULL) OR (jsonb_typeof(inline) = 'array'::text)"
  }
}

# page_paths: ページの祖先関係を平らに持つ派生テーブル（closure table）。
# 自分自身も depth=0 の行として持つ。pages.parent_id の連鎖だけでも木は表せるが、
# パンくず・サブツリー一括取得・移動時の循環検出を再帰クエリなしの 1 回の JOIN で済ませる
# ためにこの索引を別に持つ。正本は pages.parent_id 側。
table "page_paths" {
  schema = schema.public
  column "workspace_id" {
    null = false
    type = uuid
  }
  column "page_id" {
    null = false
    type = uuid
  }
  column "ancestor_id" {
    null = false
    type = uuid
  }
  # 祖先までの距離。自分自身が 0、親が 1。
  column "depth" {
    null = false
    type = integer
  }
  primary_key {
    columns = [column.page_id, column.ancestor_id]
  }
  # 1 行で「子孫」と「祖先」の 2 ページを組にするため、単独 FK を 2 本張るだけでは
  # 別ワークスペースの 2 ページを組にした行が作れてしまう（両方の FK を通ってしまう）。
  # 行自身の workspace_id を軸にした複合 FK にして、組になる 2 ページが同じワークスペースに
  # 属することを DB 側で保証する。ページが消えたら派生であるこの行も一緒に消す。
  #
  # FK で守るのは「組になる 2 ページが実在し、同じワークスペースに属すること」まで。
  # 「depth が実際の親子の距離と一致するか、祖先の連鎖に抜けや余りが無いか」は複数行にまたがる
  # 不変条件で、宣言的な制約（行ごとの CHECK / FK）では表せない。この表は pages.parent_id から
  # 導ける派生データなので、正本である pages 側の制約で木の形を守り、closure 全体の整合は
  # 行を書く側の責務とする。なお page_paths は常に FK の子側で、この表の行が壊れても他の行を
  # CASCADE で消すことはない（壊れ方が表示の乱れに閉じ、他のデータを失わない）。
  foreign_key "fk_page_paths_page" {
    columns     = [column.workspace_id, column.page_id]
    ref_columns = [table.pages.column.workspace_id, table.pages.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  foreign_key "fk_page_paths_ancestor" {
    columns     = [column.workspace_id, column.ancestor_id]
    ref_columns = [table.pages.column.workspace_id, table.pages.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "idx_page_paths_workspace_id" {
    columns = [column.workspace_id]
  }
  # 祖先からサブツリーを引く経路（PK は (page_id, ancestor_id) なので ancestor_id 単独では効かない）。
  index "idx_page_paths_ancestor_id" {
    columns = [column.ancestor_id]
  }
  # 1 行だけで判定できる不変条件: depth は祖先までの距離なので負にならず、
  # depth=0 の行は自分自身を指す行「だけ」（逆に自己参照の行は必ず depth=0）。
  # パンくずは ORDER BY depth で組み立てるため、ここが崩れると pages.parent_id（正本）は
  # 正しいのに表示だけが壊れ、原因を追いにくい形で顕在化する。
  check "ck_page_paths_depth" {
    expr = "(depth >= 0) AND ((depth = 0) = (page_id = ancestor_id))"
  }
}

# page_snapshots: ページのブロック行を組み直した ProseMirror ドキュメント（読み取り用のキャッシュ）。
# 表示のたびにブロック行を木に組み直すと 1 ページで数百行の取得と再帰的な組み立てが要るため、
# 編集のたびに 1 つの jsonb へ焼き直して読み出しを 1 行の取得に落とす。
# 正本はあくまで blocks 側で、この行は失っても blocks から再生成できる派生データ。
table "page_snapshots" {
  schema = schema.public
  column "page_id" {
    null = false
    type = uuid
  }
  # tiptap の getJSON() 相当（type='doc' の ProseMirror ドキュメント）。
  column "doc" {
    null = false
    type = jsonb
  }
  # 焼き直した時刻。ブロックの更新時刻より古ければ作り直す判断に使う。
  column "built_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.page_id]
  }
  foreign_key "fk_page_snapshots_page" {
    columns     = [column.page_id]
    ref_columns = [table.pages.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  # 壊れた snapshot は読み取りキャッシュとしてそのまま返り、エディタがページを開けなくなるため、
  # rich_documents.doc と同じ形で入口を塞ぐ。
  check "ck_page_snapshots_doc" {
    expr = "(jsonb_typeof(doc) = 'object'::text) AND ((doc ->> 'type'::text) = 'doc'::text)"
  }
}

# =====================================================================
# ノートの権限（principals / grants / share_links）
# =====================================================================
#
# 設計の柱（骨格の 2 つに加えて）:
#
#   (3) 主体（principal）を 1 つの表に集める。ユーザー・グループ・スペース全員・公開リンクは
#       「権限を与える相手」という点で同じなので、grant 側から見て 1 本の FK で済む。
#       主体ごとに表を分けると grant が主体の種類だけ列（または表）を持つことになり、
#       権限を解く SQL が主体の種類だけ分岐する。
#
#   (4) 種類（kind）によって使う列が変わるので、CHECK で「その kind のときだけ非 NULL」を強制する。
#       任意の key/value に逃がす（EAV）ことはしない。列は意味を持ったまま、
#       「いつ埋まるか」だけを制約で表す。
#
#   (5) 権限は付与（grants）だけで表し、打ち消す層は持たない。
#       入れ物の階層に合わせて 3 段（workspace_grants / space_grants / page_grants）を置き、
#       届いた中で最も強い役割を採る。下の段が上の段を弱めることはない。
#
#       全ページへ ACL を展開する方式は解決が 1 行の取得で済む代わりに、ページを 1 回動かす /
#       メンバーを 1 人足すだけで数万行を書き換える。ページ移動が日常の道具である以上、
#       書き込み側の代償が大きすぎる。付与はごく少数のページにしか付かない性質を使い、
#       行を持つのは付与された段だけにして、解決は page_paths（closure）を 1 回 JOIN するだけで
#       済ませる。

# principals: 権限を与える相手（主体）。
#
# **この表がワークスペース所属の唯一の表現**（「そのワークスペースに kind='user' の行がある」
# ＝ そのユーザーはメンバー）。workspace_memberships のようなメンバーシップ専用の表は
# 作らない・足さないこと。作ると「principal はあるがメンバーではない」「メンバーだが
# principal が無い」の 2 通りのずれが生まれ、どちらが正かを決められなくなる。
# 所属の追加 / 削除はこの表への 1 行の INSERT / DELETE で表す。
#
# 「未所属」は行が無いことで表す。専用の値（0 や空文字）は置かない。既存の users.company_id が
# NULL と 0 の 2 通りで未所属を表していて層をまたいで混在していた轍を踏まないため。
table "principals" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "workspace_id" {
    null = false
    type = uuid
  }
  # kind の値は domain.PrincipalKind が正（user / group / space_all / share_link）。
  column "kind" {
    null = false
    type = character_varying(16)
  }
  # user_id は kind='user' のときだけ埋まる（既存 users への参照）。
  column "user_id" {
    null = true
    type = bigint
  }
  # space_id は kind='space_all'（そのスペースの全員）のときだけ埋まる。
  column "space_id" {
    null = true
    type = uuid
  }
  # page_id は kind='share_link'（公開リンクの来訪者）のときだけ埋まる。そのリンクの対象ページ。
  # 主体を「それが意味を持つ入れ物」に必ず結び付けるためで、こうするとページを物理削除したときに
  # 主体もリンクも CASCADE で一緒に消える。逆向き（share_links → principals）の FK だけでは、
  # ページを消してもリンクの行だけが消えて主体が残り、誰も指さない行が溜まる。
  column "page_id" {
    null = true
    type = uuid
  }
  # name は kind='group' の表示名。ほかの kind は名前を持たない
  # （ユーザー名は users、スペース名は spaces が正本。ここへ写すと二重管理になる）。
  column "name" {
    null    = false
    type    = character_varying(200)
    default = ""
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_principals_workspace" {
    columns     = [column.workspace_id]
    ref_columns = [table.workspaces.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  # users への FK は張る。principals はノートとアプリのユーザーを結ぶ唯一の接点で、
  # ここが緩いと「消えたユーザーの principal に権限が残る」＝ 別人が同じ id を再取得したときに
  # 権限を引き継いでしまう。骨格側の pages.created_by_user_id が FK を持たないのは、
  # あちらが既存テーブルだからで、新しく作るこの表には最初から張れる。
  foreign_key "fk_principals_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  # スペース全員の主体は「同じワークスペースの space」にしか結び付かない。
  foreign_key "fk_principals_space" {
    columns     = [column.workspace_id, column.space_id]
    ref_columns = [table.spaces.column.workspace_id, table.spaces.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  # 公開リンクの主体は「同じワークスペースの page」にしか結び付かない。
  foreign_key "fk_principals_page" {
    columns     = [column.workspace_id, column.page_id]
    ref_columns = [table.pages.column.workspace_id, table.pages.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "idx_principals_workspace_id" {
    columns = [column.workspace_id]
  }
  index "idx_principals_user_id" {
    columns = [column.user_id]
  }
  index "idx_principals_space_id" {
    columns = [column.space_id]
  }
  index "idx_principals_page_id" {
    columns = [column.page_id]
  }
  # 1 ユーザー 1 ワークスペースにつき主体は 1 つ（重複メンバーを作らない）。
  index "uq_principals_workspace_user" {
    unique  = true
    columns = [column.workspace_id, column.user_id]
    where   = "((kind)::text = 'user'::text)"
  }
  # 1 スペースにつき「全員」の主体は 1 つ。
  index "uq_principals_space_all" {
    unique  = true
    columns = [column.workspace_id, column.space_id]
    where   = "((kind)::text = 'space_all'::text)"
  }
  # グループ名はワークスペース内で一意（同名グループが 2 つあると権限を張る先を人が選べない）。
  index "uq_principals_group_name" {
    unique  = true
    columns = [column.workspace_id, column.name]
    where   = "((kind)::text = 'group'::text)"
  }
  # grant / share_link からの複合 FK の参照先。id の PK があるので実データ上は
  # 冗長だが、「別ワークスペースの principal に権限を張れない」を FK で塞ぐ足場として要る。
  unique "uq_principals_workspace_id" {
    columns = [column.workspace_id, column.id]
  }
  # kind まで含めた足場。参照側が「この列は group の principal でなければならない」を
  # FK で言えるようにする（principal_members / share_links が使う）。
  unique "uq_principals_workspace_kind_id" {
    columns = [column.workspace_id, column.kind, column.id]
  }
  # share_links からの複合 FK の参照先。リンクが持つ page_id と、その主体が持つ page_id が
  # 必ず同じページを指すことを FK で言えるようにする（2 か所に同じ値を持つ以上、
  # 食い違わないことは制約で担保する）。
  unique "uq_principals_workspace_kind_page_id" {
    columns = [column.workspace_id, column.kind, column.page_id, column.id]
  }
  check "ck_principals_kind" {
    expr = "(kind)::text = ANY (ARRAY[('user'::character varying)::text, ('group'::character varying)::text, ('space_all'::character varying)::text, ('share_link'::character varying)::text])"
  }
  # 使う列は kind で決まる。「その kind のときだけ非 NULL」を等式で書き、
  # 片方向（NOT NULL なのに kind が違う）も同時に塞ぐ。
  check "ck_principals_user_id" {
    expr = "((kind)::text = 'user'::text) = (user_id IS NOT NULL)"
  }
  check "ck_principals_space_id" {
    expr = "((kind)::text = 'space_all'::text) = (space_id IS NOT NULL)"
  }
  check "ck_principals_page_id" {
    expr = "((kind)::text = 'share_link'::text) = (page_id IS NOT NULL)"
  }
  check "ck_principals_name" {
    expr = "((kind)::text = 'group'::text) = (name <> ''::character varying)"
  }
}

# principal_members: グループの所属（group principal ↔ member principal）。
#
# グループの入れ子は許さない。member 側を kind='user' に固定することで、
# 「あるユーザーの主体の集合」を再帰なしの 1 回の UNION で出せる
# （入れ子を許すと権限解決に再帰 CTE が要るうえ、グループ同士の循環を防ぐ手当ても要る）。
#
# kind の固定は生成列（GENERATED ALWAYS AS ... STORED）で行う。定数なので INSERT / UPDATE から
# 値を渡せず、書き手が間違えようがない。CHECK 付きの普通の列にすると「書けるが必ず同じ値」に
# なり、実質使われない列を 2 つ抱えることになる。ここは足場であって属性ではない。
#
# これが無いと、たとえば member_principal_id に他人のユーザー主体ではなくグループ主体を
# 入れることでグループを入れ子にでき、解決 SQL（1 段しか辿らない）が黙って権限を取りこぼす。
# 「取りこぼす」＝ 見えるはずのページが見えないだけなので、権限の穴ではないが原因を追いにくい。
table "principal_members" {
  schema = schema.public
  column "workspace_id" {
    null = false
    type = uuid
  }
  column "group_principal_id" {
    null = false
    type = uuid
  }
  column "member_principal_id" {
    null = false
    type = uuid
  }
  # FK の足場（定数の生成列）。テーブルの属性ではない。
  column "group_kind" {
    null = true
    type = character_varying(16)
    as {
      expr = "'group'::character varying"
      type = STORED
    }
  }
  column "member_kind" {
    null = true
    type = character_varying(16)
    as {
      expr = "'user'::character varying"
      type = STORED
    }
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.group_principal_id, column.member_principal_id]
  }
  foreign_key "fk_principal_members_group" {
    columns     = [column.workspace_id, column.group_kind, column.group_principal_id]
    ref_columns = [table.principals.column.workspace_id, table.principals.column.kind, table.principals.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  foreign_key "fk_principal_members_member" {
    columns     = [column.workspace_id, column.member_kind, column.member_principal_id]
    ref_columns = [table.principals.column.workspace_id, table.principals.column.kind, table.principals.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  # 「このユーザーが属するグループ」を引く経路（PK は group 側が先頭なので member 単独では効かない）。
  index "idx_principal_members_member" {
    columns = [column.workspace_id, column.member_principal_id]
  }
}

# workspace_grants: ワークスペース全体での既定の役割。配下の全スペースに効く。
#
# スペース単位の grant だけでは「テナント全体の管理者」を表すのにスペースの数だけ grant を
# 張って回ることになり、スペースが増えるたびに漏れる。入れ物の階層が workspace ⊃ space である
# 以上、既定も同じ 2 段で持つ。
#
# scope_type / scope_id を持つ 1 枚の汎用 grants 表にまとめる案は採らない。scope_id が
# workspaces と spaces の 2 つの表を指すことになり、FK で参照先の実在もテナントの一致も
# 守れなくなる。このスキーマは一貫して「境界を FK で塞ぐ」ことを優先しており、
# 表が 1 枚増える代わりに両方とも複合 FK で守れる形を選ぶ。
#
# workspaces への直接の FK は張らない。principals への複合 FK が (workspace_id, principal_id) で
# 実在する principal との一致を要求し、その principal 自身が workspaces へ FK を持つため、
# ワークスペースの実在も削除時の CASCADE も推移的に効く。
table "workspace_grants" {
  schema = schema.public
  column "workspace_id" {
    null = false
    type = uuid
  }
  column "principal_id" {
    null = false
    type = uuid
  }
  # "role" の値は domain.GrantRole が正（admin / editor / commenter / viewer）。
  column "role" {
    null = false
    type = character_varying(16)
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.workspace_id, column.principal_id]
  }
  foreign_key "fk_workspace_grants_principal" {
    columns     = [column.workspace_id, column.principal_id]
    ref_columns = [table.principals.column.workspace_id, table.principals.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  check "ck_workspace_grants_role" {
    expr = "(role)::text = ANY (ARRAY[('admin'::character varying)::text, ('editor'::character varying)::text, ('commenter'::character varying)::text, ('viewer'::character varying)::text])"
  }
}

# space_grants: そのスペースでの既定の役割。ページに例外が無いときは、これと workspace_grants の
# うち強い方が実効権限になる（domain.GrantRole.Rank 参照。弱い方を採る規則にすると
# スペースを 1 つ作って viewer を張るだけでワークスペース管理者を締め出せてしまう）。
#
# 1 つの principal がひとつのスペースで持つ役割は 1 つなので、代理キーを置かず自然キーを PK にする
# （代理キーを置くと「同じ principal に viewer と editor の 2 行」が作れてしまい、
# どちらが正かをアプリで決める羽目になる）。
table "space_grants" {
  schema = schema.public
  column "workspace_id" {
    null = false
    type = uuid
  }
  column "space_id" {
    null = false
    type = uuid
  }
  column "principal_id" {
    null = false
    type = uuid
  }
  column "role" {
    null = false
    type = character_varying(16)
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.workspace_id, column.space_id, column.principal_id]
  }
  foreign_key "fk_space_grants_space" {
    columns     = [column.workspace_id, column.space_id]
    ref_columns = [table.spaces.column.workspace_id, table.spaces.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  # workspace_id を含めることで「別ワークスペースの principal への grant」を DB が弾く。
  # principal_id 単独の FK だと、行の workspace_id と principal の workspace_id が
  # 食い違っていても両方の FK を通ってしまう（テナント越えの権限昇格になる）。
  foreign_key "fk_space_grants_principal" {
    columns     = [column.workspace_id, column.principal_id]
    ref_columns = [table.principals.column.workspace_id, table.principals.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  # 「この principal の grant」を引く経路（PK は入れ物側が先頭）。principal を消すときの
  # CASCADE 走査にも効く。
  index "idx_space_grants_principal" {
    columns = [column.workspace_id, column.principal_id]
  }
  check "ck_space_grants_role" {
    expr = "(role)::text = ANY (ARRAY[('admin'::character varying)::text, ('editor'::character varying)::text, ('commenter'::character varying)::text, ('viewer'::character varying)::text])"
  }
}

# page_grants: そのページ以下での既定の役割。workspace_grants / space_grants に続く 3 段目で、
# 意味も合成の仕方も上の 2 つと同じ（配下へ降りる・最も強いものを採る）。
#
# これが要るのは「この人にこのページだけ編集を渡す」を書くため。
#
# 経路は page_paths を辿る。祖先のページに editor を張れば、その子孫は既定が editor 以上に
# なる（親に渡したら配下も編集できる、という素直な形）。
#
# **弱める手段はこの層にも、どの層にも無い。** 権限は 3 段の付与を足し合わせて
# 「届いた中で最も強いもの」で決まり、下の段が上の段を打ち消すことはない。
# 「親は共有、この子だけ隠す」は書けない — 狭めたい内容は private のスペースへ置く。
table "page_grants" {
  schema = schema.public
  column "workspace_id" {
    null = false
    type = uuid
  }
  column "page_id" {
    null = false
    type = uuid
  }
  column "principal_id" {
    null = false
    type = uuid
  }
  column "role" {
    null = false
    type = character_varying(16)
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.workspace_id, column.page_id, column.principal_id]
  }
  foreign_key "fk_page_grants_page" {
    columns     = [column.workspace_id, column.page_id]
    ref_columns = [table.pages.column.workspace_id, table.pages.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  # space_grants と同じ理由で workspace_id を含む複合 FK にする（別テナントの principal へ
  # 付与できてしまうと、そのままテナント越えの権限昇格になる）。
  foreign_key "fk_page_grants_principal" {
    columns     = [column.workspace_id, column.principal_id]
    ref_columns = [table.principals.column.workspace_id, table.principals.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  # 経路をさかのぼって「祖先に張られた付与」を引く向きの索引。主キーは (workspace_id, page_id,
  # principal_id) なので page_id 先頭では principal から引けない。
  index "idx_page_grants_principal" {
    columns = [column.workspace_id, column.principal_id]
  }
  check "ck_page_grants_role" {
    expr = "(role)::text = ANY (ARRAY[('admin'::character varying)::text, ('editor'::character varying)::text, ('commenter'::character varying)::text, ('viewer'::character varying)::text])"
  }
}

# share_links: ログイン不要の公開 URL。
#
# 来訪者は kind='share_link' の principal として扱う。主体の種類を 1 本に揃えておくと、
# 権限解決の入口が主体ごとに分岐しない。
#
# ただし既定（そのリンクで何ができるか）は grants ではなくこの表の capability で決める。
# リンクの来訪者はワークスペースに所属しないので、付与の 3 段はそもそも届かない。
#
# **共有リンクは広げる方向にしか働かない。** ログインしていない相手へ「見せる」を足すだけで、
# すでに見えている人から取り上げることはない。
#
# token は平文で持たない。DB が漏れた時点で全リンクが開けるのを避けるため、SHA-256 の
# ダイジェストだけを保存して照合はハッシュ同士で行う（トークンは十分な長さの乱数なので
# 総当たりに強く、bcrypt のような遅いハッシュは要らない）。パスワードは人が選ぶ値なので
# 逆に総当たりに弱く、こちらは bcrypt で持つ。
table "share_links" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "workspace_id" {
    null = false
    type = uuid
  }
  # page_id はリンクの対象ページ。このページとその子孫が対象になる。
  column "page_id" {
    null = false
    type = uuid
  }
  column "principal_id" {
    null = false
    type = uuid
  }
  # FK の足場（定数の生成列）。principal_members と同じ理由でここも生成列にする。
  column "principal_kind" {
    null = true
    type = character_varying(16)
    as {
      expr = "'share_link'::character varying"
      type = STORED
    }
  }
  # capability の値は domain.Capability が正（view / edit）。
  column "capability" {
    null = false
    type = character_varying(8)
  }
  # token_hash は共有 URL に載るトークンの SHA-256（32 バイト固定）。
  column "token_hash" {
    null = false
    type = bytea
  }
  # password_hash は bcrypt。NULL ならパスワード無しで開ける。
  column "password_hash" {
    null = true
    type = text
  }
  # expires_at が NULL なら無期限。
  column "expires_at" {
    null = true
    type = timestamptz
  }
  # revoked_at が NULL なら有効。失効は行を消さず日付で残す（誰がいつ止めたかを追えるように）。
  column "revoked_at" {
    null = true
    type = timestamptz
  }
  column "created_by_user_id" {
    null = false
    type = bigint
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_share_links_page" {
    columns     = [column.workspace_id, column.page_id]
    ref_columns = [table.pages.column.workspace_id, table.pages.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  # principal は「同じワークスペースの、kind='share_link' の、同じページに結び付いた」主体だけ。
  # page_id まで参照列に含めることで、リンクと主体が別々のページを指す状態を作れなくする。
  foreign_key "fk_share_links_principal" {
    columns     = [column.workspace_id, column.principal_kind, column.page_id, column.principal_id]
    ref_columns = [table.principals.column.workspace_id, table.principals.column.kind, table.principals.column.page_id, table.principals.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  foreign_key "fk_share_links_created_by" {
    columns     = [column.created_by_user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
  index "idx_share_links_page" {
    columns = [column.workspace_id, column.page_id]
  }
  index "idx_share_links_created_by" {
    columns = [column.created_by_user_id]
  }
  # 1 つの share_link principal は 1 本のリンクだけを表す（使い回すと失効が効かなくなる）。
  unique "uq_share_links_principal" {
    columns = [column.principal_id]
  }
  # トークンからリンクを 1 件引く経路。UNIQUE はその索引も兼ねる。
  unique "uq_share_links_token_hash" {
    columns = [column.token_hash]
  }
  check "ck_share_links_capability" {
    expr = "(capability)::text = ANY (ARRAY[('view'::character varying)::text, ('edit'::character varying)::text])"
  }
  check "ck_share_links_password_hash" {
    expr = "(password_hash IS NULL) OR (password_hash <> ''::text)"
  }
  # SHA-256 以外（平文トークンをそのまま入れた等）を入口で弾く。
  check "ck_share_links_token_hash_len" {
    expr = "octet_length(token_hash) = 32"
  }
}
