//go:build integration

package persistence_test

import (
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/pkg/fracindex"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// PostgreSQL の SQLSTATE（制約違反の種別）。どの制約が効いたのかまで固定するために使う。
const (
	sqlStateForeignKeyViolation = "23503"
	sqlStateUniqueViolation     = "23505"
	sqlStateCheckViolation      = "23514"
)

// kbTables はナレッジ基盤のテーブル（TRUNCATE 対象）。子から先に並べる。
var kbTables = []string{"blocks", "page_paths", "page_snapshots", "pages", "spaces", "workspaces"}

// TestKnowledgeBaseSchema_Integration は ApplyKnowledgeBaseConstraints が張る制約を実 Postgres で固定する。
// repository はまだ無い（段 1-b で追加する）ため GORM から直接テーブルを操作して検証する。
func TestKnowledgeBaseSchema_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)

	t.Run("別ワークスペースの space にはページを紐づけられない", func(t *testing.T) {
		testsupport.TruncateAll(t, db, kbTables...)
		wsA := createWorkspace(t, db, "ws-a")
		wsB := createWorkspace(t, db, "ws-b")
		spaceB := createSpace(t, db, wsB, "eng")

		// workspace は A なのに space は B のもの → 複合 FK が弾く。
		page := newPage(wsA.ID, spaceB.ID, nil, "V")
		err := db.Create(page).Error
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_pages_space")
	})

	t.Run("別ワークスペースのページは親にできない", func(t *testing.T) {
		testsupport.TruncateAll(t, db, kbTables...)
		wsA := createWorkspace(t, db, "ws-a")
		wsB := createWorkspace(t, db, "ws-b")
		spaceA := createSpace(t, db, wsA, "eng")
		spaceB := createSpace(t, db, wsB, "eng") // key はワークスペース内で一意なので同名で良い
		parentB := createPage(t, db, wsB, spaceB, nil, "V")

		child := newPage(wsA.ID, spaceA.ID, &parentB.ID, "V")
		err := db.Create(child).Error
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_pages_parent")
	})

	t.Run("別ワークスペースのブロックは親にできない", func(t *testing.T) {
		testsupport.TruncateAll(t, db, kbTables...)
		wsA := createWorkspace(t, db, "ws-a")
		wsB := createWorkspace(t, db, "ws-b")
		spaceA := createSpace(t, db, wsA, "eng")
		spaceB := createSpace(t, db, wsB, "eng")
		pageA := createPage(t, db, wsA, spaceA, nil, "V")
		pageB := createPage(t, db, wsB, spaceB, nil, "V")
		parentB := createBlock(t, db, wsB, pageB, nil, "V", domain.BlockTypeBulletList)

		child := newBlock(wsA.ID, pageA.ID, &parentB.ID, "V", domain.BlockTypeListItem)
		err := db.Create(child).Error
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_blocks_parent")
	})

	t.Run("自分自身を親にはできない", func(t *testing.T) {
		testsupport.TruncateAll(t, db, kbTables...)
		ws := createWorkspace(t, db, "ws-a")
		space := createSpace(t, db, ws, "eng")
		page := createPage(t, db, ws, space, nil, "V")

		t.Run("pages", func(t *testing.T) {
			selfParent := newPage(ws.ID, space.ID, nil, "a")
			selfParent.ParentID = &selfParent.ID
			err := db.Create(selfParent).Error
			requirePgError(t, err, sqlStateCheckViolation, "ck_pages_parent_not_self")
		})

		t.Run("blocks", func(t *testing.T) {
			selfParent := newBlock(ws.ID, page.ID, nil, "a", domain.BlockTypeParagraph)
			selfParent.ParentID = &selfParent.ID
			err := db.Create(selfParent).Error
			requirePgError(t, err, sqlStateCheckViolation, "ck_blocks_parent_not_self")
		})
	})

	t.Run("同じ親の中で position が重複すると弾かれる", func(t *testing.T) {
		testsupport.TruncateAll(t, db, kbTables...)
		ws := createWorkspace(t, db, "ws-a")
		space := createSpace(t, db, ws, "eng")
		parent := createPage(t, db, ws, space, nil, "V")
		parentBlock := createBlock(t, db, ws, parent, nil, "V", domain.BlockTypeBulletList)

		t.Run("子ページ", func(t *testing.T) {
			createPage(t, db, ws, space, &parent.ID, "a")
			err := db.Create(newPage(ws.ID, space.ID, &parent.ID, "a")).Error
			requirePgError(t, err, sqlStateUniqueViolation, "uq_pages_parent_position")
		})

		t.Run("子ブロック", func(t *testing.T) {
			createBlock(t, db, ws, parent, &parentBlock.ID, "a", domain.BlockTypeListItem)
			err := db.Create(newBlock(ws.ID, parent.ID, &parentBlock.ID, "a", domain.BlockTypeListItem)).Error
			requirePgError(t, err, sqlStateUniqueViolation, "uq_blocks_parent_position")
		})
	})

	// parent_id が NULL の行同士は UNIQUE 索引では衝突しない（NULL は互いに別物として扱われる）ため、
	// ルート直下はスペース / ページを軸にした部分 UNIQUE で守っている。その効きを固定する。
	t.Run("ルート直下でも position の重複は弾かれる", func(t *testing.T) {
		testsupport.TruncateAll(t, db, kbTables...)
		ws := createWorkspace(t, db, "ws-a")
		space := createSpace(t, db, ws, "eng")

		t.Run("スペース直下のページ", func(t *testing.T) {
			createPage(t, db, ws, space, nil, "V")
			err := db.Create(newPage(ws.ID, space.ID, nil, "V")).Error
			requirePgError(t, err, sqlStateUniqueViolation, "uq_pages_space_position")
		})

		t.Run("ページ直下のブロック", func(t *testing.T) {
			page := createPage(t, db, ws, space, nil, "a")
			createBlock(t, db, ws, page, nil, "V", domain.BlockTypeParagraph)
			err := db.Create(newBlock(ws.ID, page.ID, nil, "V", domain.BlockTypeParagraph)).Error
			requirePgError(t, err, sqlStateUniqueViolation, "uq_blocks_page_position")
		})
	})

	t.Run("アーカイブ済みページは position の一意性から外れる", func(t *testing.T) {
		testsupport.TruncateAll(t, db, kbTables...)
		ws := createWorkspace(t, db, "ws-a")
		space := createSpace(t, db, ws, "eng")

		archived := newPage(ws.ID, space.ID, nil, "V")
		archivedAt := time.Now()
		archived.ArchivedAt = &archivedAt
		require.NoError(t, db.Create(archived).Error)

		// 現役のページが同じ position を取れる（アーカイブは並びを占有しない）。
		require.NoError(t, db.Create(newPage(ws.ID, space.ID, nil, "V")).Error)
	})

	t.Run("親ページの物理削除で子孫と派生テーブルが CASCADE で消える", func(t *testing.T) {
		testsupport.TruncateAll(t, db, kbTables...)
		ws := createWorkspace(t, db, "ws-a")
		space := createSpace(t, db, ws, "eng")
		parent := createPage(t, db, ws, space, nil, "V")
		child := createPage(t, db, ws, space, &parent.ID, "V")
		createBlock(t, db, ws, parent, nil, "V", domain.BlockTypeParagraph)
		createBlock(t, db, ws, child, nil, "V", domain.BlockTypeParagraph)
		createPagePath(t, db, parent.ID, parent.ID, 0)
		createPagePath(t, db, child.ID, child.ID, 0)
		createPagePath(t, db, child.ID, parent.ID, 1)
		createPageSnapshot(t, db, parent.ID)
		createPageSnapshot(t, db, child.ID)

		require.NoError(t, db.Exec(`DELETE FROM pages WHERE id = ?`, parent.ID).Error)

		require.Zero(t, countRows(t, db, &domain.Page{}), "子ページも消えること")
		require.Zero(t, countRows(t, db, &domain.Block{}), "両ページのブロックが消えること")
		require.Zero(t, countRows(t, db, &domain.PagePath{}), "page_paths が消えること（page_id / ancestor_id の両方向）")
		require.Zero(t, countRows(t, db, &domain.PageSnapshot{}), "page_snapshots が消えること")
	})

	t.Run("識別子の一意性と形式", func(t *testing.T) {
		testsupport.TruncateAll(t, db, kbTables...)
		ws := createWorkspace(t, db, "ws-a")
		createSpace(t, db, ws, "eng")

		t.Run("workspaces.slug は重複できない", func(t *testing.T) {
			err := db.Create(&domain.Workspace{ID: newID(), Slug: "ws-a", Name: "重複"}).Error
			requirePgError(t, err, sqlStateUniqueViolation, "uq_workspaces_slug")
		})

		t.Run("spaces.key はワークスペース内で重複できない", func(t *testing.T) {
			err := db.Create(&domain.Space{ID: newID(), WorkspaceID: ws.ID, Key: "eng", Name: "重複"}).Error
			requirePgError(t, err, sqlStateUniqueViolation, "uq_spaces_workspace_key")
		})

		t.Run("slug / key は空文字にできない", func(t *testing.T) {
			err := db.Create(&domain.Workspace{ID: newID(), Slug: "", Name: "空 slug"}).Error
			requirePgError(t, err, sqlStateCheckViolation, "ck_workspaces_slug_len")

			err = db.Create(&domain.Space{ID: newID(), WorkspaceID: ws.ID, Key: "", Name: "空 key"}).Error
			requirePgError(t, err, sqlStateCheckViolation, "ck_spaces_key_len")
		})
	})

	t.Run("blocks の attrs は既定 {} で object に限られる", func(t *testing.T) {
		testsupport.TruncateAll(t, db, kbTables...)
		ws := createWorkspace(t, db, "ws-a")
		space := createSpace(t, db, ws, "eng")
		page := createPage(t, db, ws, space, nil, "V")

		// Attrs 未設定なら DB 既定の {} が入る（NULL と {} の二通りを作らない）。
		block := createBlock(t, db, ws, page, nil, "V", domain.BlockTypeParagraph)
		var attrs string
		require.NoError(t, db.Raw(`SELECT attrs::text FROM blocks WHERE id = ?`, block.ID).Scan(&attrs).Error)
		require.JSONEq(t, `{}`, attrs)

		// object 以外（配列）は CHECK で弾く。
		invalid := newBlock(ws.ID, page.ID, nil, "a", domain.BlockTypeParagraph)
		invalid.Attrs = `[]`
		requirePgError(t, db.Create(invalid).Error, sqlStateCheckViolation, "ck_blocks_attrs_object")

		// inline は容器ノードでは NULL、葉ノードでは content 配列。object は弾く。
		invalidInline := newBlock(ws.ID, page.ID, nil, "b", domain.BlockTypeParagraph)
		objectInline := `{"type":"text"}`
		invalidInline.Inline = &objectInline
		requirePgError(t, db.Create(invalidInline).Error, sqlStateCheckViolation, "ck_blocks_inline_array")
	})

	// 分数インデックスは「文字列の辞書順 = 並び順」が前提。Go はバイト比較で判断するので、
	// DB の ORDER BY も同じ順序（C コレーション）でなければ並びが食い違う。
	t.Run("position は Go のバイト順と同じ順序で並ぶ", func(t *testing.T) {
		testsupport.TruncateAll(t, db, kbTables...)
		ws := createWorkspace(t, db, "ws-a")
		space := createSpace(t, db, ws, "eng")
		parent := createPage(t, db, ws, space, nil, "V")

		// 大文字・小文字・数字が混ざるキー（ロケール依存のコレーションだと 'a' < 'B' で並んでしまう）。
		positions := []string{"z", "A", "a", "V", "1", "Zz", "zA"}
		for _, p := range positions {
			createPage(t, db, ws, space, &parent.ID, p)
		}

		var got []string
		require.NoError(t, db.Raw(
			`SELECT "position" FROM pages WHERE parent_id = ? ORDER BY "position"`, parent.ID,
		).Scan(&got).Error)

		want := append([]string(nil), positions...)
		sort.Strings(want) // Go のバイト比較
		require.Equal(t, want, got)
	})

	// 実際に fracindex で採番したキーを流し込み、DB の ORDER BY が挿入した論理順と一致することを見る。
	t.Run("fracindex で採番したキーで並び替えが表現できる", func(t *testing.T) {
		testsupport.TruncateAll(t, db, kbTables...)
		ws := createWorkspace(t, db, "ws-a")
		space := createSpace(t, db, ws, "eng")
		page := createPage(t, db, ws, space, nil, "V")

		// 末尾に 3 件足してから、1 番目と 2 番目の間に 1 件割り込ませる。
		var order []string
		for i := 0; i < 3; i++ {
			prev := ""
			if len(order) > 0 {
				prev = order[len(order)-1]
			}
			key, err := fracindex.Between(prev, "")
			require.NoError(t, err)
			order = append(order, key)
			createBlock(t, db, ws, page, nil, key, domain.BlockTypeParagraph)
		}
		inserted, err := fracindex.Between(order[0], order[1])
		require.NoError(t, err)
		createBlock(t, db, ws, page, nil, inserted, domain.BlockTypeParagraph)
		order = append([]string{order[0], inserted}, order[1:]...)

		var got []string
		require.NoError(t, db.Raw(
			`SELECT "position" FROM blocks WHERE page_id = ? AND parent_id IS NULL ORDER BY "position"`, page.ID,
		).Scan(&got).Error)
		require.Equal(t, order, got)
	})

	// 制約を張った後の DB に対して起動時と同じ AutoMigrate が流れても壊れないこと。
	// UNIQUE を「制約」で張ると、GORM が自分の命名規則の制約名を DROP しようとして
	// 毎起動 AutoMigrate が落ちる（そのため UNIQUE 索引で張っている）。その回帰を止める。
	t.Run("制約適用後に AutoMigrate を再実行しても落ちない", func(t *testing.T) {
		require.NoError(t, database.AutoMigrateAll(db))
		require.NoError(t, database.ApplyKnowledgeBaseConstraints(db)) // 二重適用しても冪等

		// AutoMigrate は position 列のコレーション指定を剥がさない（剥がれると並びが狂う）。
		for _, table := range []string{"pages", "blocks"} {
			var collation string
			require.NoError(t, db.Raw(
				`SELECT collation_name FROM information_schema.columns
				 WHERE table_schema = current_schema() AND table_name = ? AND column_name = 'position'`, table,
			).Scan(&collation).Error)
			require.Equalf(t, "C", collation, "%s.position のコレーションが C ではありません", table)
		}
	})
}

