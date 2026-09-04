//go:build integration

package persistence_test

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

const rdDoc = `{"type":"doc","content":[{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"見出し"}]}]}`

// TestRichDocumentRepository_Integration は rich_documents の CRUD・jsonb 往復・楽観ロック・制約を
// 実 Postgres で固定する。
func TestRichDocumentRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	userRepo := persistence.NewUserRepository(sqlDB)
	repo := persistence.NewRichDocumentRepository(sqlDB)
	ctx := context.Background()

	// FK（owner_id → users.id）を満たすため作成者を用意する。
	mkOwner := func(t *testing.T, sub, email string) uint64 {
		t.Helper()
		u := &domain.User{Email: email}
		require.NoError(t, userRepo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, sub))
		return u.ID
	}

	t.Run("Create → FindByID で jsonb が意味的に往復する", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-1", "rd1@example.com")

		doc := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "メモ", Doc: rdDoc, Revision: 1}
		require.NoError(t, repo.Create(ctx, doc))
		require.NotEmpty(t, doc.ID) // UUID 採番済み

		// 採番は UUIDv7（時系列ソート可能・URL は推測困難のまま）。
		parsedID, err := uuid.Parse(doc.ID)
		require.NoError(t, err)
		require.Equal(t, uuid.Version(7), parsedID.Version())

		got, err := repo.FindByID(ctx, doc.ID)
		require.NoError(t, err)
		require.Equal(t, "メモ", got.Title)
		require.Equal(t, owner, got.OwnerID)
		require.Equal(t, 1, got.Revision)

		// jsonb は正規化されるので文字列一致ではなく構造で比較する。
		var parsed struct {
			Type    string `json:"type"`
			Content []struct {
				Type  string         `json:"type"`
				Attrs map[string]int `json:"attrs"`
			} `json:"content"`
		}
		require.NoError(t, json.Unmarshal([]byte(got.Doc), &parsed))
		require.Equal(t, "doc", parsed.Type)
		require.Equal(t, "heading", parsed.Content[0].Type)
		require.Equal(t, 2, parsed.Content[0].Attrs["level"])
	})

	t.Run("Create は int32 を超える revision / schema_version を切り詰めず保存する", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-bigint", "rdbigint@example.com")

		// revision / schema_version は bigint 列。パラメータを int4 に落とすと
		// この値は負数へ巻き戻り、エラーも出ないまま別の値が保存される。
		const beyondInt32 = math.MaxInt32 + 1
		doc := &domain.RichDocument{
			OwnerID: owner, Kind: domain.DocumentKindNote, Title: "大きい版", Doc: rdDoc,
			Revision: beyondInt32, SchemaVersion: beyondInt32,
		}
		require.NoError(t, repo.Create(ctx, doc))

		got, err := repo.FindByID(ctx, doc.ID)
		require.NoError(t, err)
		require.Equal(t, beyondInt32, got.Revision)
		require.Equal(t, beyondInt32, got.SchemaVersion)
	})

	t.Run("UpdateWithRevision は版一致時のみ更新し revision を +1 する", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-2", "rd2@example.com")
		doc := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "v1", Doc: rdDoc, Revision: 1}
		require.NoError(t, repo.Create(ctx, doc))

		upd := &domain.RichDocument{ID: doc.ID, Title: "v2", IsPublic: true, SchemaVersion: 1, Doc: `{"type":"doc","content":[]}`}
		require.NoError(t, repo.UpdateWithRevision(ctx, upd, 1))
		require.Equal(t, 2, upd.Revision) // 反映済み
		require.Equal(t, "v2", upd.Title)
		require.True(t, upd.IsPublic)

		got, err := repo.FindByID(ctx, doc.ID)
		require.NoError(t, err)
		require.Equal(t, 2, got.Revision)
		require.Equal(t, "v2", got.Title)
	})

	t.Run("UpdateWithRevision は版不一致で ErrRichDocumentConflict", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-3", "rd3@example.com")
		doc := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "v1", Doc: rdDoc, Revision: 1}
		require.NoError(t, repo.Create(ctx, doc))

		// いま revision=1 なのに 5 を期待して更新 → 競合。
		stale := &domain.RichDocument{ID: doc.ID, Title: "x", SchemaVersion: 1, Doc: rdDoc}
		err := repo.UpdateWithRevision(ctx, stale, 5)
		require.ErrorIs(t, err, repository.ErrRichDocumentConflict)
	})

	t.Run("UpdateWithRevision は存在しない ID で ErrRichDocumentNotFound", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "rich_documents")
		ghost := &domain.RichDocument{ID: "31400a07-297e-8057-884b-c05dbdf9fa53", Title: "x", SchemaVersion: 1, Doc: rdDoc}
		err := repo.UpdateWithRevision(ctx, ghost, 1)
		require.ErrorIs(t, err, repository.ErrRichDocumentNotFound)
	})

	t.Run("SoftDelete は所有者のみ・以後 FindByID から外れる", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-4", "rd4@example.com")
		doc := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "d", Doc: rdDoc, Revision: 1}
		require.NoError(t, repo.Create(ctx, doc))

		// 他人は消せない（0 行 → NotFound）。
		require.ErrorIs(t, repo.SoftDelete(ctx, doc.ID, owner+999), repository.ErrRichDocumentNotFound)

		// 所有者は消せる。
		require.NoError(t, repo.SoftDelete(ctx, doc.ID, owner))
		_, err := repo.FindByID(ctx, doc.ID)
		require.ErrorIs(t, err, repository.ErrRichDocumentNotFound)
	})

	t.Run("DB 制約: doc が type='doc' でなければ CHECK で拒否される", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-5", "rd5@example.com")
		_, err := sqlDB.Exec(
			`INSERT INTO rich_documents (id, owner_id, kind, title, doc, revision, created_at, updated_at)
			 VALUES (gen_random_uuid(), $1, 'note', 't', '{"type":"paragraph"}'::jsonb, 1, now(), now())`, owner,
		)
		require.ErrorContains(t, err, "ck_rich_documents_doc")
	})

	t.Run("DB 制約: 存在しない owner_id は FK で拒否される", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "rich_documents", "users", "user_oidc_identities")
		_, err := sqlDB.Exec(
			`INSERT INTO rich_documents (id, owner_id, kind, title, doc, revision, created_at, updated_at)
			 VALUES (gen_random_uuid(), 424242, 'note', 't', '{"type":"doc"}'::jsonb, 1, now(), now())`,
		)
		require.ErrorContains(t, err, "fk_rich_documents_owner")
	})

	t.Run("DB 制約: title は 200 文字まで許可し 201 文字で CHECK 違反", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-title", "rdtitle@example.com")

		// char_length ベースなので多バイト文字でも文字数で数える（200 は成功）。
		ok := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: strings.Repeat("あ", 200), Doc: rdDoc, Revision: 1}
		require.NoError(t, repo.Create(ctx, ok))

		ng := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: strings.Repeat("あ", 201), Doc: rdDoc, Revision: 1}
		err := repo.Create(ctx, ng)
		require.Error(t, err)
		require.ErrorContains(t, err, "ck_rich_documents_title_len")
	})

	t.Run("NUL(U+0000)を含む doc は class22 として ErrRichDocumentInvalidData に翻訳される", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-nul", "rdnul@example.com")

		// jsonb は U+0000 のエスケープを格納できず 22P05（class 22）を返す。repository が
		// これを ErrRichDocumentInvalidData に翻訳する（アプリの多層防御の最後の砦）。
		nulDoc := `{"type":"doc","content":[{"type":"text","text":"\u0000"}]}`

		// Create 経由
		d := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "t", Doc: nulDoc, Revision: 1}
		require.ErrorIs(t, repo.Create(ctx, d), repository.ErrRichDocumentInvalidData)

		// UpdateWithRevision 経由（有効な doc を作ってから NUL で更新する）
		base := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "t", Doc: rdDoc, Revision: 1}
		require.NoError(t, repo.Create(ctx, base))
		upd := &domain.RichDocument{ID: base.ID, Title: "t", SchemaVersion: 1, Doc: nulDoc}
		require.ErrorIs(t, repo.UpdateWithRevision(ctx, upd, 1), repository.ErrRichDocumentInvalidData)
	})

	t.Run("ListByOwner は owner で絞り・論理削除除外・更新日降順・kind フィルタ・doc本体なし", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "rich_documents", "users", "user_oidc_identities")
		owner := mkOwner(t, "rd-list", "rdlist@example.com")
		other := mkOwner(t, "rd-list-other", "rdlistother@example.com")

		// owner の文書: note 2 件 + course-chapter 1 件、他人の文書 1 件、論理削除 1 件。
		n1 := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "n1", Doc: rdDoc, Revision: 1}
		n2 := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "n2", Doc: rdDoc, Revision: 1}
		cc := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindCourseChapter, Title: "cc", Doc: rdDoc, Revision: 1}
		del := &domain.RichDocument{OwnerID: owner, Kind: domain.DocumentKindNote, Title: "del", Doc: rdDoc, Revision: 1}
		foreign := &domain.RichDocument{OwnerID: other, Kind: domain.DocumentKindNote, Title: "foreign", Doc: rdDoc, Revision: 1}
		for _, d := range []*domain.RichDocument{n1, n2, cc, del, foreign} {
			require.NoError(t, repo.Create(ctx, d))
		}
		require.NoError(t, repo.SoftDelete(ctx, del.ID, owner))

		// n1 を更新して updated_at を最新にし、降順で先頭に来ることを確認する。
		require.NoError(t, repo.UpdateWithRevision(ctx, &domain.RichDocument{ID: n1.ID, Title: "n1b", SchemaVersion: 1, Doc: rdDoc}, 1))
		// Create の updated_at は Go(ホスト)の時計、UpdateWithRevision は DB の now() で入る。
		// 別々の時計なので数 ms のずれで同着・逆転が起こりうる（CI をフレークさせた実績がある）。
		// ここで検証したいのは「更新日降順」なので、時刻を明示的に置いて時計差を排除する。
		// owner_id で絞るのは、将来このサブテストの前に別の前提データが置かれても
		// 無関係な行の updated_at まで書き換えないようにするため。
		_, err := sqlDB.Exec(
			`UPDATE rich_documents SET updated_at = TIMESTAMPTZ '2026-01-01 00:00:00+00' WHERE owner_id = $1`, owner,
		)
		require.NoError(t, err)
		_, err = sqlDB.Exec(
			`UPDATE rich_documents SET updated_at = TIMESTAMPTZ '2026-01-02 00:00:00+00' WHERE id = $1`, n1.ID,
		)
		require.NoError(t, err)

		// 全 kind（owner のみ・削除除外）
		all, err := repo.ListByOwner(ctx, owner, "")
		require.NoError(t, err)
		require.Len(t, all, 3) // n1 / n2 / cc（del と foreign は出ない）
		titles := []string{all[0].Title, all[1].Title, all[2].Title}
		require.Equal(t, "n1b", titles[0])    // 更新日降順で最新の n1b が先頭
		require.NotContains(t, titles, "del") // 論理削除は出ない
		require.NotContains(t, titles, "foreign")
		require.Empty(t, all[0].Doc) // 一覧は doc 本体を読み込まない（軽量）

		// kind=note フィルタ
		notes, err := repo.ListByOwner(ctx, owner, domain.DocumentKindNote)
		require.NoError(t, err)
		require.Len(t, notes, 2) // n1b / n2（cc は除外）
		for _, d := range notes {
			require.Equal(t, domain.DocumentKindNote, d.Kind)
		}
	})
}
