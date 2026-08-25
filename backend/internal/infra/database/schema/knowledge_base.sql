-- ナレッジ基盤（workspaces / spaces / pages / blocks / page_paths / page_snapshots）の DDL。
--
-- このファイルが実スキーマの正本であり、同時に sqlc の型付け入力でもある
-- （backend/sqlc.yaml の schema に登録済み。列を足したら `make sqlc` で生成物を作り直す）。
-- 起動時に database.ApplyKnowledgeBaseSchema がこの内容をそのまま流す（冪等）。
--
-- 既存テーブル（users / notes / courses / rich_documents …）とは扱いが違う。
-- あちらは GORM の AutoMigrate が実体を作り、queries/schema.sql はそれを写した型付け専用の定義でしかない。
-- ナレッジ基盤は GORM を一切通さない: domain 構造体に GORM タグを持たず、AutoMigrate の一覧にも入れない。
-- 複合 FK / CHECK / 部分 UNIQUE / コレーション指定は AutoMigrate では表現できず、
-- 「タグに書いた定義」と「明示 SQL に書いた定義」へ二重化して食い違うのが分かっているため、
-- 最初からこの 1 枚に集約して正本を 1 つにする。
--
-- 冪等性は CREATE ... IF NOT EXISTS だけで成り立たせ、DO ブロックは書かない。
-- sqlc がこのファイルをパースして型を作るので、素の DDL に保つ必要がある
-- （手続き型の DO ブロックが混ざると sqlc がパースできない）。
--
-- 注意（開発者のローカル DB）: CREATE TABLE IF NOT EXISTS は「テーブルが無いときだけ作る」ので、
-- 既に別定義のテーブルがある DB では何もしない。過去のコミット（AutoMigrate 版）でこれらの
-- テーブルを作ったローカル DB には古い定義が残る。このテーブル群は未リリースで本番にはまだ存在せず、
-- ローカル / 結合テストの DB は使い捨てにできるため、`docker compose down -v` で作り直す運用とする。
--
-- 設計の柱は 2 つ:
--
--   (1) 境界越えを DB で塞ぐ。親子の FK は必ず「入れ物」の列を含む複合 FK にし、
--       別のテナント / スペース / ページの行を親にできないようにする。
--       木はそれぞれの入れ物の中で閉じる: ページの木はスペースの中、ブロックの木はページの中。
--       入れ物をまたぐ親子を許すと、入れ物を消したときに ON DELETE CASCADE が
--       別の入れ物に残るはずの行まで道連れにする。
--       そのために参照先へ (workspace_id, …, id) の複合 UNIQUE を張る。id 単独の PK では
--       FK の参照列に複数列を指定できないため、実データ上は冗長でも足場として要る。
--
--   (2) 並び順は分数インデックス（internal/pkg/fracindex）が採番する文字列キー。
--       同じ親の中で position が重複しないことを部分 UNIQUE で守り、既定値は置かない（採番はアプリ側）。

-- ワークスペース: ナレッジ基盤のテナント境界。
CREATE TABLE IF NOT EXISTS workspaces (
    id         uuid PRIMARY KEY,
    -- slug は URL に出る短い識別子。テナント内ではなくグローバルに一意。
    slug       varchar(64) NOT NULL,
    name       varchar(200) NOT NULL,
    -- GORM の autoCreateTime / autoUpdateTime は使わないため、DB 側の既定値で必ず埋まるようにする。
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT uq_workspaces_slug UNIQUE (slug),
    -- URL に出る識別子は空文字禁止・長さ上限（アプリ側検証と二重の壁）。
    CONSTRAINT ck_workspaces_slug_len CHECK (char_length(slug) BETWEEN 1 AND 64)
);

-- スペース: ワークスペース内のページの束（部門・プロジェクト単位の入れ物）。
CREATE TABLE IF NOT EXISTS spaces (
    id           uuid PRIMARY KEY,
    workspace_id uuid NOT NULL,
    -- key はワークスペース内で一意な短い識別子（例: "eng"）。
    "key"        varchar(64) NOT NULL,
    name         varchar(200) NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    -- ワークスペースの物理削除で配下も消える（運用ではアーカイブを使う想定で、物理削除は例外的な操作）。
    CONSTRAINT fk_spaces_workspace FOREIGN KEY (workspace_id)
        REFERENCES workspaces (id) ON DELETE CASCADE,
    CONSTRAINT uq_spaces_workspace_key UNIQUE (workspace_id, "key"),
    -- pages からの複合 FK の参照先。id の PK があるので実データ上は冗長だが、
    -- 「テナント越えを FK で塞ぐ」ための足場として要る。
    CONSTRAINT uq_spaces_workspace_id UNIQUE (workspace_id, id),
    CONSTRAINT ck_spaces_key_len CHECK (char_length("key") BETWEEN 1 AND 64)
);