// --- 以下、テスト用のヘルパ（repository が無いので GORM を直接叩く）---

// newID はテストデータの ID を採番する。本番の repository（段 1-b で追加）と同じ UUIDv7 に揃える。
func newID() string { return uuid.Must(uuid.NewV7()).String() }

func createWorkspace(t *testing.T, db *gorm.DB, slug string) *domain.Workspace {
	t.Helper()
	ws := &domain.Workspace{ID: newID(), Slug: slug, Name: slug}
	require.NoError(t, db.Create(ws).Error)
	return ws
}

func createSpace(t *testing.T, db *gorm.DB, ws *domain.Workspace, key string) *domain.Space {
	t.Helper()
	space := &domain.Space{ID: newID(), WorkspaceID: ws.ID, Key: key, Name: key}
	require.NoError(t, db.Create(space).Error)
	return space
}

func newPage(workspaceID, spaceID string, parentID *string, position string) *domain.Page {
	return &domain.Page{
		ID:              newID(),
		WorkspaceID:     workspaceID,
		SpaceID:         spaceID,
		ParentID:        parentID,
		Position:        position,
		Title:           "ページ",
		CreatedByUserID: 1, // users への FK は張らない（ナレッジ基盤の骨格に閉じるため）
	}
}

func createPage(t *testing.T, db *gorm.DB, ws *domain.Workspace, space *domain.Space, parentID *string, position string) *domain.Page {
	t.Helper()
	page := newPage(ws.ID, space.ID, parentID, position)
	require.NoError(t, db.Create(page).Error)
	return page
}

