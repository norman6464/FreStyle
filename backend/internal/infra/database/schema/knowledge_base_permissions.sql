-- ナレッジ基盤の権限モデル（principals / principal_members / workspace_grants /
-- space_grants / page_restrictions / page_allow_lists / share_links）の DDL。
--
-- knowledge_base.sql（骨格 6 テーブル）と同じ扱い: このファイルが実スキーマの正本であり、
-- 同時に sqlc の型付け入力でもある（backend/sqlc.yaml の schema に登録済み）。
-- 起動時に database.ApplyKnowledgeBaseSchema が骨格の DDL に続けてそのまま流す（冪等）。
-- CREATE ... IF NOT EXISTS だけで冪等性を成り立たせ、DO ブロックは書かない（sqlc がパースするため）。
--
-- 骨格と別ファイルにする理由は 2 つ:
--   (1) 骨格 6 テーブルは「ページの木そのもの」で、権限は「誰がそれを触れるか」という別の関心。
--       1 枚に混ぜると、どちらを読みたいときも全部を読むことになる。
--   (2) こちらは users（GORM AutoMigrate が作る既存テーブル）へ FK を張る。骨格側は
--       ナレッジ基盤だけで閉じており、その依存の有無をファイル境界で見えるようにする。
--       適用順は database.Migrate が AutoMigrate → ApplyKnowledgeBaseSchema なので users は必ず先にある。
--
-- 設計の柱（骨格の 2 つに加えて）:
--
--   (3) 主体（principal）を 1 つの表に集める。ユーザー・グループ・スペース全員・公開リンクは
--       「権限を与える相手」という点で同じなので、grant / restriction 側から見て 1 本の FK で済む。
--       主体ごとに表を分けると grant / restriction が主体の種類だけ列（または表）を持つことになり、
--       権限を解く SQL が主体の種類だけ分岐する。
--
--   (4) 種類（kind）によって使う列が変わるので、CHECK で「その kind のときだけ非 NULL」を強制する。
--       任意の key/value に逃がす（EAV）ことはしない。列は意味を持ったまま、
--       「いつ埋まるか」だけを制約で表す。
--
--   (5) 権限は「既定（grants）＋ 例外（page_restrictions）」だけで表す。
--       全ページへ ACL を展開する方式は解決が 1 行の取得で済む代わりに、ページを 1 回動かす /
--       メンバーを 1 人足すだけで数万行を書き換える。ページ移動が日常の道具である以上、
--       書き込み側の代償が大きすぎる。例外はごく少数のページにしか付かない性質を使い、
--       行を持つのは例外だけにして、解決は page_paths（closure）を 1 回 JOIN するだけで済ませる。
--       既定は入れ物の階層に合わせて 2 段（workspace_grants / space_grants）持つ。