-- ページ: ナレッジ基盤の 1 ページ。parent_id の自己参照で木をなす（無限入れ子）。
CREATE TABLE IF NOT EXISTS pages (
    id                 uuid PRIMARY KEY,
    workspace_id       uuid NOT NULL,
    space_id           uuid NOT NULL,
    -- parent_id が NULL ならスペース直下（ルート）。
    parent_id          uuid,
    -- position のコレーションは "C"（バイト順）に固定する。
    -- 分数インデックスは「文字列の辞書順 = 並び順」が前提で、Go 側はバイト比較で判断する。
    -- DB の既定がロケール依存のコレーション（例: en_US.utf8）だと 'a' < 'B' のように並び、
    -- ORDER BY position がアプリの認識とずれる。列の定義で最初から揃えておく。
    "position"         text COLLATE "C" NOT NULL,
    title              varchar(200) NOT NULL DEFAULT '',
    -- 作成者（users.id）。users への FK は張らない（ナレッジ基盤の骨格に閉じるため）。
    created_by_user_id bigint NOT NULL,
    -- archived_at が NULL の行が現役。物理削除ではなくアーカイブで隠す運用のため、
    -- position の一意性はアーカイブ済みを除外した部分 UNIQUE で守る（下の CREATE UNIQUE INDEX）。
    archived_at        timestamptz,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),

    -- ページは「同じワークスペースの space」にしか属せない。
    CONSTRAINT fk_pages_space FOREIGN KEY (workspace_id, space_id)
        REFERENCES spaces (workspace_id, id) ON DELETE CASCADE,
    -- 親は「同じワークスペースの、同じスペースの」ページだけ。親の物理削除で子孫も消える。
    -- ページの木はスペースの中で閉じる。スペースはページの入れ物であり、木がスペースをまたぐと
    -- パンくず（祖先をたどると別スペースに出る）・サブツリー一括取得・スペース単位の権限が
    -- すべて破綻するため、space_id まで一致を要求する。
    -- workspace だけの一致だと、スペース A のページがスペース B のページを親に持ててしまい、
    -- スペース B を消したときに fk_pages_space の CASCADE で B のページが消え、続けて
    -- こちらの CASCADE がスペース A に残るはずの子ページまで道連れにする。
    --
    -- parent_id は NULL 可（ルート）。複合 FK は既定の MATCH SIMPLE なので、
    -- 参照列に 1 つでも NULL があれば検査自体が行われない ＝ ルートページは素通りする。
    -- これは意図どおり: ルートの workspace_id / space_id は fk_pages_space 側で必ず検査されるため、
    -- テナント越え・スペース越えの抜け道にはならない。
    --
    -- 副作用（意図した挙動）: ページを別スペースへ移すときは、子孫の space_id も同じ文で
    -- 更新しないと FK 違反になる。木の一部だけがスペースをまたぐ「中途半端な移動」を DB が防ぐ。
    CONSTRAINT fk_pages_parent FOREIGN KEY (workspace_id, space_id, parent_id)
        REFERENCES pages (workspace_id, space_id, id) ON DELETE CASCADE,
    -- blocks / page_paths からの複合 FK の参照先。space_id を持たないテーブルから
    -- ページを参照するには (workspace_id, id) の形が要る（下の fk_blocks_page /
    -- fk_page_paths_page / fk_page_paths_ancestor の 3 本が使う）。
    CONSTRAINT uq_pages_workspace_id UNIQUE (workspace_id, id),
    -- 親ページの FK を「同じスペース」まで絞るための足場。
    CONSTRAINT uq_pages_workspace_space_id UNIQUE (workspace_id, space_id, id),
    -- 自分自身を親にできない（1 行で閉じた循環を作らせない。多段の循環はアプリ側で検出する）。
    CONSTRAINT ck_pages_parent_not_self CHECK (parent_id IS NULL OR parent_id <> id),
    -- position は空文字だと順序として意味を持たない（fracindex は空文字を返さない）。
    CONSTRAINT ck_pages_position_not_empty CHECK ("position" <> '')
);