func newBlock(workspaceID, pageID string, parentID *string, position string, blockType domain.BlockType) *domain.Block {
	return &domain.Block{
		ID:          newID(),
		WorkspaceID: workspaceID,
		PageID:      pageID,
		ParentID:    parentID,
		Position:    position,
		Type:        blockType,
	}
}

func createBlock(t *testing.T, db *gorm.DB, ws *domain.Workspace, page *domain.Page, parentID *string, position string, blockType domain.BlockType) *domain.Block {
	t.Helper()
	block := newBlock(ws.ID, page.ID, parentID, position, blockType)
	require.NoError(t, db.Create(block).Error)
	return block
}

func createPagePath(t *testing.T, db *gorm.DB, pageID, ancestorID string, depth int) {
	t.Helper()
	require.NoError(t, db.Create(&domain.PagePath{PageID: pageID, AncestorID: ancestorID, Depth: depth}).Error)
}

func createPageSnapshot(t *testing.T, db *gorm.DB, pageID string) {
	t.Helper()
	snapshot := &domain.PageSnapshot{PageID: pageID, Doc: `{"type":"doc","content":[]}`, BuiltAt: time.Now()}
	require.NoError(t, db.Create(snapshot).Error)
}

func countRows(t *testing.T, db *gorm.DB, model any) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(model).Count(&n).Error)
	return n
}

// requirePgError は err が期待した SQLSTATE の制約違反で、かつ期待した制約名で落ちたことを確かめる。
// 「何かのエラーになった」ではなく「意図した制約が効いた」ところまで固定する。
func requirePgError(t *testing.T, err error, sqlState, constraint string) {
	t.Helper()
	require.Error(t, err)
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equalf(t, sqlState, pgErr.Code, "SQLSTATE が想定と異なります: %v", err)
	require.Equalf(t, constraint, pgErr.ConstraintName, "効いた制約が想定と異なります: %v", err)
}