-- principals: 権限を与える相手（主体）。
--
-- **この表がワークスペース所属の唯一の表現**（「そのワークスペースに kind='user' の行がある」
-- ＝ そのユーザーはメンバー）。workspace_memberships のようなメンバーシップ専用の表は
-- 作らない・足さないこと。作ると「principal はあるがメンバーではない」「メンバーだが
-- principal が無い」の 2 通りのずれが生まれ、どちらが正かを決められなくなる。
-- 所属の追加 / 削除はこの表への 1 行の INSERT / DELETE で表す。
--
-- 「未所属」は行が無いことで表す。専用の値（0 や空文字）は置かない。既存の users.company_id が
-- NULL と 0 の 2 通りで未所属を表していて層をまたいで混在しているのと同じ轍を踏まないため。
CREATE TABLE IF NOT EXISTS principals (
    id           uuid PRIMARY KEY,
    workspace_id uuid NOT NULL,
    -- kind の値は domain.PrincipalKind が正（user / group / space_all / share_link）。
    kind         varchar(16) NOT NULL,
    -- user_id は kind='user' のときだけ埋まる（既存 users への参照）。
    user_id      bigint,
    -- space_id は kind='space_all'（そのスペースの全員）のときだけ埋まる。
    space_id     uuid,
    -- page_id は kind='share_link'（公開リンクの来訪者）のときだけ埋まる。そのリンクの対象ページ。
    -- 主体を「それが意味を持つ入れ物」に必ず結び付けるためで、こうするとページを物理削除したときに
    -- 主体もリンクも CASCADE で一緒に消える。逆向き（share_links → principals）の FK だけでは、
    -- ページを消してもリンクの行だけが消えて主体が残り、誰も指さない行が溜まる。
    page_id      uuid,
    -- name は kind='group' の表示名。ほかの kind は名前を持たない
    -- （ユーザー名は users、スペース名は spaces が正本。ここへ写すと二重管理になる）。
    name         varchar(200) NOT NULL DEFAULT '',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_principals_workspace FOREIGN KEY (workspace_id)
        REFERENCES workspaces (id) ON DELETE CASCADE,
    -- users への FK は張る。principals はナレッジ基盤とアプリのユーザーを結ぶ唯一の接点で、
    -- ここが緩いと「消えたユーザーの principal に権限が残る」＝ 別人が同じ id を再取得したときに
    -- 権限を引き継いでしまう。骨格側の pages.created_by_user_id が FK を持たないのは、
    -- あちらが既存テーブル（IF NOT EXISTS では後から制約を足せない）だからで、
    -- 新しく作るこの表には最初から張れる。
    CONSTRAINT fk_principals_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE,
    -- スペース全員の主体は「同じワークスペースの space」にしか結び付かない。
    CONSTRAINT fk_principals_space FOREIGN KEY (workspace_id, space_id)
        REFERENCES spaces (workspace_id, id) ON DELETE CASCADE,
    -- 公開リンクの主体は「同じワークスペースの page」にしか結び付かない。
    CONSTRAINT fk_principals_page FOREIGN KEY (workspace_id, page_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    -- grant / restriction / share_link からの複合 FK の参照先。id の PK があるので実データ上は
    -- 冗長だが、「別ワークスペースの principal に権限を張れない」を FK で塞ぐ足場として要る。
    CONSTRAINT uq_principals_workspace_id UNIQUE (workspace_id, id),
    -- kind まで含めた足場。参照側が「この列は group の principal でなければならない」を
    -- FK で言えるようにする（principal_members / share_links が使う）。
    CONSTRAINT uq_principals_workspace_kind_id UNIQUE (workspace_id, kind, id),
    -- share_links からの複合 FK の参照先。リンクが持つ page_id と、その主体が持つ page_id が
    -- 必ず同じページを指すことを FK で言えるようにする（2 か所に同じ値を持つ以上、
    -- 食い違わないことは制約で担保する）。
    CONSTRAINT uq_principals_workspace_kind_page_id UNIQUE (workspace_id, kind, page_id, id),
    CONSTRAINT ck_principals_kind CHECK (kind IN ('user', 'group', 'space_all', 'share_link')),
    -- 使う列は kind で決まる。「その kind のときだけ非 NULL」を等式で書き、
    -- 片方向（NOT NULL なのに kind が違う）も同時に塞ぐ。
    CONSTRAINT ck_principals_user_id CHECK ((kind = 'user') = (user_id IS NOT NULL)),
    CONSTRAINT ck_principals_space_id CHECK ((kind = 'space_all') = (space_id IS NOT NULL)),
    CONSTRAINT ck_principals_page_id CHECK ((kind = 'share_link') = (page_id IS NOT NULL)),
    CONSTRAINT ck_principals_name CHECK ((kind = 'group') = (name <> ''))
);

-- principal_members: グループの所属（group principal ↔ member principal）。
--
-- グループの入れ子は許さない。member 側を kind='user' に固定することで、
-- 「あるユーザーの主体の集合」を再帰なしの 1 回の UNION で出せる
-- （入れ子を許すと権限解決に再帰 CTE が要るうえ、グループ同士の循環を防ぐ手当ても要る）。
--
-- kind の固定は生成列（GENERATED ALWAYS AS ... STORED）で行う。定数なので INSERT / UPDATE から
-- 値を渡せず、書き手が間違えようがない。CHECK 付きの普通の列にすると「書けるが必ず同じ値」に
-- なり、実質使われない列を 2 つ抱えることになる。ここは足場であって属性ではない。
--
-- これが無いと、たとえば member_principal_id に他人のユーザー主体ではなくグループ主体を
-- 入れることでグループを入れ子にでき、解決 SQL（1 段しか辿らない）が黙って権限を取りこぼす。
-- 「取りこぼす」＝ 見えるはずのページが見えないだけなので、権限の穴ではないが原因を追いにくい。
CREATE TABLE IF NOT EXISTS principal_members (
    workspace_id        uuid NOT NULL,
    group_principal_id  uuid NOT NULL,
    member_principal_id uuid NOT NULL,
    -- FK の足場（定数の生成列）。テーブルの属性ではない。
    group_kind          varchar(16) GENERATED ALWAYS AS ('group') STORED,
    member_kind         varchar(16) GENERATED ALWAYS AS ('user') STORED,
    created_at          timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT principal_members_pkey PRIMARY KEY (group_principal_id, member_principal_id),
    CONSTRAINT fk_principal_members_group FOREIGN KEY (workspace_id, group_kind, group_principal_id)
        REFERENCES principals (workspace_id, kind, id) ON DELETE CASCADE,
    CONSTRAINT fk_principal_members_member FOREIGN KEY (workspace_id, member_kind, member_principal_id)
        REFERENCES principals (workspace_id, kind, id) ON DELETE CASCADE
);

-- workspace_grants: ワークスペース全体での既定の役割。配下の全スペースに効く。
--
-- スペース単位の grant だけでは「テナント全体の管理者」を表すのにスペースの数だけ grant を
-- 張って回ることになり、スペースが増えるたびに漏れる。入れ物の階層が workspace ⊃ space である
-- 以上、既定も同じ 2 段で持つ。
--
-- scope_type / scope_id を持つ 1 枚の汎用 grants 表にまとめる案は採らない。scope_id が
-- workspaces と spaces の 2 つの表を指すことになり、FK で参照先の実在もテナントの一致も
-- 守れなくなる。このスキーマは一貫して「境界を FK で塞ぐ」ことを優先しており、
-- 表が 1 枚増える代わりに両方とも複合 FK で守れる形を選ぶ。
--
-- workspaces への直接の FK は張らない。principals への複合 FK が (workspace_id, principal_id) で
-- 実在する principal との一致を要求し、その principal 自身が workspaces へ FK を持つため、
-- ワークスペースの実在も削除時の CASCADE も推移的に効く。
CREATE TABLE IF NOT EXISTS workspace_grants (
    workspace_id uuid NOT NULL,
    principal_id uuid NOT NULL,
    -- "role" の値は domain.GrantRole が正（admin / editor / commenter / viewer）。
    "role"       varchar(16) NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT workspace_grants_pkey PRIMARY KEY (workspace_id, principal_id),
    CONSTRAINT fk_workspace_grants_principal FOREIGN KEY (workspace_id, principal_id)
        REFERENCES principals (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT ck_workspace_grants_role CHECK ("role" IN ('admin', 'editor', 'commenter', 'viewer'))
);

-- space_grants: そのスペースでの既定の役割。ページに例外が無いときは、これと workspace_grants の
-- うち強い方が実効権限になる（domain.GrantRole.Rank 参照。弱い方を採る規則にすると
-- スペースを 1 つ作って viewer を張るだけでワークスペース管理者を締め出せてしまう）。
--
-- 1 つの principal がひとつのスペースで持つ役割は 1 つなので、代理キーを置かず自然キーを PK にする
-- （代理キーを置くと「同じ principal に viewer と editor の 2 行」が作れてしまい、
-- どちらが正かをアプリで決める羽目になる）。
CREATE TABLE IF NOT EXISTS space_grants (
    workspace_id uuid NOT NULL,
    space_id     uuid NOT NULL,
    principal_id uuid NOT NULL,
    -- "role" の値は domain.GrantRole が正（admin / editor / commenter / viewer）。
    "role"       varchar(16) NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT space_grants_pkey PRIMARY KEY (workspace_id, space_id, principal_id),
    CONSTRAINT fk_space_grants_space FOREIGN KEY (workspace_id, space_id)
        REFERENCES spaces (workspace_id, id) ON DELETE CASCADE,
    -- workspace_id を含めることで「別ワークスペースの principal への grant」を DB が弾く。
    -- principal_id 単独の FK だと、行の workspace_id と principal の workspace_id が
    -- 食い違っていても両方の FK を通ってしまう（テナント越えの権限昇格になる）。
    CONSTRAINT fk_space_grants_principal FOREIGN KEY (workspace_id, principal_id)
        REFERENCES principals (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT ck_space_grants_role CHECK ("role" IN ('admin', 'editor', 'commenter', 'viewer'))
);

-- page_restrictions: そのページ以下だけ既定を上書きする例外。行を持つのは例外だけ。
--
-- 実効権限の決め方（domain.ResolvePagePermission が唯一の実装。ここは同じ規則の要約）:
--   1. 対象ページ自身から根までの経路のどこかに自分宛ての deny があれば不許可。
--      deny は allow に勝ち、経路全体で効く
--   2. 経路上に許可リスト制の段（page_allow_lists の印）があれば、そのうち最も近い段に
--      自分の allow 行があるかで決まる。無ければ既定が admin でも不許可（限定公開）
--   3. 許可リスト制の段が経路に無ければ grants の既定に従う。
--      deny 行だけの段は「名指しの除外」で、ほかの人の既定は変えない
--
-- 2 と 3 の分かれ目が要るのは、allow と deny で意味が逆だから。allow を 1 つ足した瞬間に
-- 「載っていない人は入れない」に切り替わり（限定公開。その印は page_allow_lists）、deny だけの段は
-- 「その人だけ外す」で他は既定のまま、という 2 つの使い方を 1 つの表で両立させる。
--
-- deny を「最も近い段」だけで見てはいけない。deny 行しか無い段が最近段になると 3 が働き、
-- より遠い祖先の許可リストが第三者への deny 1 行で解除されてしまう。
-- 一方 allow は最も近い段だけで決める（近い許可リストが遠い許可リストを上書きする）。
-- 「この枝だけもう少し広く共有する」を書けるようにするための意図した非対称。
--
-- PK に capability を含めるので、同じ (ページ, 主体, ケイパビリティ) に allow と deny の
-- 2 行は作れない。矛盾した設定はそもそも保存できない。
CREATE TABLE IF NOT EXISTS page_restrictions (
    workspace_id uuid NOT NULL,
    page_id      uuid NOT NULL,
    principal_id uuid NOT NULL,
    -- capability の値は domain.Capability が正（view / edit）。
    capability   varchar(8) NOT NULL,
    -- mode の値は domain.RestrictionMode が正（allow / deny）。
    mode         varchar(8) NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT page_restrictions_pkey PRIMARY KEY (workspace_id, page_id, principal_id, capability),
    CONSTRAINT fk_page_restrictions_page FOREIGN KEY (workspace_id, page_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_page_restrictions_principal FOREIGN KEY (workspace_id, principal_id)
        REFERENCES principals (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT ck_page_restrictions_capability CHECK (capability IN ('view', 'edit')),
    CONSTRAINT ck_page_restrictions_mode CHECK (mode IN ('allow', 'deny'))
);

-- page_allow_lists: 「このページのこのケイパビリティは許可リスト制（限定公開）である」という印。
--
-- 限定公開かどうかを page_restrictions の allow 行の有無で表してはいけない。
-- principal_id は principals へ ON DELETE CASCADE なので、許可リストに載っている主体を
-- 消すと allow 行も一緒に消える。行の有無が印を兼ねていると、その瞬間にその段の制限が
-- 0 行になり、解決は「制限が無い」＝ 既定（例: スペース全員 editor）へ戻る。
-- つまり退職者のオフボーディングや部署の統廃合という通常運用の 1 操作で、
-- 無関係な第三者に限定公開のページが子孫ごと開く。ページ移動での失効を
-- ErrPageMoveVoidsSpaceRestriction が経路で止めているのと違い、こちらは
-- 「思いついた経路を塞ぐ」では足りない（削除の入口は増える）ため構造で断つ。
--
-- 印は主体を参照しないので、どの主体が消えても残る。残った結果は
-- 「許可リストが空 ＝ 誰も載っていない」で、fail-closed（閉じる側）に倒れる。
-- 権限管理の usecase（GrantWorkspaceRole / GrantSpaceRole / SetPageRestriction）は
-- ページの閲覧・編集を要求しないので、閉じても管理者は張り直して復旧できる。
--
-- 印の増減は「明示的に allow 行を足した / 減らした」操作だけが行う（repository が同じ
-- トランザクションで揃える）。deny 行の解除など allow に触れない操作では動かさない。
-- 動かしてしまうと、無関係な 1 行の解除で限定公開が解けるという同じ穴を作ることになる。
CREATE TABLE IF NOT EXISTS page_allow_lists (
    workspace_id uuid NOT NULL,
    page_id      uuid NOT NULL,
    -- capability の値は domain.Capability が正（view / edit）。
    capability   varchar(8) NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT page_allow_lists_pkey PRIMARY KEY (workspace_id, page_id, capability),
    CONSTRAINT fk_page_allow_lists_page FOREIGN KEY (workspace_id, page_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    CONSTRAINT ck_page_allow_lists_capability CHECK (capability IN ('view', 'edit'))
);

-- share_links: ログイン不要の公開 URL。
--
-- 来訪者は kind='share_link' の principal として扱う。こうすると「公開リンクからの閲覧者」も
-- ほかの主体と同じく page_restrictions の対象にでき、「ページ全体を公開しつつ 1 枚の子ページだけ
-- 除外する」を deny 行 1 つで書ける。権限解決の入口が主体ごとに分岐しない。
--
-- ただし既定（そのリンクで何ができるか）は grants ではなくこの表の capability で決める。
-- 公開リンクのために allow 行を足す設計にすると、その瞬間にそのページが「許可リスト」状態に
-- 切り替わり（上の 3）、それまで見えていたチームの全員が締め出される。既定の出どころだけを
-- 分け、例外の層は共有する。
--
-- token は平文で持たない。DB が漏れた時点で全リンクが開けるのを避けるため、SHA-256 の
-- ダイジェストだけを保存して照合はハッシュ同士で行う（トークンは十分な長さの乱数なので
-- 総当たりに強く、bcrypt のような遅いハッシュは要らない）。パスワードは人が選ぶ値なので
-- 逆に総当たりに弱く、こちらは bcrypt で持つ。
CREATE TABLE IF NOT EXISTS share_links (
    id                 uuid PRIMARY KEY,
    workspace_id       uuid NOT NULL,
    -- page_id はリンクの対象ページ。このページとその子孫が対象になる。
    page_id            uuid NOT NULL,
    principal_id       uuid NOT NULL,
    -- FK の足場（定数の生成列）。principal_members と同じ理由でここも生成列にする。
    principal_kind     varchar(16) GENERATED ALWAYS AS ('share_link') STORED,
    -- capability の値は domain.Capability が正（view / edit）。
    capability         varchar(8) NOT NULL,
    -- token_hash は共有 URL に載るトークンの SHA-256（32 バイト固定）。
    token_hash         bytea NOT NULL,
    -- password_hash は bcrypt。NULL ならパスワード無しで開ける。
    password_hash      text,
    -- expires_at が NULL なら無期限。
    expires_at         timestamptz,
    -- revoked_at が NULL なら有効。失効は行を消さず日付で残す（誰がいつ止めたかを追えるように）。
    revoked_at         timestamptz,
    created_by_user_id bigint NOT NULL,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_share_links_page FOREIGN KEY (workspace_id, page_id)
        REFERENCES pages (workspace_id, id) ON DELETE CASCADE,
    -- principal は「同じワークスペースの、kind='share_link' の、同じページに結び付いた」主体だけ。
    -- page_id まで参照列に含めることで、リンクと主体が別々のページを指す状態を作れなくする。
    CONSTRAINT fk_share_links_principal FOREIGN KEY (workspace_id, principal_kind, page_id, principal_id)
        REFERENCES principals (workspace_id, kind, page_id, id) ON DELETE CASCADE,
    CONSTRAINT fk_share_links_created_by FOREIGN KEY (created_by_user_id)
        REFERENCES users (id) ON DELETE CASCADE,
    -- トークンからリンクを 1 件引く経路。UNIQUE はその索引も兼ねる。
    CONSTRAINT uq_share_links_token_hash UNIQUE (token_hash),
    -- 1 つの share_link principal は 1 本のリンクだけを表す（使い回すと失効が効かなくなる）。
    CONSTRAINT uq_share_links_principal UNIQUE (principal_id),
    CONSTRAINT ck_share_links_capability CHECK (capability IN ('view', 'edit')),
    -- SHA-256 以外（平文トークンをそのまま入れた等）を入口で弾く。
    CONSTRAINT ck_share_links_token_hash_len CHECK (octet_length(token_hash) = 32),
    CONSTRAINT ck_share_links_password_hash CHECK (password_hash IS NULL OR password_hash <> '')
);

-- --- 一意性（部分 UNIQUE は WHERE 付きなのでテーブル定義に書けず、索引として張る）---
-- 1 ユーザー 1 ワークスペースにつき主体は 1 つ（重複メンバーを作らない）。
CREATE UNIQUE INDEX IF NOT EXISTS uq_principals_workspace_user
    ON principals (workspace_id, user_id) WHERE kind = 'user';
-- 1 スペースにつき「全員」の主体は 1 つ。
CREATE UNIQUE INDEX IF NOT EXISTS uq_principals_space_all
    ON principals (workspace_id, space_id) WHERE kind = 'space_all';
-- グループ名はワークスペース内で一意（同名グループが 2 つあると権限を張る先を人が選べない）。
CREATE UNIQUE INDEX IF NOT EXISTS uq_principals_group_name
    ON principals (workspace_id, name) WHERE kind = 'group';

-- --- 取得経路の索引（FK の子側は自動では索引が張られないため明示する）---
CREATE INDEX IF NOT EXISTS idx_principals_workspace_id ON principals (workspace_id);
CREATE INDEX IF NOT EXISTS idx_principals_user_id ON principals (user_id);
CREATE INDEX IF NOT EXISTS idx_principals_space_id ON principals (space_id);
CREATE INDEX IF NOT EXISTS idx_principals_page_id ON principals (page_id);
-- 「このユーザーが属するグループ」を引く経路（PK は group 側が先頭なので member 単独では効かない）。
CREATE INDEX IF NOT EXISTS idx_principal_members_member
    ON principal_members (workspace_id, member_principal_id);
-- 「この principal の grant / restriction」を引く経路（PK は入れ物側が先頭）。
-- principal を消すときの CASCADE 走査にも効く。
-- workspace_grants は PK が (workspace_id, principal_id) そのものなので追加の索引は要らない。
CREATE INDEX IF NOT EXISTS idx_space_grants_principal ON space_grants (workspace_id, principal_id);
CREATE INDEX IF NOT EXISTS idx_page_restrictions_principal
    ON page_restrictions (workspace_id, principal_id);
CREATE INDEX IF NOT EXISTS idx_share_links_page ON share_links (workspace_id, page_id);
CREATE INDEX IF NOT EXISTS idx_share_links_created_by ON share_links (created_by_user_id);