-- ブロック: ページ本文を構成する 1 行（段落・見出し・リスト項目・表のセル …）。
-- 入れ子（リストや表）は parent_id の自己参照で表す。
CREATE TABLE IF NOT EXISTS blocks (
    id           uuid PRIMARY KEY,
    workspace_id uuid NOT NULL,
    page_id      uuid NOT NULL,
    -- parent_id が NULL ならページ直下（トップレベル）。
    parent_id    uuid,
    -- pages.position と同じ理由でバイト順に固定する。
    "position"   text COLLATE "C" NOT NULL,
    -- ProseMirror（tiptap）のノード名。値は domain.BlockType が正。
    type         varchar(32) NOT NULL,
    -- ProseMirror の attrs（見出しの level、コードブロックの language など）。
    -- 属性が無いノードでも空オブジェクト {} を入れる（NULL と {} の二通りを作らない）。
    attrs        jsonb NOT NULL DEFAULT '{}'::jsonb,
    -- 葉ノードのインライン内容（text ノードとマークの配列）。
    -- リストや表のような容器ノードは子をブロック行として持つため NULL にする。
    inline       jsonb,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    -- ブロックは「同じワークスペースの page」にしか属せない。
    CONSTRAINT fk_blocks_page FOREIGN KEY (workspace_id, page_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    -- 親は「同じワークスペースの、同じページの」ブロックだけ。
    -- ブロックの木は 1 ページの中で閉じるものなので、page_id まで一致を要求する。
    -- workspace だけを一致させると、ページ A のブロックをページ B のブロックの親にでき、
    -- ページ A を消したときに ON DELETE CASCADE がページ B の本文まで消してしまう。
    -- MATCH SIMPLE の扱いは pages と同じで、parent_id が NULL（トップレベル）なら検査されない。
    -- その場合の workspace_id / page_id の正しさは fk_blocks_page 側で担保される。
    CONSTRAINT fk_blocks_parent FOREIGN KEY (workspace_id, page_id, parent_id)
        REFERENCES blocks (workspace_id, page_id, id) ON DELETE CASCADE,
    -- 親ブロックの FK を「同じページ」まで絞るための足場。
    CONSTRAINT uq_blocks_workspace_page_id UNIQUE (workspace_id, page_id, id),
    CONSTRAINT ck_blocks_parent_not_self CHECK (parent_id IS NULL OR parent_id <> id),
    CONSTRAINT ck_blocks_position_not_empty CHECK ("position" <> ''),
    -- attrs は ProseMirror の attrs なので必ず object（属性が無いノードでも {}）。
    CONSTRAINT ck_blocks_attrs_object CHECK (jsonb_typeof(attrs) = 'object'),
    -- inline は葉ノードの content 配列。容器ノードでは NULL にする。
    CONSTRAINT ck_blocks_inline_array CHECK (inline IS NULL OR jsonb_typeof(inline) = 'array')
);

-- page_paths: ページの祖先関係を平らに持つ派生テーブル（closure table）。自分自身も depth=0 の行として持つ。
-- pages.parent_id の連鎖だけでも木は表せるが、パンくず・サブツリー一括取得・移動時の循環検出を
-- 再帰クエリなしの 1 回の JOIN で済ませるためにこの索引を別に持つ。正本は pages.parent_id 側。
CREATE TABLE IF NOT EXISTS page_paths (
    workspace_id uuid NOT NULL,
    page_id      uuid NOT NULL,
    ancestor_id  uuid NOT NULL,
    -- 祖先までの距離。自分自身が 0、親が 1。
    depth        integer NOT NULL,

    CONSTRAINT page_paths_pkey PRIMARY KEY (page_id, ancestor_id),
    -- 1 行で「子孫」と「祖先」の 2 ページを組にするため、単独 FK を 2 本張るだけでは
    -- 別ワークスペースの 2 ページを組にした行が作れてしまう（両方の FK を通ってしまう）。
    -- 行自身の workspace_id を軸にした複合 FK にして、組になる 2 ページが同じワークスペースに
    -- 属することを DB 側で保証する。ページが消えたら派生であるこの行も一緒に消す。
    --
    -- FK で守るのは「組になる 2 ページが実在し、同じワークスペースに属すること」まで。
    -- 1 行だけで判定できる depth の不変条件は下の ck_page_paths_depth で別に塞ぐ。
    --
    -- 一方「depth が実際の親子の距離と一致するか、祖先の連鎖に抜けや余りが無いか」は DB では守らない。
    -- それは 1 行を見ても判定できず、pages の木をたどって初めて分かる複数行にまたがる不変条件で、
    -- 宣言的な制約（行ごとの CHECK / FK）では表せないため。この表は pages.parent_id から導ける
    -- 派生データなので、正本である pages 側の制約で木の形を守り、closure 全体の整合は行を書く側の
    -- 責務とする。なお page_paths は常に FK の子側で、この表の行が壊れても他の行を CASCADE で
    -- 消すことはない（壊れ方が表示の乱れに閉じ、他のデータを失わない）。
    CONSTRAINT fk_page_paths_page FOREIGN KEY (workspace_id, page_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_page_paths_ancestor FOREIGN KEY (workspace_id, ancestor_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    -- 1 行だけで判定できる不変条件: depth は祖先までの距離なので負にならず、
    -- depth=0 の行は自分自身を指す行「だけ」（逆に自己参照の行は必ず depth=0）。
    -- パンくずは ORDER BY depth で組み立てるため、ここが崩れると pages.parent_id（正本）は
    -- 正しいのに表示だけが壊れ、原因を追いにくい形で顕在化する。
    CONSTRAINT ck_page_paths_depth CHECK (depth >= 0 AND (depth = 0) = (page_id = ancestor_id))
);

-- page_snapshots: ページのブロック行を組み直した ProseMirror ドキュメント（読み取り用のキャッシュ）。
-- 表示のたびにブロック行を木に組み直すと 1 ページで数百行の取得と再帰的な組み立てが要るため、
-- 編集のたびに 1 つの jsonb へ焼き直して読み出しを 1 行の取得に落とす。
-- 正本はあくまで blocks 側で、この行は失っても blocks から再生成できる派生データ。
CREATE TABLE IF NOT EXISTS page_snapshots (
    page_id  uuid PRIMARY KEY,
    -- tiptap の getJSON() 相当（type='doc' の ProseMirror ドキュメント）。
    doc      jsonb NOT NULL,
    -- 焼き直した時刻。ブロックの更新時刻より古ければ作り直す判断に使う。
    built_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_page_snapshots_page FOREIGN KEY (page_id)
        REFERENCES pages (id) ON DELETE CASCADE,
    -- 壊れた snapshot は読み取りキャッシュとしてそのまま返り、エディタがページを開けなくなるため、
    -- rich_documents.doc と同じ形で入口を塞ぐ。
    CONSTRAINT ck_page_snapshots_doc CHECK (jsonb_typeof(doc) = 'object' AND doc->>'type' = 'doc')
);

-- --- 並び順の一意性（部分 UNIQUE は WHERE 付きなのでテーブル定義に書けず、索引として張る）---
-- 同じ親の中で position が重複しないこと。ページはアーカイブ済みを除外する
-- （アーカイブは「一覧から隠す」だけで行は残るため、現役の並びだけを守る）。
CREATE UNIQUE INDEX IF NOT EXISTS uq_pages_parent_position
    ON pages (parent_id, "position") WHERE archived_at IS NULL;
-- ルート直下（parent_id IS NULL）は上の索引では守れない。UNIQUE 索引は NULL 同士を
-- 別物として扱うため、parent_id が NULL の行同士は何度でも同じ position を持ててしまう。
-- ルートの並びはスペース単位なので、スペースを軸にした部分 UNIQUE を別に張る。
CREATE UNIQUE INDEX IF NOT EXISTS uq_pages_space_position
    ON pages (space_id, "position") WHERE parent_id IS NULL AND archived_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_blocks_parent_position
    ON blocks (parent_id, "position");
-- ブロックも同じ理由で、ページ直下（parent_id IS NULL）はページを軸に守る。
CREATE UNIQUE INDEX IF NOT EXISTS uq_blocks_page_position
    ON blocks (page_id, "position") WHERE parent_id IS NULL;

-- --- 取得経路の索引（FK の子側は自動では索引が張られないため明示する）---
-- ワークスペース / スペース単位の一覧と、親を辿る取得・CASCADE 削除の走査に効かせる。
CREATE INDEX IF NOT EXISTS idx_spaces_workspace_id ON spaces (workspace_id);
CREATE INDEX IF NOT EXISTS idx_pages_workspace_id ON pages (workspace_id);
CREATE INDEX IF NOT EXISTS idx_pages_space_id ON pages (space_id);
CREATE INDEX IF NOT EXISTS idx_pages_parent_id ON pages (parent_id);
-- アーカイブ済みの除外・アーカイブ一覧の取得に使う。
CREATE INDEX IF NOT EXISTS idx_pages_archived_at ON pages (archived_at);
CREATE INDEX IF NOT EXISTS idx_blocks_workspace_id ON blocks (workspace_id);
CREATE INDEX IF NOT EXISTS idx_blocks_page_id ON blocks (page_id);
CREATE INDEX IF NOT EXISTS idx_blocks_parent_id ON blocks (parent_id);
CREATE INDEX IF NOT EXISTS idx_page_paths_workspace_id ON page_paths (workspace_id);
-- 祖先からサブツリーを引く経路（PK は (page_id, ancestor_id) なので ancestor_id 単独では効かない）。
CREATE INDEX IF NOT EXISTS idx_page_paths_ancestor_id ON page_paths (ancestor_id);
